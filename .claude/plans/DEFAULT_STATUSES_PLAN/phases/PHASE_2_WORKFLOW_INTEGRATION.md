# Phase 2: Workflow Integration for Status Transitions

**Status**: Pending
**Category**: backend
**Dependencies**:
- Phase 1 (Form Configuration FK Default Resolution)
- **WORKFLOW_EVENT_FIRING_INFRASTRUCTURE** (see `.claude/plans/WORKFLOW_EVENT_FIRING_INFRASTRUCTURE.md`)

---

## PREREQUISITE: Workflow Event Firing Infrastructure

**Before starting this phase**, complete the Workflow Event Firing Infrastructure plan. That plan covers:
- Wiring `QueueManager` into application startup (`all.go`)
- Adding `QueueEvent` calls to formdata's `UpsertFormData`
- RabbitMQ client setup

Without that infrastructure, automation rules have nothing to trigger them. Once events are firing, return here to implement the status transition rules.

**Plan Location**: `.claude/plans/WORKFLOW_EVENT_FIRING_INFRASTRUCTURE.md`

---

## Overview

Wire automation rules to trigger allocation on order creation and update statuses based on results. This phase connects the workflow engine to order lifecycle events, enabling automatic status transitions from Pending → Allocated when inventory is successfully reserved.

**Key Insight**: The workflow system is fully implemented and tested (see `TestQueueManager_ProcessMessage` in `queue_test.go`). The flow is:
1. Entity is created → `QueueManager.QueueEvent()` is called with a `TriggerEvent`
2. Event is published to RabbitMQ
3. Consumer calls `Engine.ExecuteWorkflow()` which matches automation rules
4. Matched rule actions are executed (e.g., `allocate_inventory`, `send_email`)

**The Gap**: Currently, formdata's `UpsertFormData` does NOT call `QueueEvent` after creating entities. This means no workflow events are fired when orders are created via the form system. Phase 2 must wire formdata to the workflow system.

**Allocation Consumer Gap**: The allocation action queues messages to RabbitMQ via `QueueTypeInventory`, but there's no specialized consumer that calls `ProcessAllocation`. The generic consumer only calls `ExecuteWorkflow`.

## Goals

1. Implement a consumer for inventory allocation messages that calls `ProcessAllocation`
2. Fire workflow events after allocation completes (success or failure)
3. Seed automation rules for status transitions based on allocation results
4. Enable cascading workflows: allocation complete → update line item status

---

## Current Architecture Understanding

### How the Workflow System Works Today

Reference: `TestQueueManager_ProcessMessage` in `business/sdk/workflow/queue_test.go:323`

```
Entity Created (e.g., order via formdata)
    ↓
QueueManager.QueueEvent(ctx, TriggerEvent{
    EventType:  "on_create",
    EntityName: "orders",
    EntityID:   orderID,
    RawData:    {...},
})
    ↓
RabbitMQ (workflow.general queue)
    ↓
Consumer.processMessage() → Engine.ExecuteWorkflow()
    ↓
TriggerProcessor.ProcessEvent() → Matches automation rules by entity + trigger type
    ↓
ActionExecutor → Executes matched rule actions sequentially
    ↓
ActionHandler.Execute() (e.g., allocate_inventory, update_field, send_email)
```

**Currently Missing**: The `QueueEvent()` call from formdata after entity creation.

### Architectural Scope: Event Firing Across the Application

Workflow event firing is a **cross-cutting concern** that should be present throughout the application wherever entities are modified:

| Layer | Current State | Phase 2 Scope |
|-------|--------------|---------------|
| **formdata** (`UpsertFormData`) | No events fired | **Primary focus** - wire QueueEvent after create/update |
| **Domain bus layers** (e.g., `ordersbus.Create`) | Delegate calls exist for some (userbus) | Future consideration - add QueueEvent |
| **Direct API routes** (e.g., `POST /orders`) | No events fired | Future consideration |

For Phase 2, we focus on **formdata** since that's the primary entry point for order creation. Broader architectural integration can be a follow-up phase.

### The Allocation Disconnect

Currently, `allocate_inventory` action:
1. `Execute()` queues an `AllocationRequest` to RabbitMQ `QueueTypeInventory`
2. Returns immediately with tracking ID
3. **BUT**: No consumer calls `ProcessAllocation()` to process the queued work

The generic `processMessage` handler in `queue.go:272` converts messages to `TriggerEvent` and calls `ExecuteWorkflow`, which is designed for automation rules, not for processing allocation requests.

### Files and Locations

| File | Purpose | Key Functions |
|------|---------|---------------|
| `business/sdk/workflow/engine.go` | Workflow execution orchestration | `ExecuteWorkflow()` |
| `business/sdk/workflow/queue.go` | RabbitMQ consumer management | `QueueEvent()`, `processMessage()`, `startConsumer()` |
| `business/sdk/workflow/workflowactions/inventory/allocate.go` | Allocation logic | `Execute()`, `ProcessAllocation()` |
| `foundation/rabbitmq/client.go` | RabbitMQ client & consumers | `Consume()`, `Publish()` |

---

## Tasks

### Task 1: Create Inventory Allocation Consumer

**Problem**: The inventory queue has messages but no specialized handler.

**Files to Modify:**
- `business/sdk/workflow/queue.go`

**Implementation Steps:**
1. Modify `startConsumer()` to use specialized handlers for certain queue types
2. Create `processInventoryMessage()` that:
   - Deserializes `AllocationRequest` from the message payload
   - Gets the `AllocateInventoryHandler` from the action registry
   - Calls `ProcessAllocation()` with the request
   - Fires a workflow event with the result

**Code Example:**
```go
// In queue.go, modify startConsumer:
func (qm *QueueManager) startConsumer(ctx context.Context, queueType rabbitmq.QueueType) error {
    var handler rabbitmq.MessageHandler

    switch queueType {
    case rabbitmq.QueueTypeInventory:
        handler = func(ctx context.Context, msg *rabbitmq.Message) error {
            return qm.processInventoryMessage(ctx, msg)
        }
    default:
        handler = func(ctx context.Context, msg *rabbitmq.Message) error {
            return qm.processMessage(ctx, msg)
        }
    }

    consumer, err := qm.queue.Consume(ctx, queueType, handler)
    // ... rest unchanged
}

// New method:
func (qm *QueueManager) processInventoryMessage(ctx context.Context, msg *rabbitmq.Message) error {
    // Check message type
    if msg.Type != "inventory_allocation" {
        // Fall back to standard workflow processing
        return qm.processMessage(ctx, msg)
    }

    // Deserialize AllocationRequest from payload
    requestData, err := json.Marshal(msg.Payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    var request inventory.AllocationRequest
    if err := json.Unmarshal(requestData, &request); err != nil {
        return fmt.Errorf("failed to unmarshal allocation request: %w", err)
    }

    // Get handler from registry
    handler := qm.engine.actionRegistry.Get("allocate_inventory")
    allocHandler, ok := handler.(*inventory.AllocateInventoryHandler)
    if !ok {
        return fmt.Errorf("allocate_inventory handler not found or wrong type")
    }

    // Process the allocation
    result, err := allocHandler.ProcessAllocation(ctx, request)
    if err != nil {
        return fmt.Errorf("allocation failed: %w", err)
    }

    // Fire workflow event for the result (Task 2)
    return qm.fireAllocationResultEvent(ctx, result, request)
}
```

**Notes:**
- Requires access to action registry from QueueManager
- May need to inject `AllocateInventoryHandler` directly into QueueManager

---

### Task 2: Fire Workflow Event After Allocation Completes

**Files to Modify:**
- `business/sdk/workflow/queue.go` (add `fireAllocationResultEvent`)

**Implementation Steps:**
1. Create a new method that constructs a `TriggerEvent` from allocation result
2. Call `QueueEvent` with the new event to trigger downstream rules

**Code Example:**
```go
func (qm *QueueManager) fireAllocationResultEvent(ctx context.Context, result *inventory.InventoryAllocationResult, request inventory.AllocationRequest) error {
    event := TriggerEvent{
        EventType:  "on_create",
        EntityName: "allocation_results",
        EntityID:   result.AllocationID,
        Timestamp:  result.CompletedAt,
        UserID:     request.Context.UserID,
        RawData: map[string]interface{}{
            "status":           result.Status,  // "success", "partial", "failed"
            "reference_id":     request.Config.ReferenceID,
            "reference_type":   request.Config.ReferenceType,
            "total_allocated":  result.TotalAllocated,
            "total_requested":  result.TotalRequested,
            "allocated_items":  result.AllocatedItems,
            "failed_items":     result.FailedItems,
            "idempotency_key":  result.IdempotencyKey,
            "execution_time_ms": result.ExecutionTimeMs,
        },
    }

    return qm.QueueEvent(ctx, event)
}
```

**Notes:**
- This creates a new workflow trigger that can match automation rules
- Rules listening for `entity_name: "allocation_results"` with `trigger_type: "on_create"` will fire
- The RawData contains all the information needed for downstream actions

---

### Task 3: Seed Automation Rules for Status Transitions

**Files to Modify:**
- `business/sdk/dbtest/seedFrontend.go`

**Implementation Steps:**
1. Add entity type for `allocation_results` if not exists
2. Create automation rules for the order workflow

**Rule 1: Order Created → Allocate Inventory**

This rule triggers when an order is created via formdata and starts the allocation workflow.

```go
// In seedFrontend.go, add to automation rules section:
{
    RuleName:    "Sales Order Created - Allocate Inventory",
    EntityName:  "orders",  // Must match entity_name in TriggerEvent
    TriggerType: "on_create",
    Conditions:  nil,  // No conditions = always fire
    IsActive:    true,
    Actions: []seedmodels.SeedRuleAction{
        {
            ActionType: "allocate_inventory",
            Config: map[string]interface{}{
                "allocation_mode":     "reserve",
                "allocation_strategy": "fifo",
                "allow_partial":       false,
                "priority":            "medium",
                "reference_id":        "{{entity_id}}",
                "reference_type":      "order",
                // Note: inventory_items must be extracted from order line items
                // This may require a custom action or pre-processing
            },
            ExecutionOrder: 1,
            IsActive:       true,
        },
    },
},
```

**Rule 2: Allocation Success → Update Line Item Status**
```go
{
    RuleName:    "Allocation Success - Update Line Items",
    EntityName:  "allocation_results",
    TriggerType: "on_create",
    Conditions: []seedmodels.SeedRuleCondition{
        {
            FieldName: "status",
            Operator:  "equals",
            Value:     "success",
        },
    },
    IsActive: true,
    Actions: []seedmodels.SeedRuleAction{
        {
            ActionType: "update_field",
            Config: map[string]interface{}{
                "target_entity": "order_line_items",
                "target_field":  "line_item_fulfillment_statuses_id",
                "new_value":     "ALLOCATED",
                "field_type":    "foreign_key",
                "foreign_key_config": map[string]interface{}{
                    "reference_table": "sales.line_item_fulfillment_statuses",
                    "lookup_field":    "name",
                },
                "conditions": []map[string]interface{}{
                    {
                        "field_name": "order_id",
                        "operator":   "equals",
                        "value":      "{{reference_id}}",
                    },
                },
            },
            ExecutionOrder: 1,
            IsActive:       true,
        },
    },
},
```

**Rule 3: Allocation Failure → Create Alert**
```go
{
    RuleName:    "Allocation Failed - Alert Operations",
    EntityName:  "allocation_results",
    TriggerType: "on_create",
    Conditions: []seedmodels.SeedRuleCondition{
        {
            FieldName: "status",
            Operator:  "equals",
            Value:     "failed",
        },
    },
    IsActive: true,
    Actions: []seedmodels.SeedRuleAction{
        {
            ActionType: "create_alert",
            Config: map[string]interface{}{
                "alert_type": "inventory_allocation_failed",
                "severity":   "high",
                "message":    "Allocation failed for order {{reference_id}}: insufficient inventory",
                "context": map[string]interface{}{
                    "reference_id":   "{{reference_id}}",
                    "reference_type": "{{reference_type}}",
                    "failed_items":   "{{failed_items}}",
                },
            },
            ExecutionOrder: 1,
            IsActive:       true,
        },
    },
},
```

**Notes:**
- Template variables like `{{entity_id}}`, `{{reference_id}}` are resolved from `ActionExecutionContext.RawData`
- The `update_field` action already supports FK resolution via `ForeignKeyConfig`
- The `create_alert` action is currently a stub (Phase 3 work)

---

### Task 4: Verify update_field Batch Update Behavior

**Files to Review:**
- `business/sdk/workflow/workflowactions/data/updatefield.go`

**Implementation Steps:**
1. Review the conditions handling in `update_field` action
2. Verify that multiple records are updated when conditions match
3. If needed, add batch update support

**Current Behavior (from updatefield.go:254-325):**
The `update_field` action:
1. Builds a WHERE clause from conditions
2. Executes a single UPDATE statement
3. Returns `records_affected` count

**Verification:**
```sql
-- If conditions are: [{"field_name": "order_id", "operator": "equals", "value": "uuid-here"}]
-- Generated query should be:
UPDATE sales.order_line_items
SET line_item_fulfillment_statuses_id = 'resolved-uuid'
WHERE order_id = 'uuid-here'
-- This should update ALL line items for that order
```

**Test Case:**
1. Create order with 5 line items
2. Trigger allocation success workflow
3. Verify all 5 line items have status = ALLOCATED

---

### Task 5: Wire Domain Layers to Fire Workflow Events

**Problem**: No domain layer currently fires workflow events.

**Files to Modify:**
- `app/domain/formdata/formdataapp/formdataapp.go` (for order creation)
- Potentially `business/domain/sales/orderbus/orderbus.go`

**Implementation Steps:**
1. After successful formdata processing (order creation), fire workflow event
2. Inject `QueueManager` into formdataapp
3. Fire event with entity details

**Code Example:**
```go
// In formdataapp.go, after successful entity creation:
if qm := a.queueManager; qm != nil {
    event := workflow.TriggerEvent{
        EventType:  "on_create",
        EntityName: entityName,  // e.g., "orders"
        EntityID:   createdEntityID,
        Timestamp:  time.Now(),
        UserID:     userID,
        RawData:    createdEntityData,
    }

    if err := qm.QueueEvent(ctx, event); err != nil {
        a.log.Error(ctx, "failed to queue workflow event", "error", err)
        // Don't fail the request - workflow is async
    }
}
```

**Notes:**
- This is a significant architectural change
- Consider making event firing optional/configurable
- Events should be fire-and-forget (don't fail the request if event fails)

---

## Validation Criteria

- [ ] Go compilation passes (`go build ./...`)
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] Inventory queue messages are processed by specialized handler
- [ ] `ProcessAllocation()` is called for inventory messages
- [ ] Allocation completion fires workflow event to `allocation_results` entity
- [ ] Automation rules for `allocation_results` are seeded correctly
- [ ] `update_field` updates ALL matching line items (batch behavior)
- [ ] Template variables resolve correctly in rule conditions and actions
- [ ] Order creation via formdata triggers workflow events (if implemented)

---

## Testing Strategy

### Unit Tests
- Test `processInventoryMessage` correctly deserializes and processes requests
- Test `fireAllocationResultEvent` constructs correct TriggerEvent
- Test automation rule conditions match correctly

### Integration Tests

**Test File**: `api/cmd/services/ichor/tests/workflow/orderstatus/orderstatus_test.go`

```go
func Test_OrderStatusWorkflow(t *testing.T) {
    // 1. Seed inventory with available stock
    // 2. Create order via formdata with line items
    // 3. Wait for workflow processing (may need polling)
    // 4. Verify line items have status = ALLOCATED
}

func Test_AllocationFailureAlert(t *testing.T) {
    // 1. Create order for product with no inventory
    // 2. Wait for workflow processing
    // 3. Verify line items still have status = PENDING
    // 4. Verify alert was created (when Phase 3 complete)
}
```

### Manual Testing
```bash
# Start queue consumers
make dev-up

# Create order via formdata
curl -X POST /v1/formdata/upsert -d @order.json

# Check allocation result in logs
make dev-logs | grep allocation

# Query line items to verify status
curl /v1/sales/order-line-items?order_id=<uuid>
```

---

## Deliverables

- [ ] Specialized inventory message consumer in `queue.go`
- [ ] `fireAllocationResultEvent()` method
- [ ] Three automation rules seeded:
  - Order Created → Allocate Inventory
  - Allocation Success → Update Line Items
  - Allocation Failed → Create Alert
- [ ] Domain layer event firing (optional, may defer to Phase 2.5)
- [ ] Integration tests for workflow transitions

---

## Notes & Gotchas

### Dependency Injection
- `QueueManager` needs access to action handlers (registry or direct injection)
- `AllocateInventoryHandler` is currently constructed in `all.go`
- May need to pass handler reference to QueueManager

### Template Variable Context
- `{{entity_id}}` refers to the entity that triggered the rule
- For allocation_results triggers: `entity_id` = AllocationID
- `{{reference_id}}` comes from RawData (order ID in our case)

### Order of Operations
1. Order created via formdata (Phase 1 ensures default status = Pending)
2. Formdata fires `on_create` event for `orders` (this phase)
3. Rule matches, queues `allocate_inventory` action
4. Allocation consumer processes request
5. Allocation result event fires
6. Rule matches based on status, updates line items

### Partial Allocation Handling
Current config uses `allow_partial: false`. If we want partial allocation:
- Some line items would be ALLOCATED
- Others would remain PENDING
- Need to track which items succeeded/failed

### Entity Type for allocation_results
The `allocation_results` entity may need to be registered in the workflow entity types table for rules to match. Check if this is auto-registered or needs seeding.

### Alert Action is a Stub
The `create_alert` action in `workflowactions/communication/alert.go` currently just returns a mock response. Phase 3 will implement proper alert persistence and delivery.

---

## Architectural Considerations

### Alternative: Action Chaining
Instead of firing events after allocation, could use action chaining within a single rule:
```
Rule: Order Created
  Action 1: allocate_inventory
  Action 2: update_field (status = ALLOCATED) -- runs if Action 1 succeeds
  Action 3: create_alert -- runs if Action 1 fails
```

This requires implementing action result inspection and conditional execution within the ActionExecutor. Current implementation executes all actions sequentially regardless of previous results.

### Alternative: Synchronous Allocation
Instead of async queue processing, could make allocation synchronous:
- Simpler flow, no consumer needed
- But blocks order creation on allocation
- May cause timeouts for large orders

The current async approach is better for user experience but requires the consumer implementation in Task 1.

---

**Last Updated**: 2025-12-29
**Phase Author**: Claude Code

# Workflow Entry Points - Brainstorming & Architecture

**Date**: 2026-02-04
**Purpose**: Explore and define workflow entry points for the Ichor workflow system
**Related**: [action-sequence.md](.claude/plans/action-sequence.md) - Phase 12 implementation
**Focus**: Inventory management workflows as visualizable process diagrams

---

## Vision: Workflows as Visual Process Diagrams

The goal is to make workflows **visualizable** like Lucid diagrams - users should be able to see:
- "I created a PO, what is the process associated with this?"
- The full flow from entry point → conditions → actions → cascading effects
- How FormData multi-entity transactions trigger downstream processes

This is about making **business processes visible and configurable**, not just automation.

---

## Current State Analysis

### What We Have (Implemented)

The workflow system currently has **ACTIONS** (what happens when a workflow fires) but limited **ENTRY POINTS** (triggers that initiate workflows).

#### Currently Implemented Entry Points

| Entry Point | Mechanism | Status | Notes |
|-------------|-----------|--------|-------|
| **CRUD Delegates** | Business layer `delegate.Call()` | ✅ Fully Implemented | ~40+ domains registered |
| **FormData (Multi-Entity)** | App layer after tx.Commit() | ⚠️ Partial | Infrastructure exists, NOT wired |

#### Currently Implemented Trigger Types

| Trigger Type | Status | Implementation |
|--------------|--------|----------------|
| `on_create` | ✅ | DelegateHandler converts `created` action |
| `on_update` | ✅ | DelegateHandler converts `updated` action |
| `on_delete` | ✅ | DelegateHandler converts `deleted` action |
| `scheduled` | ❌ Documented but NOT implemented | No scheduler |

#### Currently Implemented Actions

| Action Type | Description | Status |
|-------------|-------------|--------|
| `create_alert` | Creates in-app alerts | ✅ |
| `update_field` | Updates database fields | ✅ |
| `send_email` | Sends email notifications | ✅ |
| `send_notification` | Multi-channel notifications | ✅ |
| `seek_approval` | Initiates approval workflows | ✅ |
| `allocate_inventory` | Reserves/allocates inventory | ✅ |
| `evaluate_condition` | Branching logic (Phase 12) | ✅ |

---

## Gap Analysis: Entry Points

### Problem Statement

We have a powerful action system (7 action types) and a sophisticated execution engine (linear + graph-based with branching), but workflows can ONLY be triggered by:
1. Direct CRUD operations via API → Business Layer → Delegate

This is limiting because real ERP workflows need to be triggered by:
- Scheduled jobs (daily reports, recurring tasks)
- Manual operator intervention (admin triggers)
- External webhooks (integrations)
- Custom application events
- Threshold/condition monitoring

### Missing Entry Points

| Entry Point | Use Cases | Priority |
|-------------|-----------|----------|
| **Scheduled/Timer** | Daily reports, recurring cleanup, batch processing, reminders | HIGH |
| **Manual/API Trigger** | Admin overrides, testing, operator-initiated processes | HIGH |
| **Webhook/External** | Third-party integrations, external system events | MEDIUM |
| **Custom Events** | Application-specific events not tied to CRUD | MEDIUM |
| **Threshold Monitoring** | Inventory low, budget exceeded, SLA breach | LOW (can use scheduled + conditions) |

---

## Current Event Flow (Reference)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CURRENT: CRUD ENTRY POINT                          │
│                                                                             │
│  API Request → App Layer → Business Layer                                   │
│                                 │                                           │
│                                 ▼                                           │
│                          delegate.Call()                                    │
│                                 │                                           │
│                                 ▼                                           │
│                        DelegateHandler                                      │
│                                 │                                           │
│                                 ▼                                           │
│                        EventPublisher                                       │
│                                 │                                           │
│                                 ▼                                           │
│                         RabbitMQ Queue                                      │
│                                 │                                           │
│                                 ▼                                           │
│                        QueueManager Consumer                                │
│                                 │                                           │
│                                 ▼                                           │
│                        WorkflowEngine.ExecuteWorkflow()                     │
│                                 │                                           │
│                                 ▼                                           │
│                     TriggerProcessor.ProcessEvent()                         │
│                                 │                                           │
│                                 ▼                                           │
│                     ActionExecutor.ExecuteRuleActions()                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Proposed Entry Points Architecture

### Entry Point 1: Scheduled/Timer Triggers

**Description**: Time-based workflow triggering using cron expressions

**Use Cases**:
- Daily inventory reconciliation
- Weekly report generation
- Monthly billing reminders
- SLA monitoring checks
- Cleanup/maintenance jobs

**Architecture Options**:

#### Option A: Internal Go Scheduler
- Use `robfig/cron` or similar Go library
- Scheduler runs within the Ichor service
- Polls `automation_rules` for `trigger_type=scheduled`
- Fires synthetic `TriggerEvent` with `EventType="scheduled"`

**Pros**: Simple, no external dependencies, transactional with app
**Cons**: Single instance only (needs leader election for HA)

#### Option B: External Job Scheduler
- Use Kubernetes CronJob, Temporal, or dedicated scheduler
- Calls internal API endpoint to trigger workflows
- Better for distributed/HA deployments

**Pros**: Production-grade scheduling, HA out of box
**Cons**: External dependency, more complex setup

#### Option C: Database-Driven Timer
- Store next_run_time in automation_rules
- Background goroutine polls for due rules
- Updates next_run_time after execution

**Pros**: Survives restarts, simple state management
**Cons**: Polling overhead, clock drift concerns

**Recommended**: Start with **Option A** (internal scheduler) for MVP, migrate to **Option B** for production HA.

**Schema Changes Needed**:
```sql
-- Add to automation_rules
ALTER TABLE workflow.automation_rules ADD COLUMN schedule_cron VARCHAR(100);
ALTER TABLE workflow.automation_rules ADD COLUMN next_scheduled_run TIMESTAMP;
ALTER TABLE workflow.automation_rules ADD COLUMN last_scheduled_run TIMESTAMP;
```

**New Components**:
- `business/sdk/workflow/scheduler.go` - Cron scheduler wrapper
- Schedule evaluation in TriggerProcessor

---

### Entry Point 2: Manual/API Trigger

**Description**: REST API endpoint to manually trigger a workflow rule

**Use Cases**:
- Testing workflows during development
- Admin override/intervention
- Operator-initiated processes
- Retry failed workflows
- "Run Now" button in UI

**API Design**:
```
POST /v1/workflow/rules/{ruleID}/trigger
{
  "entity_id": "uuid",           // Optional: specific entity to process
  "raw_data": {...},             // Optional: entity data override
  "user_id": "uuid",             // Optional: acting user
  "skip_conditions": false,      // Optional: bypass trigger conditions
  "dry_run": false               // Optional: simulate without executing
}

Response:
{
  "execution_id": "uuid",
  "status": "queued" | "executing" | "completed",
  "matched_rules": [...],
  "actions_executed": [...]
}
```

**Architecture**:
```
POST /trigger → ruleapi.trigger() → EventPublisher.PublishCustomEvent()
                                              │
                                              ▼
                                    TriggerEvent {
                                      EventType: "manual",
                                      EntityName: rule.entity_name,
                                      TriggeredBy: user_id,
                                      ...
                                    }
```

**Implementation**:
1. Add `manual` trigger type to `workflow.trigger_types`
2. Create API endpoint in `ruleapi/`
3. Use existing `EventPublisher.PublishCustomEvent()`
4. Add authorization checks (admin/operator only?)

**Considerations**:
- Should manual triggers bypass conditions? (configurable)
- Rate limiting to prevent abuse
- Audit logging for manual triggers

---

### Entry Point 3: Webhook/External Events

**Description**: HTTP endpoint that external systems can call to trigger workflows

**Use Cases**:
- Stripe payment webhook → trigger order fulfillment
- Shippo tracking update → trigger customer notification
- Calendar event → trigger reminder workflow
- External inventory system sync

**API Design**:
```
POST /v1/workflow/webhooks/{webhookID}
Headers:
  X-Webhook-Signature: {hmac_signature}  // For verification

Body: (external system's payload)
{
  "event_type": "payment.completed",
  "data": {...}
}

Response:
{
  "received": true,
  "execution_id": "uuid"   // If workflow triggered
}
```

**Architecture**:
```
┌──────────────────────────────────────────────────────────────────┐
│ New Table: workflow.webhooks                                      │
│ - id, name, secret_key, entity_name, field_mappings (JSONB)      │
│ - Maps external payload to TriggerEvent fields                   │
└──────────────────────────────────────────────────────────────────┘

External System → POST /webhooks/{id} → Verify Signature
                                              │
                                              ▼
                                       Map Payload to Entity
                                              │
                                              ▼
                                       EventPublisher.PublishCustomEvent()
```

**Schema**:
```sql
CREATE TABLE workflow.webhooks (
  id UUID PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  secret_key VARCHAR(256) NOT NULL,  -- For HMAC signature
  entity_name VARCHAR(100) NOT NULL,
  trigger_type_id UUID NOT NULL,
  field_mappings JSONB,              -- Maps external fields to entity fields
  is_active BOOLEAN DEFAULT TRUE,
  created_date TIMESTAMP DEFAULT NOW()
);
```

**Field Mappings Example**:
```json
{
  "external_field": "entity_field",
  "payment_id": "external_reference",
  "amount_cents": "total_amount",
  "customer_email": "customer.email"
}
```

---

### Entry Point 4: Custom Application Events

**Description**: Allow application code to fire custom events that aren't tied to CRUD

**Use Cases**:
- Login failed 5 times → trigger account lockout workflow
- Session expired → trigger re-auth reminder
- Batch import completed → trigger validation workflow
- Report generated → trigger distribution workflow

**API (Internal Go)**:
```go
// Already exists: EventPublisher.PublishCustomEvent()
// But needs better documentation and helper methods

// Proposed additions:
func (ep *EventPublisher) PublishApplicationEvent(
    ctx context.Context,
    eventType string,      // e.g., "login_failed", "report_ready"
    entityName string,     // e.g., "users", "reports"
    entityID uuid.UUID,
    data map[string]interface{},
    userID uuid.UUID,
)
```

**Trigger Type**:
- Add `custom` or `application` to `workflow.trigger_types`
- Or use entity-specific types: `on_login_failure`, `on_report_ready`

**Implementation**:
1. Define custom event types in trigger_types table
2. Add helper methods to EventPublisher
3. Document usage patterns for app developers

---

## Priority Ranking

Based on ERP use cases and implementation complexity:

| Priority | Entry Point | Rationale |
|----------|-------------|-----------|
| 1 | **Manual/API Trigger** | Easiest to implement, immediate value for testing/admin |
| 2 | **Scheduled/Timer** | Core ERP need (reports, reminders, batch jobs) |
| 3 | **Custom Application Events** | Already partially exists, needs documentation |
| 4 | **Webhook/External** | Important for integrations, but more complex |

---

## Questions for User

1. **Scheduled Triggers**:
   - Do you need HA/distributed scheduling, or is single-instance OK for now?
   - What's the minimum schedule granularity needed? (seconds, minutes, hours?)

2. **Manual Triggers**:
   - Should manual triggers bypass trigger conditions by default?
   - Who should be authorized to manually trigger workflows? (admin only, operators, any user?)

3. **Webhooks**:
   - Are there specific external systems you need to integrate with?
   - Is HMAC signature verification sufficient, or need OAuth/API key auth?

4. **Custom Events**:
   - Are there specific non-CRUD events you're already planning to fire?
   - Should custom events go through the same queue or have their own?

5. **FormData Integration**:
   - The infrastructure exists to fire events from FormData transactions but isn't wired. Should this be prioritized?

---

## Next Steps

After clarifying requirements:

1. **Phase 13**: Manual/API Trigger Implementation
   - Add `manual` trigger type
   - Create `/v1/workflow/rules/{id}/trigger` endpoint
   - Add dry-run support

2. **Phase 14**: Scheduled Trigger Implementation
   - Add cron fields to automation_rules schema
   - Implement Go scheduler component
   - Add scheduler management API

3. **Phase 15**: Webhook Entry Point
   - Create webhooks table and business layer
   - Implement webhook receiver endpoint
   - Add signature verification

4. **Phase 16**: Custom Events Documentation & Helpers
   - Document PublishCustomEvent usage
   - Add typed helper methods
   - Create examples for common patterns

---

## Files Reference

| Purpose | Location |
|---------|----------|
| Event Publisher | `business/sdk/workflow/eventpublisher.go` |
| Delegate Handler | `business/sdk/workflow/delegatehandler.go` |
| Trigger Processor | `business/sdk/workflow/trigger.go` |
| Workflow Engine | `business/sdk/workflow/engine.go` |
| Queue Manager | `business/sdk/workflow/queue.go` |
| Action Registry | `business/sdk/workflow/workflowactions/register.go` |
| Rule API | `api/domain/http/workflow/ruleapi/` |
| Models | `business/sdk/workflow/models.go` |
| Schema | `business/sdk/migrate/sql/migrate.sql` |
| Docs | `docs/workflow/` |

---

## Summary

**Current State**: Workflows can only be triggered by CRUD operations via the delegate pattern (~40 domains registered).

**Gap**: No scheduled triggers, no manual triggers, no webhooks, limited custom event support.

**Recommendation**:
1. Start with **Manual/API Trigger** (quick win, useful for testing)
2. Then **Scheduled Triggers** (core ERP need)
3. Then **Webhooks** (for integrations)
4. Finally, formalize **Custom Events** pattern

This document captures the exploration and brainstorming. Ready for discussion on priorities and implementation approach.

---

## Inventory Management Domain Analysis

### Core Inventory Tables (from migrate.sql)

| Table | Description | Key Workflow Events |
|-------|-------------|---------------------|
| `inventory.warehouses` | Physical warehouse locations | on_create (new warehouse setup) |
| `inventory.zones` | Zones within warehouses | on_create/update |
| `inventory.inventory_locations` | Specific storage locations (aisle/rack/shelf/bin) | on_create/update |
| `inventory.inventory_items` | Product quantities at locations | **on_update** (quantity changes are major) |
| `inventory.lot_trackings` | Lot/batch tracking | on_create (receiving), on_update (quality) |
| `inventory.serial_numbers` | Individual item tracking | on_create, on_update (status changes) |
| `inventory.quality_inspections` | Quality control records | on_create (inspection scheduled), on_update (result recorded) |
| `inventory.inventory_transactions` | Transaction log (receipts, picks, etc.) | **on_create** (every transaction) |
| `inventory.inventory_adjustments` | Manual adjustments | **on_create** (requires approval workflow) |
| `inventory.transfer_orders` | Inter-location transfers | **on_create, on_update** (status progression) |

### Procurement Tables (Supply Side)

| Table | Description | Key Workflow Events |
|-------|-------------|---------------------|
| `procurement.suppliers` | Supplier master data | on_create/update |
| `procurement.supplier_products` | What suppliers sell | on_create/update |
| `procurement.purchase_orders` | **PO header** | **on_create**, **on_update** (status changes) |
| `procurement.purchase_order_line_items` | **PO details** | on_create, on_update (receiving updates) |
| `procurement.purchase_order_statuses` | PO status lookup | - |
| `procurement.purchase_order_line_item_statuses` | Line item status lookup | - |

### Sales Tables (Demand Side)

| Table | Description | Key Workflow Events |
|-------|-------------|---------------------|
| `sales.orders` | Sales order header | **on_create**, **on_update** (status/fulfillment) |
| `sales.order_line_items` | Sales order details | on_create, on_update |
| `sales.customers` | Customer master data | on_create/update |

---

## Key Inventory Process Flows (Visualizable Workflows)

### Flow 1: Purchase Order Lifecycle

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     PURCHASE ORDER WORKFLOW                                  │
│                                                                             │
│  ENTRY POINTS:                                                              │
│  ├─ Manual: "Create PO" button in UI                                       │
│  ├─ FormData: Multi-entity PO + line items creation                        │
│  └─ API: POST /v1/procurement/purchase-orders                              │
│                                                                             │
│  ┌──────────────┐                                                          │
│  │ PO Created   │ on_create                                                │
│  │ (status:     │                                                          │
│  │  Draft)      │                                                          │
│  └──────┬───────┘                                                          │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────┐     ┌──────────────────────────────────────────────────┐ │
│  │ Condition:   │ YES │ Actions:                                          │ │
│  │ total_amount │────▶│ • seek_approval from purchasing_manager          │ │
│  │ > $10,000?   │     │ • create_alert "High value PO requires approval" │ │
│  └──────┬───────┘     └──────────────────────────────────────────────────┘ │
│         │ NO                                                                │
│         ▼                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ Actions:                                                              │  │
│  │ • send_email to supplier (PO notification)                           │  │
│  │ • update_field: status → "submitted"                                 │  │
│  │ • create_alert "PO submitted to supplier"                            │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│         │                                                                   │
│         ▼ (Cascading - status changed to "submitted")                       │
│  ┌──────────────┐                                                          │
│  │ PO Updated   │ on_update (status changed_to "submitted")                │
│  │ (status:     │                                                          │
│  │  Submitted)  │                                                          │
│  └──────┬───────┘                                                          │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ Actions:                                                              │  │
│  │ • update_field: expected_delivery_date calculation                   │  │
│  │ • create_alert "Expecting delivery on {{expected_delivery_date}}"    │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│         │                                                                   │
│         ▼ ... (continues through receiving, inspection, closed)            │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flow 2: Inventory Receiving (PO Receipt)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     INVENTORY RECEIVING WORKFLOW                             │
│                                                                             │
│  ENTRY POINTS:                                                              │
│  ├─ FormData: Create lot_tracking + inventory_transaction + update PO line │
│  ├─ API: POST /v1/inventory/transactions (type: "receipt")                 │
│  └─ UI: "Receive Items" action on PO line item                             │
│                                                                             │
│  ┌──────────────────┐                                                      │
│  │ inventory_       │ on_create (type: "receipt")                          │
│  │ transaction      │                                                      │
│  │ Created          │                                                      │
│  └──────┬───────────┘                                                      │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ Actions:                                                              │  │
│  │ • update_field: inventory_items.quantity += transaction.quantity     │  │
│  │ • create_alert "Received {{quantity}} of {{product_name}}"           │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│         │                                                                   │
│         ▼ (Cascading - inventory_items updated)                             │
│  ┌──────────────────┐                                                      │
│  │ inventory_item   │ on_update (quantity changed)                         │
│  │ Updated          │                                                      │
│  └──────┬───────────┘                                                      │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────┐     ┌──────────────────────────────────────────────────┐ │
│  │ Condition:   │ YES │ Actions:                                          │ │
│  │ quantity <   │────▶│ • create_alert "LOW STOCK: {{product_name}}"     │ │
│  │ reorder_point│     │ • send_email to inventory_manager                │ │
│  └──────┬───────┘     │ • (optional) trigger auto-reorder workflow       │ │
│         │ NO          └──────────────────────────────────────────────────┘ │
│         ▼                                                                   │
│  ┌──────────────┐     ┌──────────────────────────────────────────────────┐ │
│  │ Condition:   │ YES │ Actions:                                          │ │
│  │ quantity >   │────▶│ • create_alert "OVERSTOCK: {{product_name}}"     │ │
│  │ maximum_stock│     │ • send_email to inventory_planner                │ │
│  └──────────────┘     └──────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flow 3: Sales Order Fulfillment (Inventory Consumption)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     SALES ORDER FULFILLMENT WORKFLOW                         │
│                                                                             │
│  ENTRY POINTS:                                                              │
│  ├─ API/FormData: Create sales order                                       │
│  └─ UI: "Create Order" button                                              │
│                                                                             │
│  ┌──────────────┐                                                          │
│  │ Order        │ on_create                                                │
│  │ Created      │                                                          │
│  └──────┬───────┘                                                          │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ Actions:                                                              │  │
│  │ • allocate_inventory (reserve stock for order lines)                 │  │
│  │ • send_email to customer (order confirmation)                        │  │
│  │ • create_alert "New order {{order_number}} from {{customer_name}}"   │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│         │                                                                   │
│         ▼ (Cascading - inventory_items updated via allocate_inventory)      │
│  ┌──────────────────┐                                                      │
│  │ inventory_items  │ on_update (allocated_quantity changed)               │
│  │ Updated          │                                                      │
│  └──────┬───────────┘                                                      │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────┐     ┌──────────────────────────────────────────────────┐ │
│  │ Condition:   │ YES │ Actions:                                          │ │
│  │ available <  │────▶│ • create_alert "BACKORDER: {{product_name}}"     │ │
│  │ 0? (oversold)│     │ • update_field: order.status → "partial"         │ │
│  └──────────────┘     │ • send_email to sales team                       │ │
│                       └──────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flow 4: Inventory Adjustment (Cycle Count/Correction)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     INVENTORY ADJUSTMENT WORKFLOW                            │
│                                                                             │
│  ENTRY POINTS:                                                              │
│  ├─ UI: "Adjust Inventory" button                                          │
│  ├─ API: POST /v1/inventory/adjustments                                    │
│  └─ Scheduled: Daily cycle count task (future)                             │
│                                                                             │
│  ┌──────────────────┐                                                      │
│  │ inventory_       │ on_create                                            │
│  │ adjustment       │                                                      │
│  │ Created          │                                                      │
│  └──────┬───────────┘                                                      │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────┐     ┌──────────────────────────────────────────────────┐ │
│  │ Condition:   │ YES │ Actions:                                          │ │
│  │ abs(quantity │────▶│ • seek_approval from warehouse_manager           │ │
│  │ _change) >   │     │ • create_alert "Large adjustment requires approval│ │
│  │ threshold?   │     └──────────────────────────────────────────────────┘ │
│  └──────┬───────┘                                                          │
│         │ NO                                                                │
│         ▼                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ Actions:                                                              │  │
│  │ • update_field: inventory_items.quantity += adjustment.quantity_change│  │
│  │ • create inventory_transaction (type: "adjustment")                  │  │
│  │ • create_alert "Inventory adjusted: {{reason_code}}"                 │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flow 5: Transfer Order Workflow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     TRANSFER ORDER WORKFLOW                                  │
│                                                                             │
│  ENTRY POINTS:                                                              │
│  ├─ UI: "Create Transfer" button                                           │
│  ├─ API: POST /v1/inventory/transfer-orders                                │
│  └─ Auto: Low stock triggers replenishment from reserve location           │
│                                                                             │
│  ┌──────────────────┐                                                      │
│  │ transfer_order   │ on_create (status: "requested")                      │
│  │ Created          │                                                      │
│  └──────┬───────────┘                                                      │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ Actions:                                                              │  │
│  │ • seek_approval from warehouse_supervisor (if cross-zone)            │  │
│  │ • create_alert "Transfer requested: {{from_location}} → {{to_location}}"│
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│         │                                                                   │
│         ▼ (After approval - status changes)                                 │
│  ┌──────────────────┐                                                      │
│  │ transfer_order   │ on_update (status changed_to "in_transit")           │
│  │ Updated          │                                                      │
│  └──────┬───────────┘                                                      │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ Actions:                                                              │  │
│  │ • update_field: from_location.inventory_items.reserved_quantity      │  │
│  │ • create_alert "Transfer in progress"                                │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│         │                                                                   │
│         ▼ (Completion - status changes)                                     │
│  ┌──────────────────┐                                                      │
│  │ transfer_order   │ on_update (status changed_to "completed")            │
│  │ Updated          │                                                      │
│  └──────┬───────────┘                                                      │
│         │                                                                   │
│         ▼                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ Actions:                                                              │  │
│  │ • update_field: from_location.inventory_items.quantity -= qty        │  │
│  │ • update_field: to_location.inventory_items.quantity += qty          │  │
│  │ • create inventory_transactions (pickup and putaway)                 │  │
│  │ • create_alert "Transfer completed"                                  │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## FormData Integration for Visualizable Workflows

### The Key Insight

FormData operations are **multi-entity transactions** that should be visualizable as workflow entry points. When someone creates a PO with line items via FormData, this should:

1. Fire events for ALL entities created (PO header + each line item)
2. Be visualizable as a single "Create PO" entry point in the workflow diagram
3. Show cascading effects from each sub-entity

### FormData → Workflow Event Mapping

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     FORMDATA WORKFLOW VISUALIZATION                          │
│                                                                             │
│  FormData Request:                                                          │
│  {                                                                          │
│    "entities": [                                                            │
│      {"type": "purchase_orders", "operation": "CREATE", "data": {...}},    │
│      {"type": "purchase_order_line_items", "operation": "CREATE", ...},    │
│      {"type": "purchase_order_line_items", "operation": "CREATE", ...}     │
│    ]                                                                        │
│  }                                                                          │
│                                                                             │
│         │                                                                   │
│         ▼ (After tx.Commit())                                               │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ EventPublisher fires (in sequence or parallel):                        │ │
│  │                                                                         │ │
│  │   TriggerEvent{EntityName: "purchase_orders", EventType: "on_create"}  │ │
│  │         │                                                               │ │
│  │         ▼                                                               │ │
│  │   Rules for purchase_orders.on_create fire                             │ │
│  │                                                                         │ │
│  │   TriggerEvent{EntityName: "purchase_order_line_items", ...}           │ │
│  │         │                                                               │ │
│  │         ▼                                                               │ │
│  │   Rules for purchase_order_line_items.on_create fire                   │ │
│  │                                                                         │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  UI Visualization:                                                          │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                                                                        │ │
│  │  ┌─────────────┐         ┌─────────────────────────────────────────┐  │ │
│  │  │ FormData    │         │         CASCADING RULES                 │  │ │
│  │  │ "Create PO" │────────▶│  purchase_orders.on_create              │  │ │
│  │  │ Entry Point │         │    ├─► PO approval workflow             │  │ │
│  │  └─────────────┘         │    └─► PO notification workflow         │  │ │
│  │                          │                                          │  │ │
│  │                          │  purchase_order_line_items.on_create    │  │ │
│  │                          │    ├─► Inventory check workflow         │  │ │
│  │                          │    └─► Supplier notification workflow   │  │ │
│  │                          └─────────────────────────────────────────┘  │ │
│  │                                                                        │ │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### FormData Event Publishing (Implementation Needed)

The infrastructure exists but needs to be wired:

```go
// In formdataapp.go, after tx.Commit():
func (a *App) UpsertFormData(ctx context.Context, fd FormData) ([]Result, error) {
    // ... existing transaction logic ...

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    // NEW: Fire workflow events for all created/updated entities
    for _, result := range results {
        switch result.Operation {
        case "CREATE":
            a.eventPublisher.PublishCreateEvent(ctx, result.EntityName, result.Data, userID)
        case "UPDATE":
            a.eventPublisher.PublishUpdateEvent(ctx, result.EntityName, result.Data, nil, userID)
        }
    }

    return results, nil
}
```

---

## Logical Entry Points Summary (Inventory Focus)

### Primary Entry Points (High Priority)

| Entry Point | Entities | Workflow Trigger | Use Case |
|-------------|----------|------------------|----------|
| **Create PO** | purchase_orders + line_items | FormData + CRUD | PO approval, supplier notification |
| **Receive Inventory** | inventory_transactions, lot_tracking | FormData + CRUD | Update quantities, quality check |
| **Create Sales Order** | orders + line_items | FormData + CRUD | Allocation, fulfillment process |
| **Adjust Inventory** | inventory_adjustments | CRUD | Approval flow, audit trail |
| **Transfer Inventory** | transfer_orders | CRUD | Inter-location movement approval |

### Secondary Entry Points

| Entry Point | Entities | Workflow Trigger | Use Case |
|-------------|----------|------------------|----------|
| Quality Inspection | quality_inspections | CRUD | Inspection result actions |
| Low Stock Alert | inventory_items | on_update (condition) | Auto-reorder trigger |
| PO Status Change | purchase_orders | on_update | Status-driven notifications |
| Order Fulfillment | orders | on_update | Pick/pack/ship process |

---

## Implementation Phases (Revised)

### Phase 13: FormData Workflow Integration (Priority 1)

**Goal**: Make FormData transactions fire workflow events and be visualizable.

**Tasks**:
1. Wire EventPublisher to FormData app layer
2. Fire events after transaction commits (blocking or async?)
3. Add "source: formdata" metadata to TriggerEvent
4. Test with PO creation workflow

### Phase 14: Manual/API Trigger (Priority 2)

**Goal**: Allow manual triggering of workflows from UI or API.

**Tasks**:
1. Add `manual` trigger type
2. Create `/v1/workflow/rules/{id}/trigger` endpoint
3. Support dry-run mode
4. Add to workflow visualization

### Phase 15: Workflow Visualization Enhancements

**Goal**: Improve cascade visualization for process diagrams.

**Tasks**:
1. Enhance `/cascade-map` to show FormData entry points
2. Add "entry point" concept to automation_rules
3. Create workflow diagram export (JSON for frontend)
4. Link FormData operations to their triggered workflows

### Phase 16: Scheduled Triggers (Later)

Defer to later as requested - focus on manual + FormData first.

---

## Questions Resolved

Based on user feedback:
- **Priority**: Manual triggers + FormData integration
- **Scheduling**: Defer to later
- **FormData**: Priority - wire EventPublisher
- **Focus**: Inventory management workflows (PO, receiving, adjustments, transfers)
- **Visualization**: Workflows should be like Lucid diagrams - show full process flows

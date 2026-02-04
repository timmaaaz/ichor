# Ichor Workflow System - Phase 12 Implementation Plan

**Date**: 2026-02-02
**Purpose**: Phase 12 - Condition Nodes + Cascade Visualization

---

## Execution Phases Overview

| Phase | Name | Status | Jump To |
|-------|------|--------|---------|
| 12.1 | Schema Migration | ✅ | [Details](#phase-121-schema-migration) |
| 12.2 | Business Layer Models | ✅ | [Details](#phase-122-business-layer-models) |
| 12.3 | Condition Action Handler | ✅ | [Details](#phase-123-condition-action-handler) |
| 12.4 | Database Store for Edges | ✅ | [Details](#phase-124-database-store-for-edges) |
| 12.5 | Graph-Based Executor | ✅ | [Details](#phase-125-graph-based-executor) |
| 12.6 | EntityModifier Interface | ✅ | [Details](#phase-126-entitymodifier-interface) |
| 12.7 | API Layer - Edge Management | ✅ | [Details](#phase-127-api-layer---edge-management) |
| 12.8 | API Layer - Cascade Visualization | ✅ | [Details](#phase-128-api-layer---cascade-visualization) |

---

## Phase 12 Requirements

### Requirement 1: Condition Nodes (New Behavior)
Enable **branching within a workflow's action sequence** based on runtime conditions:
```
[Action 1] → [Condition: total > $1000?]
                    ├─ TRUE:  [Action 2a] → [Action 2b]
                    └─ FALSE: [Action 2c]
```

### Requirement 2: Cascade Visualization (New Visibility)
Show users **which workflows will be triggered** when an action modifies an entity:
```
[Update Order Status]
        │
        └─ ⚡ Will trigger:
             • "Send Shipping Email" workflow (listens to orders.on_update)
             • "Update Inventory" workflow (listens to orders.on_update)
```

The event-driven cascade behavior stays the same - we're just making it **visible** in the UI.

---

# Backend Architecture Analysis

---

## Executive Summary

The Ichor workflow system implements a **linear sequential action execution model** within rules, with branching only at the **rule level** via `rule_dependencies`. The system is event-driven: entity changes (create/update/delete) publish events to RabbitMQ, which are consumed by the workflow engine to execute matching automation rules.

**Key Finding**: Action-level branching (condition nodes) does **NOT** currently exist. The system evaluates conditions only at **workflow activation time** (trigger conditions), not during action execution. Implementing condition nodes requires schema changes, a new action type, and modifications to the execution engine.

**Cascading Workflows**: Actions CAN trigger other workflows via the `EventPublisher`, but this is event-driven (state machine model), not direct action chaining.

---

## 1. Workflow Execution Engine

### 1.1 Execution Entry Point

**File**: [engine.go:126](business/sdk/workflow/engine.go#L126)

```go
func (e *Engine) ExecuteWorkflow(ctx context.Context, event TriggerEvent) (*WorkflowExecution, error)
```

**Trigger Flow**:
1. Entity changes emit events via `DelegateHandler` → [delegatehandler.go:36](business/sdk/workflow/delegatehandler.go#L36)
2. `EventPublisher.PublishCreateEvent/UpdateEvent/DeleteEvent` queues to RabbitMQ → [eventpublisher.go:31](business/sdk/workflow/eventpublisher.go#L31)
3. `QueueManager.processMessage()` dequeues and calls `Engine.ExecuteWorkflow()` → [queue.go:535](business/sdk/workflow/queue.go#L535)

### 1.2 Execution Flow Diagram

```
TriggerEvent (on_create/update/delete)
    │
    ▼
Engine.ExecuteWorkflow()
    │
    ├─► createExecutionPlan()
    │       ├─► TriggerProcessor.ProcessEvent()     [Match rules by entity + conditions]
    │       ├─► DependencyResolver.CalculateBatchOrder()  [Topological sort of rules]
    │       └─► Create ExecutionBatches
    │
    └─► executeWorkflowInternal()  [Sequential batch processing]
            │
            └─► For each batch:
                    └─► executeBatch()
                            └─► executeRule() [or parallel if multiple rules]
                                    │
                                    └─► ActionExecutor.ExecuteRuleActions()
                                            │
                                            └─► For each action (ordered by execution_order ASC):
                                                    ├─► Validate config
                                                    ├─► Process template variables
                                                    ├─► Execute with retry logic
                                                    └─► Stop on critical failure (seek_approval only)
```

### 1.3 Action Execution - Sequential Loop (No Branching)

**File**: [executor.go:112-209](business/sdk/workflow/executor.go#L112-L209)

```go
// ExecuteRuleActions executes all actions for a given rule
func (ae *ActionExecutor) ExecuteRuleActions(ctx context.Context, ruleID uuid.UUID, executionContext ActionExecutionContext) (BatchExecutionResult, error) {
    // Load actions for the rule (ordered by execution_order ASC)
    actions, err := ae.workflowBus.QueryRoleActionsViewByRuleID(ctx, ruleID)

    // Execute actions in order - NO BRANCHING
    for _, action := range actions {
        if !action.IsActive {
            skippedCount++
            continue
        }

        result := ae.executeAction(ctx, action, executionContext)

        switch result.Status {
        case "success":
            successCount++
        case "failed":
            failedCount++
            // Only stop for seek_approval failures
            if ae.shouldStopOnFailure(action) {
                break
            }
        }
    }
    // ... build result
}
```

**Key Observation**: The loop is purely sequential. Every active action executes regardless of prior action outcomes (except for `seek_approval` failures).

### 1.4 Action Ordering - Database Query

**File**: [workflowdb.go:1135-1160](business/sdk/workflow/stores/workflowdb/workflowdb.go#L1135-L1160)

```sql
SELECT id, automation_rules_id, name, description, action_config,
       execution_order, is_active, template_id, template_name,
       template_action_type, template_default_config
FROM workflow.rule_actions_view
WHERE automation_rules_id = :automation_rules_id
ORDER BY execution_order ASC
```

**Conclusion**: `execution_order` is the ONLY ordering mechanism. No parent/child or graph traversal exists.

---

## 2. Trigger & Condition System

### 2.1 Condition Evaluation - BEFORE Workflow Only

**File**: [trigger.go:14-35](business/sdk/workflow/trigger.go#L14-L35)

```go
// FieldCondition represents a condition for field evaluation
type FieldCondition struct {
    FieldName     string      `json:"field_name"`
    Operator      string      `json:"operator"`
    Value         interface{} `json:"value,omitempty"`
    PreviousValue interface{} `json:"previous_value,omitempty"`
}

// TriggerConditions represents the conditions for triggering a rule
type TriggerConditions struct {
    FieldConditions []FieldCondition `json:"field_conditions,omitempty"`
}
```

**When Evaluated**: [trigger.go:116](business/sdk/workflow/trigger.go#L116)
- `TriggerProcessor.ProcessEvent()` evaluates conditions BEFORE any actions execute
- These are **activation filters**, not execution-time branching
- All conditions use AND logic - all must pass for rule to match

### 2.2 Supported Operators

**File**: [trigger.go:329-371](business/sdk/workflow/trigger.go#L329-L371)

| Operator | Description |
|----------|-------------|
| `equals` | Exact match |
| `not_equals` | Negation |
| `changed_from` | Previous value match (update events only) |
| `changed_to` | Value changed to specific value |
| `greater_than` | Numeric/string comparison |
| `less_than` | Numeric/string comparison |
| `contains` | Substring match |
| `in` | Array membership |

**Value Comparison Logic**: [trigger.go:377-430](business/sdk/workflow/trigger.go#L377-L430)
- Handles nil, numeric (float64 conversion), and string comparisons
- Can be REUSED for action-level condition evaluation

### 2.3 Action-Level Conditions - DO NOT EXIST

**Search Results**: No evidence of:
- Condition evaluation DURING action execution
- "condition" action type in `action_types` table
- Branching in the action execution loop

The only "branching" is stopping execution when `seek_approval` fails: [executor.go:159-164](business/sdk/workflow/executor.go#L159-L164)

---

## 3. Event System & Workflow Triggering

### 3.1 How Workflows Get Triggered

**DelegateHandler Registration**: [delegatehandler.go:36-55](business/sdk/workflow/delegatehandler.go#L36-L55)

```go
func (h *DelegateHandler) RegisterDomain(del *delegate.Delegate, domainName, entityName string) {
    // Register created action -> on_create event
    del.Register(domainName, ActionCreated, func(ctx context.Context, data delegate.Data) error {
        return h.handleCreated(ctx, entityName, data)
    })
    // ... similar for updated, deleted
}
```

**EventPublisher**: [eventpublisher.go:31-66](business/sdk/workflow/eventpublisher.go#L31-L66)

```go
func (ep *EventPublisher) PublishCreateEvent(ctx context.Context, entityName string, result any, userID uuid.UUID)
func (ep *EventPublisher) PublishUpdateEvent(ctx context.Context, entityName string, result any, fieldChanges map[string]FieldChange, userID uuid.UUID)
func (ep *EventPublisher) PublishDeleteEvent(ctx context.Context, entityName string, entityID uuid.UUID, userID uuid.UUID)
```

Events are queued non-blocking via goroutine to RabbitMQ.

### 3.2 Can One Workflow Trigger Another?

**YES** - via the event system (state machine model):

1. An action (e.g., `update_field`) modifies an entity
2. If that entity has workflow event publishing enabled, it emits an event
3. The event matches rules for that entity, triggering new workflows

**Custom Events**: [eventpublisher.go:68-76](business/sdk/workflow/eventpublisher.go#L68-L76)

```go
// PublishCustomEvent fires a custom event with full control over the TriggerEvent.
// Used by async action handlers to fire result events for downstream workflow rules.
func (ep *EventPublisher) PublishCustomEvent(ctx context.Context, event TriggerEvent)
```

This enables explicit workflow chaining from within action handlers.

### 3.3 Action Completion - No Automatic Event Emission

**File**: [executor.go:211-353](business/sdk/workflow/executor.go#L211-L353)

After action execution, there is NO automatic event emission. The only way actions trigger other workflows is:
1. By modifying an entity that has registered delegate handlers
2. By explicitly calling `EventPublisher.PublishCustomEvent()` from within the action handler

---

## 4. Data Model

### 4.1 Core Tables

**automation_rules** (v1.66): [migrate.sql](business/sdk/migrate/sql/migrate.sql)
```sql
CREATE TABLE workflow.automation_rules (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   entity_id UUID NOT NULL,                    -- Monitored table/view
   entity_type_id UUID NOT NULL,
   trigger_type_id UUID NOT NULL,              -- on_create, on_update, on_delete
   trigger_conditions JSONB NULL,              -- Activation conditions (evaluated BEFORE workflow)
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   updated_date TIMESTAMP NOT NULL DEFAULT NOW(),
   created_by UUID NOT NULL,
   updated_by UUID NOT NULL,
   deactivated_by UUID NULL
);
```

**rule_actions** (v1.69):
```sql
CREATE TABLE workflow.rule_actions (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   automation_rules_id UUID NOT NULL REFERENCES workflow.automation_rules(id),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   action_config JSONB NOT NULL,               -- Action-specific configuration
   execution_order INTEGER NOT NULL DEFAULT 1, -- Sequential execution order (ONLY ordering mechanism)
   is_active BOOLEAN DEFAULT TRUE,
   template_id UUID NULL,
   deactivated_by UUID NULL
);
```

**rule_dependencies** (v1.70) - Rule-to-Rule ONLY:
```sql
CREATE TABLE workflow.rule_dependencies (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   parent_rule_id UUID REFERENCES workflow.automation_rules(id),
   child_rule_id UUID REFERENCES workflow.automation_rules(id)
);
```

### 4.2 Relationship Storage Analysis

| Relationship Type | Supported | Mechanism |
|-------------------|-----------|-----------|
| Rule → Actions | YES | `automation_rules_id` foreign key |
| Action → Action (sequential) | YES | `execution_order` integer |
| Action → Action (branching) | **NO** | No schema support |
| Rule → Rule (dependency) | YES | `rule_dependencies` table |

### 4.3 Example action_config JSON

**allocate_inventory**:
```json
{
  "inventory_items": [{"product_id": "uuid", "quantity": 100}],
  "allocation_mode": "reserve",
  "allocation_strategy": "fifo",
  "allow_partial": true
}
```

**send_email**:
```json
{
  "recipients": ["user@example.com"],
  "subject": "Order {{entity_id}} Processed",
  "body": "Your order has been allocated"
}
```

### 4.4 Can Current Schema Represent a DAG?

**NO** - The current schema cannot represent:
- "Action A → if true → Action B, if false → Action C"
- Parent/child relationships between actions
- Edge-based graph traversal

The only way to branch is at the RULE level using `rule_dependencies`, not at the ACTION level.

---

## 5. Operator Support

### 5.1 Complete Operator List

**Trigger Processor** ([trigger.go:329-371](business/sdk/workflow/trigger.go#L329-L371)):
- `equals`, `not_equals`, `changed_from`, `changed_to`, `greater_than`, `less_than`, `contains`, `in`

**Update Field Action** ([updatefield.go:432-445](business/sdk/workflow/workflowactions/data/updatefield.go#L432-L445)):
- Additional: `is_null`, `is_not_null`, `not_in`

**Rule Simulation API** ([simulate.go:238-273](api/domain/http/workflow/ruleapi/simulate.go#L238-L273)):
- Extended: `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `starts_with`, `ends_with`, `is_empty`, `is_not_empty`

### 5.2 Operator Evaluation Function

**File**: [trigger.go:377-430](business/sdk/workflow/trigger.go#L377-L430)

```go
func (tp *TriggerProcessor) compareValues(a, b interface{}, op string) bool {
    // Handle nil cases
    if a == nil || b == nil {
        if op == "==" {
            return a == b
        }
        return false
    }

    switch op {
    case "==":
        return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
    case ">":
        aFloat, aOk := tp.toFloat64(a)
        bFloat, bOk := tp.toFloat64(b)
        if aOk && bOk {
            return aFloat > bFloat
        }
        return fmt.Sprintf("%v", a) > fmt.Sprintf("%v", b)
    case "<":
        // ... similar
    }
}
```

This logic can be extracted and reused for action-level condition evaluation.

---

## 6. Action Types & Extensibility

### 6.1 Current Action Types

**File**: [register.go](business/sdk/workflow/workflowactions/register.go)

| Action Type | File | Description |
|-------------|------|-------------|
| `update_field` | [updatefield.go](business/sdk/workflow/workflowactions/data/updatefield.go) | Update entity fields |
| `seek_approval` | [seek.go](business/sdk/workflow/workflowactions/approval/seek.go) | Create approval task |
| `send_email` | [email.go](business/sdk/workflow/workflowactions/communication/email.go) | Queue email |
| `send_notification` | [notification.go](business/sdk/workflow/workflowactions/communication/notification.go) | In-app notification |
| `create_alert` | [alert.go](business/sdk/workflow/workflowactions/communication/alert.go) | Create workflow alert |
| `allocate_inventory` | [allocate.go](business/sdk/workflow/workflowactions/inventory/allocate.go) | Reserve inventory |

**NO "condition" action type exists.**

### 6.2 Action Handler Interface

**File**: [interfaces.go](business/sdk/workflow/interfaces.go)

```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string
    SupportsManualExecution() bool
    IsAsync() bool
    GetDescription() string
}
```

### 6.3 Action Registration

**File**: [register.go:46-69](business/sdk/workflow/workflowactions/register.go#L46-L69)

```go
func RegisterAll(registry *workflow.ActionRegistry, config ActionConfig) {
    registry.Register(data.NewUpdateFieldHandler(config.Log, config.DB))
    registry.Register(approval.NewSeekApprovalHandler(config.Log, config.DB))
    registry.Register(communication.NewSendEmailHandler(config.Log, config.DB))
    // ... etc
}
```

New action types follow this pattern - implement `ActionHandler` and register.

---

## 7. Gap Analysis

### 7.1 What Already Exists

| Component | Status | Location |
|-----------|--------|----------|
| Trigger condition evaluation (before workflow) | ✅ | [trigger.go](business/sdk/workflow/trigger.go) |
| Sequential action execution | ✅ | [executor.go](business/sdk/workflow/executor.go) |
| Operator evaluation logic | ✅ | [trigger.go:329-430](business/sdk/workflow/trigger.go#L329-L430) |
| Action handler interface | ✅ | [interfaces.go](business/sdk/workflow/interfaces.go) |
| Action registry pattern | ✅ | [interfaces.go:62-92](business/sdk/workflow/interfaces.go#L62-L92) |
| Template variable processing | ✅ | [template.go](business/sdk/workflow/template.go) |
| Event publisher (cascading) | ✅ | [eventpublisher.go](business/sdk/workflow/eventpublisher.go) |
| Rule-level dependencies | ✅ | `rule_dependencies` table |

### 7.2 What Needs to Be Built

| Component | Gap | Priority |
|-----------|-----|----------|
| Condition action type (`evaluate_condition`) | Does not exist | HIGH |
| Action edge/graph table | No schema for branching | HIGH |
| Graph-based executor | Linear loop only | HIGH |
| Branch result tracking in `ActionResult` | No `branch_taken` field | MEDIUM |
| API for creating action edges | No endpoint | MEDIUM |
| Cycle detection for action graphs | Not needed currently | LOW |

### 7.3 Architectural Questions

**Q: Can the execution engine be modified without breaking existing workflows?**
A: YES - Use fallback strategy: if no action edges exist, use `execution_order` (linear execution). This maintains 100% backwards compatibility.

**Q: Is the schema extensible for graph relationships?**
A: YES - Add new `workflow.action_edges` table without modifying existing tables. Existing rules continue to work.

**Q: Are there blocking technical limitations?**
A: NO - The architecture is well-designed for extension. The action handler interface, registry pattern, and modular executor can be enhanced.

---

## 8. Implementation Plan for Phase 12

---

### PART A: Condition Nodes (Branching Within Workflows)

#### A.1 Schema Changes

**New Table**: `workflow.action_edges`
```sql
CREATE TABLE workflow.action_edges (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   rule_id UUID NOT NULL REFERENCES workflow.automation_rules(id) ON DELETE CASCADE,
   source_action_id UUID REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,  -- NULL = start edge
   target_action_id UUID NOT NULL REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,
   edge_type VARCHAR(20) NOT NULL CHECK (edge_type IN ('start', 'sequence', 'true_branch', 'false_branch', 'always')),
   edge_order INTEGER DEFAULT 0,
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   CONSTRAINT unique_edge UNIQUE(source_action_id, target_action_id, edge_type)
);

CREATE INDEX idx_action_edges_source ON workflow.action_edges(source_action_id);
CREATE INDEX idx_action_edges_target ON workflow.action_edges(target_action_id);
CREATE INDEX idx_action_edges_rule ON workflow.action_edges(rule_id);
```

#### A.2 New Action Type

**File**: `business/sdk/workflow/workflowactions/control/condition.go`

```go
type EvaluateConditionHandler struct {
    log *logger.Logger
}

func (h *EvaluateConditionHandler) GetType() string { return "evaluate_condition" }

func (h *EvaluateConditionHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
    // 1. Parse condition config
    // 2. Evaluate conditions against execCtx.RawData
    // 3. Return ConditionResult{Result: true/false, Branch: "true_branch"/"false_branch"}
}
```

#### A.3 Executor Modifications

**File**: [executor.go](business/sdk/workflow/executor.go)

Add `ExecuteRuleActionsGraph()` method that:
1. Loads actions and edges for rule
2. If no edges, falls back to linear `execution_order` execution
3. Finds start edges (source_action_id = NULL)
4. Traverses graph, following edges based on action results
5. For condition actions, filters outgoing edges by `BranchTaken`

#### A.4 API Endpoints for Edge Management

- `POST /v1/workflow/rules/{ruleID}/edges` - Create action edge
- `GET /v1/workflow/rules/{ruleID}/edges` - List edges for rule
- `DELETE /v1/workflow/rules/{ruleID}/edges/{edgeID}` - Remove edge

---

### PART B: Cascade Visualization (Show Downstream Workflows)

#### B.1 New API Endpoint

**Endpoint**: `GET /v1/workflow/actions/{actionID}/downstream-triggers`

**Purpose**: Given an action, return all workflows that would be triggered if this action modifies an entity.

**Logic**:
1. Parse the action's `action_config` to determine:
   - Which entity it modifies (e.g., `orders`)
   - Which fields it changes (e.g., `status`)
   - What event type this causes (`on_update`)
2. Query `automation_rules` for rules that:
   - Listen to that entity (`entity_id` matches)
   - Have matching trigger type (`on_update`)
   - Are active
3. Return list of downstream workflows with their trigger conditions

**Response Example**:
```json
{
  "action_id": "uuid",
  "action_type": "update_field",
  "modifies_entity": "orders",
  "triggers_event": "on_update",
  "downstream_workflows": [
    {
      "rule_id": "uuid",
      "rule_name": "Send Shipping Email",
      "trigger_conditions": {"field_conditions": [{"field_name": "status", "operator": "equals", "value": "shipped"}]},
      "will_trigger_if": "orders.status changes to 'shipped'"
    },
    {
      "rule_id": "uuid",
      "rule_name": "Update Inventory Levels",
      "trigger_conditions": null,
      "will_trigger_if": "any orders update"
    }
  ]
}
```

#### B.2 Alternative: Rule-Level Cascade View

**Endpoint**: `GET /v1/workflow/rules/{ruleID}/cascade-map`

**Purpose**: For an entire rule, show all downstream workflows that could be triggered by any of its actions.

**Response Example**:
```json
{
  "rule_id": "uuid",
  "rule_name": "Process Order",
  "actions": [
    {
      "action_id": "uuid",
      "action_name": "Update Order Status",
      "action_type": "update_field",
      "downstream_workflows": [
        {"rule_id": "uuid", "rule_name": "Send Shipping Email"},
        {"rule_id": "uuid", "rule_name": "Notify Warehouse"}
      ]
    },
    {
      "action_id": "uuid",
      "action_name": "Send Confirmation",
      "action_type": "send_email",
      "downstream_workflows": []  // Doesn't modify entities
    }
  ]
}
```

#### B.3 Action Type Metadata

To enable cascade detection, we need to know which action types modify entities. Add metadata to action handlers:

```go
type ActionHandler interface {
    // ... existing methods ...

    // NEW: Returns entity modification info for cascade visualization
    GetEntityModifications(config json.RawMessage) []EntityModification
}

type EntityModification struct {
    EntityName string   `json:"entity_name"`
    EventType  string   `json:"event_type"`  // on_create, on_update, on_delete
    Fields     []string `json:"fields,omitempty"`  // Which fields are modified
}
```

**Implementation per action type**:
- `update_field`: Returns `{EntityName: config.entity, EventType: "on_update", Fields: config.fields}`
- `send_email`: Returns `nil` (doesn't modify entities)
- `allocate_inventory`: Returns `{EntityName: "inventory_items", EventType: "on_update"}`

---

### PART C: Implementation Phases

| Phase | Description | Files |
|-------|-------------|-------|
| **12.1** | Add `action_edges` schema migration | `migrate.sql` |
| **12.2** | Create `evaluate_condition` action handler | `workflowactions/control/condition.go` |
| **12.3** | Add `BranchTaken` field to `ActionResult` | `models.go` |
| **12.4** | Implement graph-based executor | `executor.go` |
| **12.5** | Backwards compatibility fallback | `executor.go` |
| **12.6** | API endpoints for edge management | `actionedgeapi/` |
| **12.7** | Add `GetEntityModifications()` to action handlers | `workflowactions/*.go` |
| **12.8** | Cascade visualization API endpoint | `ruleapi/cascade.go` |

---

### PART D: Frontend Integration Points

The frontend workflow editor will need:

1. **For Condition Nodes**:
   - Render condition nodes as diamond shapes
   - Allow connecting true/false branches to different actions
   - Save/load edge data via new API endpoints

2. **For Cascade Visualization**:
   - Call `/cascade-map` when user views a workflow
   - Display "triggers" badges on entity-modifying actions
   - Show tooltip/panel with downstream workflow list
   - Optionally allow clicking through to the downstream workflow

---

## Appendix: Critical File References

| Purpose | File | Key Lines |
|---------|------|-----------|
| Execution entry point | [engine.go](business/sdk/workflow/engine.go) | 126-193 |
| Action execution loop | [executor.go](business/sdk/workflow/executor.go) | 112-209 |
| Trigger condition evaluation | [trigger.go](business/sdk/workflow/trigger.go) | 265-374 |
| Operator comparison | [trigger.go](business/sdk/workflow/trigger.go) | 377-430 |
| Action handler interface | [interfaces.go](business/sdk/workflow/interfaces.go) | 14-28 |
| Action registry | [interfaces.go](business/sdk/workflow/interfaces.go) | 62-92 |
| Action registration | [register.go](business/sdk/workflow/workflowactions/register.go) | 46-69 |
| Event publisher | [eventpublisher.go](business/sdk/workflow/eventpublisher.go) | 31-76 |
| Delegate handler | [delegatehandler.go](business/sdk/workflow/delegatehandler.go) | 36-55 |
| Database queries | [workflowdb.go](business/sdk/workflow/stores/workflowdb/workflowdb.go) | 1135-1160 |
| Schema migrations | [migrate.sql](business/sdk/migrate/sql/migrate.sql) | v1.64-1.98 |

---

## Summary Answers

1. **Is action-level branching already implemented?** NO - Only trigger conditions (before workflow) and rule-level dependencies exist.

2. **Can actions trigger other workflows?** YES - Via event publishing (state machine model), not direct chaining.

3. **What's the execution model?** Sequential loop by `execution_order`, with batched parallel rule execution via `rule_dependencies`.

4. **What schema changes are needed?** New `workflow.action_edges` table for graph relationships.

5. **What backend code needs to be written?**
   - New `evaluate_condition` action handler
   - Graph-based executor method
   - `BranchTaken` field in `ActionResult`
   - API endpoints for edge management
   - `GetEntityModifications()` on action handlers for cascade visualization
   - Cascade visualization API endpoint

---

# Implementation Details

## Phase 12.1: Schema Migration

**File**: `business/sdk/migrate/sql/migrate.sql`

Find the last version number and add:

```sql
-- Version: X.XX
-- Description: Add action edges for workflow branching
CREATE TABLE workflow.action_edges (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   rule_id UUID NOT NULL REFERENCES workflow.automation_rules(id) ON DELETE CASCADE,
   source_action_id UUID REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,
   target_action_id UUID NOT NULL REFERENCES workflow.rule_actions(id) ON DELETE CASCADE,
   edge_type VARCHAR(20) NOT NULL CHECK (edge_type IN ('start', 'sequence', 'true_branch', 'false_branch', 'always')),
   edge_order INTEGER DEFAULT 0,
   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   CONSTRAINT unique_edge UNIQUE(source_action_id, target_action_id, edge_type)
);

CREATE INDEX idx_action_edges_source ON workflow.action_edges(source_action_id);
CREATE INDEX idx_action_edges_target ON workflow.action_edges(target_action_id);
CREATE INDEX idx_action_edges_rule ON workflow.action_edges(rule_id);
```

---

## Phase 12.2: Business Layer Models

**File**: `business/sdk/workflow/models.go`

1. Add `BranchTaken` field to `ActionResult` (around line 122):
```go
type ActionResult struct {
    ActionID     uuid.UUID              `json:"action_id"`
    ActionName   string                 `json:"action_name"`
    ActionType   string                 `json:"action_type"`
    Status       string                 `json:"status"`
    ResultData   map[string]interface{} `json:"result_data,omitempty"`
    ErrorMessage string                 `json:"error_message,omitempty"`
    Duration     time.Duration          `json:"duration_ms"`
    StartedAt    time.Time              `json:"started_at"`
    CompletedAt  *time.Time             `json:"completed_at,omitempty"`
    BranchTaken  string                 `json:"branch_taken,omitempty"` // NEW: "true_branch", "false_branch", or empty
}
```

2. Add new edge models after `RuleDependency` (around line 380):
```go
// ActionEdge represents a directed edge between actions in a workflow graph
type ActionEdge struct {
    ID             uuid.UUID
    RuleID         uuid.UUID
    SourceActionID *uuid.UUID // nil for start edges
    TargetActionID uuid.UUID
    EdgeType       string // start, sequence, true_branch, false_branch, always
    EdgeOrder      int
    CreatedDate    time.Time
}

// NewActionEdge contains information needed to create a new action edge
type NewActionEdge struct {
    RuleID         uuid.UUID
    SourceActionID *uuid.UUID
    TargetActionID uuid.UUID
    EdgeType       string
    EdgeOrder      int
}

// ConditionResult represents the result of evaluating a condition action
type ConditionResult struct {
    Evaluated   bool   `json:"evaluated"`
    Result      bool   `json:"result"`
    BranchTaken string `json:"branch_taken"` // "true_branch" or "false_branch"
}
```

---

## Phase 12.3: Condition Action Handler

**New File**: `business/sdk/workflow/workflowactions/control/condition.go`

```go
package control

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/foundation/logger"
)

// FieldCondition matches the trigger.go structure for reuse
type FieldCondition struct {
    FieldName     string      `json:"field_name"`
    Operator      string      `json:"operator"`
    Value         interface{} `json:"value,omitempty"`
    PreviousValue interface{} `json:"previous_value,omitempty"`
}

// ConditionConfig defines the configuration for evaluate_condition action
type ConditionConfig struct {
    Conditions []FieldCondition `json:"conditions"`
    LogicType  string           `json:"logic_type"` // "and" (default) or "or"
}

// EvaluateConditionHandler evaluates conditions and returns branch direction
type EvaluateConditionHandler struct {
    log *logger.Logger
}

func NewEvaluateConditionHandler(log *logger.Logger) *EvaluateConditionHandler {
    return &EvaluateConditionHandler{log: log}
}

func (h *EvaluateConditionHandler) GetType() string {
    return "evaluate_condition"
}

func (h *EvaluateConditionHandler) GetDescription() string {
    return "Evaluates conditions against entity data and determines branch direction"
}

func (h *EvaluateConditionHandler) SupportsManualExecution() bool {
    return false // Conditions only make sense in workflow context
}

func (h *EvaluateConditionHandler) IsAsync() bool {
    return false
}

func (h *EvaluateConditionHandler) Validate(config json.RawMessage) error {
    var cfg ConditionConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("invalid condition config: %w", err)
    }
    if len(cfg.Conditions) == 0 {
        return fmt.Errorf("at least one condition is required")
    }
    return nil
}

func (h *EvaluateConditionHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
    var cfg ConditionConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse condition config: %w", err)
    }

    // Default to AND logic
    logicType := cfg.LogicType
    if logicType == "" {
        logicType = "and"
    }

    // Evaluate conditions against RawData
    // Reuse comparison logic from trigger.go
    result := h.evaluateConditions(cfg.Conditions, execCtx.RawData, execCtx.FieldChanges, logicType)

    branchTaken := "false_branch"
    if result {
        branchTaken = "true_branch"
    }

    return workflow.ConditionResult{
        Evaluated:   true,
        Result:      result,
        BranchTaken: branchTaken,
    }, nil
}

func (h *EvaluateConditionHandler) evaluateConditions(conditions []FieldCondition, data map[string]interface{}, changes map[string]workflow.FieldChange, logicType string) bool {
    // Implementation mirrors TriggerProcessor.evaluateFieldConditions in trigger.go
    // ... comparison logic for equals, not_equals, greater_than, less_than, contains, etc.
}
```

**Update**: `business/sdk/workflow/workflowactions/register.go`

Add to imports:
```go
"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
```

Add to `RegisterAll()`:
```go
registry.Register(control.NewEvaluateConditionHandler(config.Log))
```

---

## Phase 12.4: Database Store for Edges

**New File**: `business/sdk/workflow/stores/workflowdb/actionedge.go`

```go
package workflowdb

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
)

// CreateActionEdge creates a new action edge
func (s *Store) CreateActionEdge(ctx context.Context, edge workflow.NewActionEdge) error {
    const q = `
    INSERT INTO workflow.action_edges (rule_id, source_action_id, target_action_id, edge_type, edge_order)
    VALUES (:rule_id, :source_action_id, :target_action_id, :edge_type, :edge_order)`

    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBActionEdge(edge)); err != nil {
        return fmt.Errorf("namedexeccontext: %w", err)
    }
    return nil
}

// QueryEdgesByRuleID returns all edges for a rule
func (s *Store) QueryEdgesByRuleID(ctx context.Context, ruleID uuid.UUID) ([]workflow.ActionEdge, error) {
    const q = `
    SELECT id, rule_id, source_action_id, target_action_id, edge_type, edge_order, created_date
    FROM workflow.action_edges
    WHERE rule_id = :rule_id
    ORDER BY edge_order ASC`

    data := struct{ RuleID uuid.UUID `db:"rule_id"` }{RuleID: ruleID}

    var dbEdges []dbActionEdge
    if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbEdges); err != nil {
        return nil, fmt.Errorf("namedqueryslice: %w", err)
    }

    return toBusActionEdges(dbEdges), nil
}

// DeleteActionEdge deletes an edge by ID
func (s *Store) DeleteActionEdge(ctx context.Context, edgeID uuid.UUID) error {
    const q = `DELETE FROM workflow.action_edges WHERE id = :id`
    data := struct{ ID uuid.UUID `db:"id"` }{ID: edgeID}

    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
        return fmt.Errorf("namedexeccontext: %w", err)
    }
    return nil
}
```

---

## Phase 12.5: Graph-Based Executor

**File**: `business/sdk/workflow/executor.go`

Add after `ExecuteRuleActions()` (around line 209):

```go
// ExecuteRuleActionsGraph executes actions following the edge graph.
// Falls back to linear execution_order if no edges exist (backwards compatible).
func (ae *ActionExecutor) ExecuteRuleActionsGraph(ctx context.Context, ruleID uuid.UUID, executionContext ActionExecutionContext) (BatchExecutionResult, error) {
    startTime := time.Now()

    // Load edges for this rule
    edges, err := ae.workflowBus.QueryEdgesByRuleID(ctx, ruleID)
    if err != nil {
        return BatchExecutionResult{}, fmt.Errorf("failed to load edges: %w", err)
    }

    // Backwards compatibility: if no edges, use linear execution
    if len(edges) == 0 {
        return ae.ExecuteRuleActions(ctx, ruleID, executionContext)
    }

    // Load all actions
    actions, err := ae.workflowBus.QueryRoleActionsViewByRuleID(ctx, ruleID)
    if err != nil {
        return BatchExecutionResult{}, fmt.Errorf("failed to load actions: %w", err)
    }

    // Build action map for quick lookup
    actionMap := make(map[uuid.UUID]RuleActionView)
    for _, action := range actions {
        actionMap[action.ID] = action
    }

    // Build adjacency list from edges
    outgoingEdges := make(map[uuid.UUID][]ActionEdge) // source -> edges
    var startEdges []ActionEdge

    for _, edge := range edges {
        if edge.SourceActionID == nil {
            startEdges = append(startEdges, edge)
        } else {
            outgoingEdges[*edge.SourceActionID] = append(outgoingEdges[*edge.SourceActionID], edge)
        }
    }

    // Execute using BFS from start edges
    actionResults := make([]ActionResult, 0)
    executed := make(map[uuid.UUID]bool)
    queue := make([]uuid.UUID, 0)

    // Add start edge targets to queue
    for _, edge := range startEdges {
        queue = append(queue, edge.TargetActionID)
    }

    for len(queue) > 0 {
        actionID := queue[0]
        queue = queue[1:]

        if executed[actionID] {
            continue
        }
        executed[actionID] = true

        action, exists := actionMap[actionID]
        if !exists || !action.IsActive {
            continue
        }

        // Execute the action
        result := ae.executeAction(ctx, action, executionContext)
        actionResults = append(actionResults, result)

        // Determine which edges to follow
        nextEdges := outgoingEdges[actionID]
        for _, edge := range nextEdges {
            shouldFollow := false

            switch edge.EdgeType {
            case "always", "sequence":
                shouldFollow = true
            case "true_branch":
                if result.BranchTaken == "true_branch" {
                    shouldFollow = true
                }
            case "false_branch":
                if result.BranchTaken == "false_branch" {
                    shouldFollow = true
                }
            }

            if shouldFollow {
                queue = append(queue, edge.TargetActionID)
            }
        }
    }

    // Build result (similar to ExecuteRuleActions)
    // ...
}
```

Update `executeAction()` to set `BranchTaken` from condition results:
```go
// In executeAction(), after getting resultData:
if condResult, ok := resultData.(ConditionResult); ok {
    result.BranchTaken = condResult.BranchTaken
}
```

---

## Phase 12.6: EntityModifier Interface

**File**: `business/sdk/workflow/interfaces.go`

Add after `AsyncActionHandler` (around line 113):

```go
// EntityModifier is an optional interface for action handlers that modify entities.
// Used for cascade visualization to determine which downstream workflows may trigger.
type EntityModifier interface {
    // GetEntityModifications returns what entities/fields this action modifies.
    // Returns nil if the action doesn't modify entities (e.g., send_email).
    GetEntityModifications(config json.RawMessage) []EntityModification
}

// EntityModification describes an entity modification caused by an action
type EntityModification struct {
    EntityName string   `json:"entity_name"`
    EventType  string   `json:"event_type"` // on_create, on_update, on_delete
    Fields     []string `json:"fields,omitempty"`
}
```

**Implement on update_field handler** (`workflowactions/data/updatefield.go`):
```go
func (h *UpdateFieldHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
    var cfg UpdateFieldConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil
    }

    fields := make([]string, 0, len(cfg.Updates))
    for _, update := range cfg.Updates {
        fields = append(fields, update.FieldName)
    }

    return []workflow.EntityModification{{
        EntityName: cfg.Entity,
        EventType:  "on_update",
        Fields:     fields,
    }}
}
```

---

## Phase 12.7: API Layer - Edge Management

**New Directory**: `api/domain/http/workflow/edgeapi/`

**File**: `api/domain/http/workflow/edgeapi/route.go`
```go
package edgeapi

import (
    "net/http"
    // ... imports
)

type Config struct {
    Log         *logger.Logger
    WorkflowBus *workflow.Business
    AuthClient  *authclient.Client
}

func Routes(app *web.App, cfg Config) {
    const version = "v1"
    api := newAPI(cfg.WorkflowBus)
    authen := mid.Authenticate(cfg.AuthClient)

    app.HandlerFunc(http.MethodPost, version, "/workflow/rules/{ruleID}/edges", api.create, authen)
    app.HandlerFunc(http.MethodGet, version, "/workflow/rules/{ruleID}/edges", api.query, authen)
    app.HandlerFunc(http.MethodDelete, version, "/workflow/rules/{ruleID}/edges/{edgeID}", api.delete, authen)
}
```

**File**: `api/domain/http/workflow/edgeapi/edgeapi.go`
```go
package edgeapi

type api struct {
    workflowBus *workflow.Business
}

func newAPI(workflowBus *workflow.Business) *api {
    return &api{workflowBus: workflowBus}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
    // Parse ruleID from path
    // Decode NewEdge from body
    // Call workflowBus.CreateActionEdge
    // Return created edge
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
    // Parse ruleID from path
    // Call workflowBus.QueryEdgesByRuleID
    // Return edges
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
    // Parse edgeID from path
    // Call workflowBus.DeleteActionEdge
    // Return success
}
```

---

## Phase 12.8: API Layer - Cascade Visualization

**New File**: `api/domain/http/workflow/ruleapi/cascade.go`

```go
package ruleapi

// CascadeResponse represents downstream workflows for a rule
type CascadeResponse struct {
    RuleID   string                `json:"rule_id"`
    RuleName string                `json:"rule_name"`
    Actions  []ActionCascadeInfo   `json:"actions"`
}

type ActionCascadeInfo struct {
    ActionID            string                    `json:"action_id"`
    ActionName          string                    `json:"action_name"`
    ActionType          string                    `json:"action_type"`
    DownstreamWorkflows []DownstreamWorkflowInfo `json:"downstream_workflows"`
}

type DownstreamWorkflowInfo struct {
    RuleID   string `json:"rule_id"`
    RuleName string `json:"rule_name"`
}

func (api *api) cascadeMap(ctx context.Context, r *http.Request) web.Encoder {
    ruleID := web.Param(r, "id")
    // ... parse UUID

    // Load rule and its actions
    rule, _ := api.workflowBus.QueryAutomationRuleByID(ctx, ruleID)
    actions, _ := api.workflowBus.QueryRoleActionsViewByRuleID(ctx, ruleID)

    response := CascadeResponse{
        RuleID:   rule.ID.String(),
        RuleName: rule.Name,
        Actions:  make([]ActionCascadeInfo, 0, len(actions)),
    }

    for _, action := range actions {
        info := ActionCascadeInfo{
            ActionID:   action.ID.String(),
            ActionName: action.Name,
            ActionType: action.TemplateActionType,
        }

        // Check if handler implements EntityModifier
        handler, exists := api.registry.Get(action.TemplateActionType)
        if exists {
            if modifier, ok := handler.(workflow.EntityModifier); ok {
                mods := modifier.GetEntityModifications(action.ActionConfig)
                for _, mod := range mods {
                    // Query rules listening to this entity
                    rules, _ := api.workflowBus.QueryRulesByEntity(ctx, mod.EntityName, mod.EventType)
                    for _, r := range rules {
                        info.DownstreamWorkflows = append(info.DownstreamWorkflows, DownstreamWorkflowInfo{
                            RuleID:   r.ID.String(),
                            RuleName: r.Name,
                        })
                    }
                }
            }
        }

        response.Actions = append(response.Actions, info)
    }

    return response
}
```

Add to `route.go`:
```go
app.HandlerFunc(http.MethodGet, version, "/workflow/rules/{id}/cascade-map", api.cascadeMap, authen)
```

---

# Test Plan

## Unit Tests (Business Layer)

**New File**: `business/sdk/workflow/condition_test.go`
```go
func TestEvaluateConditionHandler_Execute(t *testing.T) {
    testCases := []struct {
        name       string
        conditions []FieldCondition
        data       map[string]interface{}
        wantBranch string
    }{
        {
            name: "equals_true",
            conditions: []FieldCondition{{FieldName: "status", Operator: "equals", Value: "active"}},
            data: map[string]interface{}{"status": "active"},
            wantBranch: "true_branch",
        },
        {
            name: "equals_false",
            conditions: []FieldCondition{{FieldName: "status", Operator: "equals", Value: "active"}},
            data: map[string]interface{}{"status": "inactive"},
            wantBranch: "false_branch",
        },
        {
            name: "greater_than_true",
            conditions: []FieldCondition{{FieldName: "amount", Operator: "greater_than", Value: 1000}},
            data: map[string]interface{}{"amount": 1500},
            wantBranch: "true_branch",
        },
        // ... test all operators: not_equals, less_than, contains, in, changed_from, changed_to
        // ... test AND/OR logic
        // ... test nil values, type mismatches
    }
    // ... run tests
}
```

**New File**: `business/sdk/workflow/graph_executor_test.go`
```go
func TestExecuteRuleActionsGraph(t *testing.T) {
    testCases := []struct {
        name           string
        actions        []RuleActionView
        edges          []ActionEdge
        conditionResult bool
        wantExecuted   []string // action names in order
    }{
        {
            name: "no_edges_falls_back_to_linear",
            actions: []RuleActionView{{Name: "A"}, {Name: "B"}},
            edges: nil,
            wantExecuted: []string{"A", "B"},
        },
        {
            name: "simple_branch_true",
            actions: []RuleActionView{
                {Name: "Condition", TemplateActionType: "evaluate_condition"},
                {Name: "TrueBranch"},
                {Name: "FalseBranch"},
            },
            edges: []ActionEdge{
                {EdgeType: "start", TargetActionID: conditionID},
                {EdgeType: "true_branch", SourceActionID: &conditionID, TargetActionID: trueID},
                {EdgeType: "false_branch", SourceActionID: &conditionID, TargetActionID: falseID},
            },
            conditionResult: true,
            wantExecuted: []string{"Condition", "TrueBranch"},
        },
        // ... test complex graphs, parallel branches, etc.
    }
}
```

## Integration Tests (API Layer)

**New Directory**: `api/cmd/services/ichor/tests/workflow/edgeapi/`

**File**: `edge_test.go`
```go
func Test_EdgeAPI(t *testing.T) {
    test := apitest.StartTest(t, "edgeapi_test")
    sd := test.SeedData()

    test.Run(t, createEdge200(sd), "create-200")
    test.Run(t, createEdgeInvalidType400(sd), "create-invalid-type-400")
    test.Run(t, queryEdges200(sd), "query-200")
    test.Run(t, deleteEdge200(sd), "delete-200")
    test.Run(t, deleteEdgeNotFound404(sd), "delete-not-found-404")
}
```

**File**: `seed_test.go` - Seed test rules and actions

**New File**: `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go`
```go
func Test_CascadeMap(t *testing.T) {
    test := apitest.StartTest(t, "cascade_test")
    sd := test.SeedData()

    test.Run(t, cascadeMapWithDownstream200(sd), "cascade-with-downstream-200")
    test.Run(t, cascadeMapNoDownstream200(sd), "cascade-no-downstream-200")
    test.Run(t, cascadeMapRuleNotFound404(sd), "cascade-not-found-404")
}
```

## End-to-End Tests

**New File**: `api/cmd/services/ichor/tests/workflow/ruleapi/branching_e2e_test.go`
```go
func Test_WorkflowBranching_E2E(t *testing.T) {
    // 1. Create rule with condition action
    // 2. Create true/false branch actions
    // 3. Create edges connecting them
    // 4. Trigger workflow with data that evaluates to TRUE
    // 5. Verify only true branch action executed
    // 6. Trigger workflow with data that evaluates to FALSE
    // 7. Verify only false branch action executed
}
```

---

# Files Summary

## Files to Modify
| File | Changes |
|------|---------|
| `business/sdk/migrate/sql/migrate.sql` | Add action_edges table |
| `business/sdk/workflow/models.go` | Add BranchTaken, ActionEdge, ConditionResult |
| `business/sdk/workflow/interfaces.go` | Add EntityModifier interface |
| `business/sdk/workflow/executor.go` | Add ExecuteRuleActionsGraph method |
| `business/sdk/workflow/workflowactions/register.go` | Register condition handler |
| `business/sdk/workflow/stores/workflowdb/workflowdb.go` | Add edge CRUD methods |
| `business/sdk/workflow/workflowactions/data/updatefield.go` | Implement EntityModifier |
| `api/domain/http/workflow/ruleapi/route.go` | Add cascade endpoint |
| `api/cmd/services/ichor/build/all/all.go` | Wire up edge API |

## New Files to Create
| File | Purpose |
|------|---------|
| `business/sdk/workflow/workflowactions/control/condition.go` | Condition handler |
| `business/sdk/workflow/stores/workflowdb/actionedge.go` | Edge database ops |
| `api/domain/http/workflow/edgeapi/edgeapi.go` | Edge API handlers |
| `api/domain/http/workflow/edgeapi/route.go` | Edge API routes |
| `api/domain/http/workflow/edgeapi/model.go` | Edge app models |
| `api/domain/http/workflow/ruleapi/cascade.go` | Cascade handler |
| `business/sdk/workflow/condition_test.go` | Condition tests |
| `business/sdk/workflow/graph_executor_test.go` | Graph executor tests |
| `api/cmd/services/ichor/tests/workflow/edgeapi/*` | Edge API tests |
| `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go` | Cascade tests |
| `api/cmd/services/ichor/tests/workflow/ruleapi/branching_e2e_test.go` | E2E tests |

---

# Verification Steps

1. **Run migrations**: `make migrate`

2. **Build and lint**:
   ```bash
   go build ./...
   make lint
   ```

3. **Run all tests**:
   ```bash
   make test
   ```

4. **Run new tests specifically**:
   ```bash
   go test -v ./business/sdk/workflow/... -run "Condition|Graph"
   go test -v ./api/cmd/services/ichor/tests/workflow/edgeapi/...
   go test -v ./api/cmd/services/ichor/tests/workflow/ruleapi/... -run "Cascade|Branching"
   ```

5. **Manual verification**:
   - Create rule with condition action via API
   - Create edges for true/false branches
   - Trigger workflow and verify correct branch executes
   - Call cascade-map endpoint and verify response

---

# Phase 12 Testing - Comprehensive Unit & Integration Tests

**Date Added**: 2026-02-03
**Purpose**: Add rigorous testing for Phase 12 implementation (Condition Nodes + Cascade Visualization)

---

## Testing Phases Overview

| Phase | Name | Status | Priority | Jump To |
|-------|------|--------|----------|---------|
| 12.9 | Condition Handler Unit Tests | ✅ | Critical | [Details](#phase-129-condition-handler-unit-tests) |
| 12.10 | Graph Executor Unit Tests | ✅ | Critical | [Details](#phase-1210-graph-executor-unit-tests) |
| 12.11 | Edge API Integration Tests | ✅ | High | [Details](#phase-1211-edge-api-integration-tests) |
| 12.12 | Cascade API Integration Tests | ✅ | High | [Details](#phase-1212-cascade-api-integration-tests) |
| 12.13 | End-to-End Workflow Branching Tests | ✅ | Critical | [Details](#phase-1213-end-to-end-workflow-branching-tests) |

---

## Summary of Testing Gaps

| Component | Current Coverage | Gap |
|-----------|-----------------|-----|
| Graph-based executor (`ExecuteRuleActionsGraph`) | ✅ 100% | Covered in `executor_graph_test.go` |
| Condition handler (`evaluate_condition`) | ✅ 100% | Covered in `condition_test.go` |
| Edge management (CRUD) | ✅ 100% | Covered in `edgeapi/edge_test.go` (23 tests) |
| Cascade visualization API | ✅ 100% | Covered in `ruleapi/cascade_test.go` (9 tests) |
| Branch following logic (`shouldFollowEdge`) | ✅ 100% | Covered in `executor_graph_test.go` |

---

## Phase 12.9: Condition Handler Unit Tests

**File**: `business/sdk/workflow/workflowactions/control/condition_test.go`

### 12.9.1 Validation Tests

```go
// Test cases to implement:
func TestValidate_EmptyConditions(t *testing.T)    // Error when no conditions provided
func TestValidate_MissingFieldName(t *testing.T)   // Error when field_name is empty
func TestValidate_MissingOperator(t *testing.T)    // Error when operator is empty
func TestValidate_InvalidOperator(t *testing.T)    // Error for unknown operator
func TestValidate_InvalidLogicType(t *testing.T)   // Error for invalid logic_type
func TestValidate_ValidConfig(t *testing.T)        // Success for well-formed config
```

### 12.9.2 Operator Tests (10 operators)

```go
// equals operator
func TestOperator_Equals_Match(t *testing.T)       // equals: "active" == "active"
func TestOperator_Equals_NoMatch(t *testing.T)     // equals: "active" != "inactive"

// not_equals operator
func TestOperator_NotEquals_Match(t *testing.T)    // not_equals: "active" != "inactive"
func TestOperator_NotEquals_NoMatch(t *testing.T)  // not_equals: "active" == "active"

// greater_than operator
func TestOperator_GreaterThan_Numeric(t *testing.T) // greater_than: 150 > 100
func TestOperator_GreaterThan_String(t *testing.T)  // greater_than: "b" > "a"

// less_than operator
func TestOperator_LessThan_Numeric(t *testing.T)   // less_than: 50 < 100

// contains operator
func TestOperator_Contains_Match(t *testing.T)     // contains: "hello world" contains "world"
func TestOperator_Contains_NoMatch(t *testing.T)   // contains: "hello" not contains "world"

// in operator
func TestOperator_In_Match(t *testing.T)           // in: "red" in ["red", "green", "blue"]
func TestOperator_In_NoMatch(t *testing.T)         // in: "yellow" not in ["red", "green"]

// is_null / is_not_null operators
func TestOperator_IsNull_True(t *testing.T)        // is_null: nil == nil
func TestOperator_IsNull_False(t *testing.T)       // is_null: "value" != nil
func TestOperator_IsNotNull_True(t *testing.T)     // is_not_null: "value" != nil

// changed_from / changed_to operators
func TestOperator_ChangedFrom_Match(t *testing.T)  // changed_from: old="draft" matches
func TestOperator_ChangedFrom_NoUpdate(t *testing.T) // changed_from: fails on on_create
func TestOperator_ChangedTo_Match(t *testing.T)    // changed_to: new="shipped" from "pending"
func TestOperator_ChangedTo_SameValue(t *testing.T) // changed_to: fails when old==new
```

### 12.9.3 Logic Combination Tests

```go
func TestLogic_And_AllTrue(t *testing.T)           // AND: all conditions true = true
func TestLogic_And_OneFalse(t *testing.T)          // AND: one false = false
func TestLogic_Or_AllFalse(t *testing.T)           // OR: all false = false
func TestLogic_Or_OneTrue(t *testing.T)            // OR: one true = true
func TestLogic_DefaultIsAnd(t *testing.T)          // Empty logic_type defaults to AND
```

### 12.9.4 Branch Result Tests

```go
func TestExecute_ReturnsConditionResult(t *testing.T) // Returns proper ConditionResult struct
func TestExecute_TrueBranch(t *testing.T)          // BranchTaken = "true_branch" when true
func TestExecute_FalseBranch(t *testing.T)         // BranchTaken = "false_branch" when false
```

### 12.9.5 Edge Cases

```go
func TestEdgeCase_NilData(t *testing.T)            // Handles nil RawData gracefully
func TestEdgeCase_MissingField(t *testing.T)       // Field not in data returns nil
func TestEdgeCase_TypeMismatch(t *testing.T)       // Numeric comparison with strings
func TestEdgeCase_JsonNumber(t *testing.T)         // Handles json.Number type correctly
```

### Verification for Phase 12.9

```bash
go test -v ./business/sdk/workflow/workflowactions/control/... -run "Test"
```

---

## Phase 12.10: Graph Executor Unit Tests ✅ COMPLETE

**File**: `business/sdk/workflow/executor_graph_test.go`

> **Status**: All tests implemented in `executor_graph_test.go`. Run `go test -v ./business/sdk/workflow/... -run "GraphExec|ShouldFollowEdge"` to verify.

### 12.10.1 Backwards Compatibility

```go
func TestGraphExec_NoEdges_FallsBackToLinear(t *testing.T)    // When no edges, uses execution_order
func TestGraphExec_EmptyEdges_FallsBackToLinear(t *testing.T) // Empty edge slice = linear execution
```

### 12.10.2 Start Edge Tests

```go
func TestGraphExec_SingleStartEdge(t *testing.T)     // One start edge → one entry point
func TestGraphExec_MultipleStartEdges(t *testing.T)  // Multiple entry points execute in order
func TestGraphExec_NoStartEdge(t *testing.T)         // Error or empty result when no start edge
```

### 12.10.3 Sequential Execution Tests

```go
func TestGraphExec_LinearChain(t *testing.T)         // A → B → C executes in order
func TestGraphExec_SequenceEdgeType(t *testing.T)    // edge_type="sequence" always follows
func TestGraphExec_AlwaysEdgeType(t *testing.T)      // edge_type="always" always follows
```

### 12.10.4 Branch Execution Tests

```go
func TestGraphExec_TrueBranch_WhenConditionTrue(t *testing.T)  // Follows true_branch when result=true
func TestGraphExec_FalseBranch_WhenConditionFalse(t *testing.T) // Follows false_branch when result=false
func TestGraphExec_SkipsTrueBranch_WhenFalse(t *testing.T)     // Doesn't follow true_branch when false
func TestGraphExec_SkipsFalseBranch_WhenTrue(t *testing.T)     // Doesn't follow false_branch when true
```

### 12.10.5 Complex Graph Tests

```go
// Diamond pattern: A → Cond → B/C → D (converging)
func TestGraphExec_DiamondPattern(t *testing.T)

// Multiple branches from one node
func TestGraphExec_ParallelBranches(t *testing.T)

// Condition → Condition → Action
func TestGraphExec_DeepNesting(t *testing.T)
```

### 12.10.6 Cycle Prevention

```go
func TestGraphExec_NoCycleInfiniteLoop(t *testing.T) // Visited nodes not re-executed
func TestGraphExec_SelfLoop_Ignored(t *testing.T)    // A → A edge doesn't infinite loop
```

### 12.10.7 Edge Order Tests

```go
func TestGraphExec_EdgeOrderRespected(t *testing.T)      // Lower edge_order executes first
func TestGraphExec_DeterministicExecution(t *testing.T)  // Same graph = same execution order
```

### 12.10.8 shouldFollowEdge Unit Tests

```go
func TestShouldFollow_Always(t *testing.T)              // Always returns true
func TestShouldFollow_Sequence(t *testing.T)            // Sequence returns true
func TestShouldFollow_Start(t *testing.T)               // Start returns false (not a real edge)
func TestShouldFollow_TrueBranch_Match(t *testing.T)    // true_branch + BranchTaken=true_branch
func TestShouldFollow_TrueBranch_NoMatch(t *testing.T)  // true_branch + BranchTaken=false_branch
func TestShouldFollow_FalseBranch_Match(t *testing.T)   // false_branch + BranchTaken=false_branch
func TestShouldFollow_FalseBranch_NoMatch(t *testing.T) // false_branch + BranchTaken=true_branch
```

### Verification for Phase 12.10

```bash
go test -v ./business/sdk/workflow/... -run "GraphExec|ShouldFollow"
```

---

## Phase 12.11: Edge API Integration Tests

**Directory**: `api/cmd/services/ichor/tests/workflow/edgeapi/`

### Files to Create

| File | Purpose |
|------|---------|
| `seed_test.go` | Test data seeding |
| `edge_test.go` | Main test orchestrator |
| `create_test.go` | Create edge tests |
| `query_test.go` | Query edges tests |
| `delete_test.go` | Delete edge tests |

### 12.11.1 seed_test.go Structure

```go
type EdgeSeedData struct {
    apitest.SeedData
    Rules       []workflow.AutomationRule
    Actions     []workflow.RuleAction
    Edges       []workflow.ActionEdge
    OtherRule   workflow.AutomationRule  // For cross-rule validation tests
    OtherAction workflow.RuleAction      // Action in OtherRule
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (EdgeSeedData, error) {
    // 1. Create admin user with permissions
    // 2. Create trigger types, entity types, entities
    // 3. Create 2 automation rules (one for testing, one for cross-rule validation)
    // 4. Create 3+ actions per rule
    // 5. Create some initial edges for query tests
    // 6. Set up table access permissions
}
```

### 12.11.2 edge_test.go (Main Orchestrator)

```go
func Test_EdgeAPI(t *testing.T) {
    t.Parallel()

    test := apitest.StartTest(t, "Test_EdgeAPI")

    sd, err := insertSeedData(test.DB, test.Auth)
    if err != nil {
        t.Fatalf("Seeding error: %s", err)
    }

    // Create edge tests
    test.Run(t, createEdge200(sd), "createEdge-200")
    test.Run(t, createEdgeInvalidType400(sd), "createEdge-invalid-type-400")
    test.Run(t, createEdgeMissingTarget400(sd), "createEdge-missing-target-400")
    test.Run(t, createEdgeTargetNotFound404(sd), "createEdge-target-not-found-404")
    test.Run(t, createEdgeSourceNotFound404(sd), "createEdge-source-not-found-404")
    test.Run(t, createEdgeActionNotInRule400(sd), "createEdge-action-not-in-rule-400")
    test.Run(t, createEdgeRuleNotFound404(sd), "createEdge-rule-not-found-404")
    test.Run(t, createEdge401(sd), "createEdge-401")

    // Query edge tests
    test.Run(t, queryEdges200(sd), "queryEdges-200")
    test.Run(t, queryEdgesEmpty200(sd), "queryEdges-empty-200")
    test.Run(t, queryEdgesRuleNotFound404(sd), "queryEdges-rule-not-found-404")
    test.Run(t, queryEdges401(sd), "queryEdges-401")
    test.Run(t, queryEdgeByID200(sd), "queryEdgeByID-200")
    test.Run(t, queryEdgeByIDNotFound404(sd), "queryEdgeByID-not-found-404")
    test.Run(t, queryEdgeByIDWrongRule404(sd), "queryEdgeByID-wrong-rule-404")

    // Delete edge tests
    test.Run(t, deleteEdge200(sd), "deleteEdge-200")
    test.Run(t, deleteEdgeNotFound404(sd), "deleteEdge-not-found-404")
    test.Run(t, deleteEdgeWrongRule404(sd), "deleteEdge-wrong-rule-404")
    test.Run(t, deleteEdge401(sd), "deleteEdge-401")
    test.Run(t, deleteAllEdges200(sd), "deleteAllEdges-200")
    test.Run(t, deleteAllEdgesRuleNotFound404(sd), "deleteAllEdges-rule-not-found-404")
}
```

### 12.11.3 create_test.go

```go
// Create edge - success cases
func createEdge200(sd EdgeSeedData) []apitest.Table {
    // POST /v1/workflow/rules/{ruleID}/edges
    // Valid edge_type: start, sequence, true_branch, false_branch, always
    // source_action_id can be nil for start edges
    // Returns 200 with created edge
}

func createEdgeStartEdge200(sd EdgeSeedData) []apitest.Table {
    // Start edge has source_action_id = nil
}

// Create edge - validation errors (400)
func createEdgeInvalidType400(sd EdgeSeedData) []apitest.Table {
    // edge_type not in allowed list
}

func createEdgeMissingTarget400(sd EdgeSeedData) []apitest.Table {
    // target_action_id is required
}

func createEdgeActionNotInRule400(sd EdgeSeedData) []apitest.Table {
    // target_action_id belongs to a different rule
}

func createEdgeDuplicateEdge400(sd EdgeSeedData) []apitest.Table {
    // Same source→target→type already exists (unique constraint)
}

// Create edge - not found errors (404)
func createEdgeTargetNotFound404(sd EdgeSeedData) []apitest.Table {
    // target_action_id doesn't exist
}

func createEdgeSourceNotFound404(sd EdgeSeedData) []apitest.Table {
    // source_action_id doesn't exist (when not nil)
}

func createEdgeRuleNotFound404(sd EdgeSeedData) []apitest.Table {
    // ruleID in path doesn't exist
}

// Create edge - auth errors
func createEdge401(sd EdgeSeedData) []apitest.Table {
    // No auth token
}
```

### 12.11.4 query_test.go

```go
// Query edges for rule - success cases
func queryEdges200(sd EdgeSeedData) []apitest.Table {
    // GET /v1/workflow/rules/{ruleID}/edges
    // Returns all edges for the rule ordered by edge_order
}

func queryEdgesEmpty200(sd EdgeSeedData) []apitest.Table {
    // Rule exists but has no edges
    // Returns empty array
}

func queryEdgesOrderedByEdgeOrder200(sd EdgeSeedData) []apitest.Table {
    // Verify edges are sorted by edge_order ASC
}

// Query single edge - success cases
func queryEdgeByID200(sd EdgeSeedData) []apitest.Table {
    // GET /v1/workflow/rules/{ruleID}/edges/{edgeID}
    // Returns single edge
}

// Query edges - not found errors
func queryEdgesRuleNotFound404(sd EdgeSeedData) []apitest.Table {
    // ruleID doesn't exist
}

func queryEdgeByIDNotFound404(sd EdgeSeedData) []apitest.Table {
    // edgeID doesn't exist
}

func queryEdgeByIDWrongRule404(sd EdgeSeedData) []apitest.Table {
    // edgeID exists but belongs to different rule
}

// Query edges - auth errors
func queryEdges401(sd EdgeSeedData) []apitest.Table {
    // No auth token
}
```

### 12.11.5 delete_test.go

```go
// Delete single edge - success cases
func deleteEdge200(sd EdgeSeedData) []apitest.Table {
    // DELETE /v1/workflow/rules/{ruleID}/edges/{edgeID}
    // Returns 200/204 on success
}

// Delete all edges for rule - success cases
func deleteAllEdges200(sd EdgeSeedData) []apitest.Table {
    // DELETE /v1/workflow/rules/{ruleID}/edges
    // Removes all edges for the rule
}

// Delete edge - not found errors
func deleteEdgeNotFound404(sd EdgeSeedData) []apitest.Table {
    // edgeID doesn't exist
}

func deleteEdgeWrongRule404(sd EdgeSeedData) []apitest.Table {
    // edgeID exists but belongs to different rule
}

func deleteAllEdgesRuleNotFound404(sd EdgeSeedData) []apitest.Table {
    // ruleID doesn't exist
}

// Delete edge - auth errors
func deleteEdge401(sd EdgeSeedData) []apitest.Table {
    // No auth token
}
```

### Verification for Phase 12.11

```bash
# Run all edge API tests
go test -v ./api/cmd/services/ichor/tests/workflow/edgeapi/...

# Run individual test files
go test -v ./api/cmd/services/ichor/tests/workflow/edgeapi/... -run "create"
go test -v ./api/cmd/services/ichor/tests/workflow/edgeapi/... -run "query"
go test -v ./api/cmd/services/ichor/tests/workflow/edgeapi/... -run "delete"
```

---

## Phase 12.12: Cascade API Integration Tests

**File**: `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go`

### Test Cases

```go
func cascadeMap200WithDownstream(sd RuleSeedData) []apitest.Table {
    // Rule with update_field action that modifies an entity
    // Other active rules listen to that entity
    // Verify downstream workflows are returned
}

func cascadeMap200NoDownstream(sd RuleSeedData) []apitest.Table {
    // Rule with send_email action (doesn't modify entities)
    // Verify empty downstream_workflows array
}

func cascadeMap200MultipleActions(sd RuleSeedData) []apitest.Table {
    // Rule with multiple actions, some modify entities, some don't
    // Verify correct cascade info for each action
}

func cascadeMapRuleNotFound404(sd RuleSeedData) []apitest.Table {
    // Non-existent rule ID
}

func cascadeMap401(sd RuleSeedData) []apitest.Table {
    // No auth token
}

func cascadeMapExcludesSelf(sd RuleSeedData) []apitest.Table {
    // Rule that modifies entity it also listens to
    // Verify it doesn't show itself as downstream
}

func cascadeMapOnlyActiveRules(sd RuleSeedData) []apitest.Table {
    // Inactive downstream rules should be excluded
}
```

### Add to rule_test.go

```go
// In Test_RuleAPI function, add:
test.Run(t, cascadeMap200WithDownstream(sd), "cascadeMap-with-downstream-200")
test.Run(t, cascadeMap200NoDownstream(sd), "cascadeMap-no-downstream-200")
test.Run(t, cascadeMap200MultipleActions(sd), "cascadeMap-multiple-actions-200")
test.Run(t, cascadeMapRuleNotFound404(sd), "cascadeMap-rule-not-found-404")
test.Run(t, cascadeMap401(sd), "cascadeMap-401")
test.Run(t, cascadeMapExcludesSelf(sd), "cascadeMap-excludes-self-200")
test.Run(t, cascadeMapOnlyActiveRules(sd), "cascadeMap-only-active-200")
```

### Verification for Phase 12.12

```bash
go test -v ./api/cmd/services/ichor/tests/workflow/ruleapi/... -run "Cascade"
```

---

## Phase 12.13: End-to-End Workflow Branching Tests ✅ COMPLETE

**Status**: All scenarios are already covered by `executor_graph_test.go`. No separate `branching_e2e_test.go` file is needed.

### Coverage Analysis

| Proposed Test | Existing Coverage |
|--------------|-------------------|
| Simple branch true path | `TestGraphExec_TrueBranch_WhenConditionTrue` |
| Simple branch false path | `TestGraphExec_FalseBranch_WhenConditionFalse` |
| Converging branches (diamond) | `TestGraphExec_DiamondPattern` |
| Nested conditions | `TestGraphExec_NestedConditions` |
| Backwards compatibility | `TestGraphExec_NoEdges_FallsBackToLinear` |

The existing tests in `executor_graph_test.go` use real database integration (via `dbtest.NewDatabase`), create actual rules/actions/edges via `workflowBus.CreateRule/CreateRuleAction/CreateActionEdge`, and exercise the full `ExecuteRuleActionsGraph` code path with real condition evaluation.

**Verification**: Run `go test -v ./business/sdk/workflow/... -run "GraphExec"` to confirm coverage.

---

### Original Proposed Implementation (Kept for Reference)

**File**: `api/cmd/services/ichor/tests/workflow/ruleapi/branching_e2e_test.go`

### Implementation Notes

**Key Architecture Decisions:**

1. **ActionExecutor Creation**: Tests must create their own `ActionExecutor` since `dbtest.BusDomain.Workflow` is only the business layer (no executor attached).

2. **Handler Registration**: Use `workflowactions.RegisterCoreActions()` to register handlers that don't require RabbitMQ:
   - `evaluate_condition` - Required for branching tests
   - `update_field` - For cascade/data tests
   - `seek_approval` - For approval flow tests
   - `send_email`, `send_notification` - Communication (no-op without SMTP)

3. **Test Pattern**:
```go
func testSomeBranchingScenario(t *testing.T, test *apitest.Test, sd BranchingSeedData) {
    ctx := context.Background()

    // Create executor with real handlers
    executor := workflow.NewActionExecutor(test.DB.Log, test.DB.DB, test.DB.BusDomain.Workflow)
    workflowactions.RegisterCoreActions(executor.GetRegistry(), test.DB.Log, test.DB.DB)

    // Build execution context
    execCtx := workflow.ActionExecutionContext{
        EntityID:      uuid.New(),
        EntityName:    "test_entity",
        EventType:     "on_update",
        RawData:       map[string]interface{}{"amount": 1500.0},
        RuleID:        &sd.SomeRule.ID,
        RuleName:      sd.SomeRule.Name,
        TriggerSource: workflow.TriggerSourceAutomation,
    }

    // Execute using graph executor
    result, err := executor.ExecuteRuleActionsGraph(ctx, sd.SomeRule.ID, execCtx)

    // Verify branch taken and actions executed
    for _, ar := range result.ActionResults {
        // Check ar.ActionID, ar.BranchTaken, ar.Status
    }
}
```

4. **Seed Data Structure**: Create rules with actions and edges in seed function, NOT in individual tests:
   - `BranchingSeedData` struct holds all pre-created rules, actions, edges
   - Each test scenario gets its own rule (SimpleBranchRule, ConvergingRule, NestedRule, NoEdgesRule)
   - Reuse `ConditionTemplate` (action_type: "evaluate_condition") across rules

5. **Imports Required**:
```go
import (
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
)
```

### 12.13.1 Simple Branch Execution

```
[Start] → [Condition: amount > 1000?]
              ├─ TRUE:  [Action A: create_alert "high_value"]
              └─ FALSE: [Action B: create_alert "standard"]
```

```go
func TestE2E_SimpleBranch_TruePath(t *testing.T) {
    // 1. Create rule with condition action
    // 2. Create action A (create_alert with message "high_value")
    // 3. Create action B (create_alert with message "standard")
    // 4. Create edges: start→condition, condition→A (true_branch), condition→B (false_branch)
    // 5. Trigger workflow with entity data: {amount: 1500}
    // 6. Verify action A executed (check for alert with "high_value")
    // 7. Verify action B did NOT execute
}

func TestE2E_SimpleBranch_FalsePath(t *testing.T) {
    // Same setup as above
    // Trigger with {amount: 500}
    // Verify action B executed, action A did NOT
}
```

### 12.13.2 Sequential After Branch

```
[Condition] → TRUE  → [Action A] → [Action C]
            → FALSE → [Action B] → [Action C]
```

```go
func TestE2E_ConvergingBranches(t *testing.T) {
    // Both paths should converge to Action C
    // Verify C executes regardless of which branch was taken
}
```

### 12.13.3 Nested Conditions

```
[Cond1: type=urgent?]
  └─ TRUE → [Cond2: priority>5?]
               ├─ TRUE:  [Escalate]
               └─ FALSE: [Standard]
  └─ FALSE → [Queue]
```

```go
func TestE2E_NestedConditions_UrgentHighPriority(t *testing.T) {
    // type=urgent, priority=8 → Escalate action
}

func TestE2E_NestedConditions_UrgentLowPriority(t *testing.T) {
    // type=urgent, priority=3 → Standard action
}

func TestE2E_NestedConditions_NotUrgent(t *testing.T) {
    // type=normal → Queue action (skips Cond2 entirely)
}
```

### 12.13.4 Backwards Compatibility

```go
func TestE2E_NoEdges_LinearExecution(t *testing.T) {
    // Rule with actions but no edges
    // Should execute by execution_order
}

func TestE2E_MixedRules(t *testing.T) {
    // Two rules: one with edges, one without
    // Both should execute correctly in same workflow trigger
}
```

### Verification for Phase 12.13

```bash
# Phase 12.13 tests are in executor_graph_test.go, NOT branching_e2e_test.go
go test -v ./business/sdk/workflow/... -run "GraphExec"
```

---

## Files Summary

### New Test Files to Create

| File | Est. Lines | Phase | Status |
|------|------------|-------|--------|
| `business/sdk/workflow/workflowactions/control/condition_test.go` | ~400 | 12.9 | ✅ Created |
| `business/sdk/workflow/executor_graph_test.go` | ~1200 | 12.10 + 12.13 | ✅ Created (covers both phases) |
| `api/cmd/services/ichor/tests/workflow/edgeapi/seed_test.go` | ~150 | 12.11.1 | ✅ Created |
| `api/cmd/services/ichor/tests/workflow/edgeapi/edge_test.go` | ~57 | 12.11.2 | ✅ Created (23 tests) |
| `api/cmd/services/ichor/tests/workflow/edgeapi/create_test.go` | ~200 | 12.11.3 | ✅ Created |
| `api/cmd/services/ichor/tests/workflow/edgeapi/query_test.go` | ~150 | 12.11.4 | ✅ Created |
| `api/cmd/services/ichor/tests/workflow/edgeapi/delete_test.go` | ~150 | 12.11.5 | ✅ Created |
| `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go` | ~425 | 12.12 | ✅ Created (9 tests) |
| `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_seed_test.go` | ~150 | 12.12 | ✅ Created |
| ~~`api/cmd/services/ichor/tests/workflow/ruleapi/branching_e2e_test.go`~~ | ~~350~~ | ~~12.13~~ | ❌ Not needed - covered by executor_graph_test.go |

### Reference Files (Testing Patterns)

| File | Use For |
|------|---------|
| `business/sdk/workflow/executor_test.go` | Unit test patterns |
| `api/cmd/services/ichor/tests/workflow/ruleapi/rule_test.go` | API test orchestration |
| `api/cmd/services/ichor/tests/workflow/ruleapi/seed_test.go` | Seeding pattern |
| `api/cmd/services/ichor/tests/workflow/ruleapi/create_test.go` | Create test pattern |

---

## Final Verification

After all phases complete:

```bash
# Run full test suite
make test

# Run all new workflow tests (Phases 12.9, 12.10, 12.13 - all complete)
go test -v ./business/sdk/workflow/... -run "Condition|GraphExec|ShouldFollowEdge"
go test -v ./api/cmd/services/ichor/tests/workflow/edgeapi/...
go test -v ./api/cmd/services/ichor/tests/workflow/ruleapi/... -run "Cascade"

# Check coverage
go test -cover ./business/sdk/workflow/...
```

### Expected Outcomes
- All 40+ new test cases pass (reduced from 50+ since Phase 12.13 is covered by Phase 12.10)
- No regressions in existing workflow tests
- Code coverage for Phase 12 features > 80%

### Current Status (as of 2026-02-03)
| Phase | Status | Test Count |
|-------|--------|------------|
| 12.9 | ✅ Complete | ~25 tests (condition_test.go) |
| 12.10 | ✅ Complete | ~20 tests (executor_graph_test.go) |
| 12.11 | ✅ Complete | 23 tests (edgeapi/edge_test.go) |
| 12.12 | ✅ Complete | 9 tests (ruleapi/cascade_test.go) |
| 12.13 | ✅ Complete | Covered by Phase 12.10 |

**ALL TESTING PHASES COMPLETE** 🎉

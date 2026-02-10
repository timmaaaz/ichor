# Workflow Architecture: Async Continuation & Parallel Branches

## Vision

Users draw workflows as visual graphs (nodes + edges). Each workflow is ONE cohesive unit that:
- Executes deterministically along each path
- Supports **parallel branches** that run concurrently (via RabbitMQ)
- **Resumes after async actions complete** (not spawn separate rules)
- Parallel branches that **converge** wait for all to complete; branches that don't converge are fire-and-forget
- Conditions evaluate **merged context** (original trigger data + all action results)

---

## Current Problem

> **Note**: The examples throughout this document use a simple "Line Item Allocation" workflow for clarity. In production, workflows can be significantly more complex with:
> - **Dozens of actions** across multiple branches
> - **Deeply nested conditions** with multiple evaluation points
> - **Multiple async operations** that must coordinate (e.g., allocate inventory, check credit, validate shipping, notify warehouse)
> - **Complex convergence patterns** where 5+ branches must complete before proceeding
> - **Chained async operations** where one async result triggers another async action
> - **Error recovery branches** with retry logic and fallback paths
> - **Time-based gates** that pause for approval or external confirmation
>
> The architecture must handle these complex scenarios while the simple examples below illustrate the core concepts.

The "Line Item Allocation" flow is split across 3 separate rules when it should be ONE workflow:

```
CURRENT (Broken):
┌─────────────────────────────────────┐
│ Rule 1: Line Item Created           │  ← Condition evaluates BEFORE
│   → Allocate (async)                │    allocation completes (always false)
│   → Condition (broken)              │
│   → Alert (false positive)          │
└─────────────────────────────────────┘
           ↓ (event chain - separate rules)
┌─────────────────────────────────────┐  ┌─────────────────────────────────┐
│ Rule 2: Allocation Success          │  │ Rule 3: Allocation Failed       │
│ Trigger: allocation_results.on_create  │ Trigger: allocation_results.on_create
│ Condition: status == "success"      │  │ Condition: status == "failed"   │
│ Action: Update line item status     │  │ Action: Create alert            │
└─────────────────────────────────────┘  └─────────────────────────────────┘

DESIRED (Single Workflow):
┌─────────────────────────────────────────────────────────────────────────────┐
│ Workflow: Line Item Created - Allocate Inventory                            │
│                                                                             │
│   [Trigger: order_line_items.on_create]                                     │
│       │                                                                     │
│       ▼                                                                     │
│   [Allocate Inventory] ──(async)──► pauses workflow                         │
│       │                                                                     │
│       │ (async completes, workflow resumes with merged context)             │
│       ▼                                                                     │
│   [Condition: allocation.status == "success"]                               │
│       │                                                                     │
│   ┌───┴───┐                                                                 │
│   ▼       ▼                                                                 │
│ [Update  [Create                                                            │
│  Status]  Alert]                                                            │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Root Cause Analysis

### The Timeline (Current Broken Behavior)

1. **T0**: Line item created with `line_item_fulfillment_statuses_id = NULL`
2. **T1**: `on_create` event fires, RawData contains original line item data
3. **T2**: Allocate Inventory action executes → queues to RabbitMQ → returns immediately
4. **T3**: Condition evaluates `NULL == "ALLOCATED"` → **FALSE** → takes `false_branch` → creates alert
5. **T4**: (LATER) Async allocation completes → creates `allocation_results` record
6. **T5**: (LATER) Separate Rule 2 fires → updates `line_item_fulfillment_statuses_id = "ALLOCATED"`

**Problem**: The condition at T3 evaluates before T5 updates the field.

### Why This Happens

From `condition.go` line 119:
```go
result := h.evaluateConditions(cfg.Conditions, execCtx.RawData, execCtx.FieldChanges, execCtx.EventType, logicType)
```

The condition evaluates against `execCtx.RawData` - the **original line item data** at trigger time, not the allocation result.

### Evidence from Logs

```
T0: Allocate Inventory queued
allocate.go:405: VERBOSE: Successfully published to RabbitMQ: allocation_id[4c2607cc...]

T1: Condition evaluates (before allocation completes)
condition.go:126: evaluate_condition action executed: result[false]: branch_taken[false_branch]

T2: Alert created (false branch)
alert.go:201: create_alert action executed

T3: LATER - Allocation actually completes
allocate.go:624: Allocation completed: total_allocated[12]: status[success]

T4: LATER - Separate rule updates line item
updatefield.go:197: Field update completed: field[line_item_fulfillment_statuses_id]
```

---

## Architecture Design: Workflow Continuation

### Core Concepts

1. **Workflow Execution State**: When an async action runs, the workflow pauses and saves its state
2. **Continuation**: When async completes, it resumes the paused workflow from the next node
3. **Merged Context**: Each action's result is added to the execution context; subsequent actions see all prior results
4. **Parallel Branches**: Multiple outgoing edges from a node run concurrently via RabbitMQ
5. **Convergence (Join)**: If branches converge to a single node, wait for all; otherwise fire-and-forget

### Context Merging

After each action completes, its result is merged into the execution context:

```go
type ExecutionContext struct {
    // Original trigger data
    TriggerData map[string]interface{}

    // Results from each completed action, keyed by action ID or name
    ActionResults map[string]interface{}

    // Flattened view for template substitution
    // Priority: ActionResults > TriggerData
    MergedData map[string]interface{}
}

// After allocation completes:
context.ActionResults["allocate_inventory"] = map[string]interface{}{
    "status": "success",
    "total_allocated": 12,
    "allocation_id": "4c2607cc-...",
}

// Condition can now check:
// - {{allocate_inventory.status}} == "success"
// - {{allocation.status}} == "success" (if using action name)
```

---

## New Components

### 1. Workflow Execution State Table

```sql
-- Version: X.XX
-- Description: Add workflow execution state for async continuation

CREATE TABLE workflow.execution_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES workflow.automation_rules(id) ON DELETE CASCADE,
    execution_id UUID NOT NULL,  -- Links to automation_executions

    -- Position in graph
    current_action_id UUID REFERENCES workflow.rule_actions(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('running', 'awaiting', 'completed', 'failed', 'timed_out')),

    -- Context preservation (merged data)
    trigger_event JSONB NOT NULL,           -- Original trigger data
    action_results JSONB DEFAULT '{}',      -- Results from completed actions (keyed by action_id)
    merged_context JSONB DEFAULT '{}',      -- Flattened view for current evaluation

    -- Async tracking
    awaiting_action_id UUID,                -- Which async action we're waiting for
    awaiting_correlation_key TEXT,          -- How to match the completion event

    -- Parallel branch tracking
    parallel_branch_id UUID,                -- If this is a branch, which parent execution
    pending_branches UUID[] DEFAULT '{}',   -- Action IDs of branches not yet complete
    completed_branches UUID[] DEFAULT '{}', -- Branches that have finished

    -- Metadata
    created_date TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_date TIMESTAMP NOT NULL DEFAULT NOW(),
    timeout_at TIMESTAMP,                   -- When to consider this execution timed out

    -- Indexes
    CONSTRAINT unique_correlation UNIQUE(awaiting_correlation_key)
);

CREATE INDEX idx_execution_states_status ON workflow.execution_states(status);
CREATE INDEX idx_execution_states_rule ON workflow.execution_states(rule_id);
CREATE INDEX idx_execution_states_correlation ON workflow.execution_states(awaiting_correlation_key);
CREATE INDEX idx_execution_states_parent ON workflow.execution_states(parallel_branch_id);
```

### 2. Execution State Model

```go
// business/sdk/workflow/execution_state.go

package workflow

import (
    "time"
    "github.com/google/uuid"
)

type ExecutionStatus string

const (
    ExecutionStatusRunning  ExecutionStatus = "running"
    ExecutionStatusAwaiting ExecutionStatus = "awaiting"
    ExecutionStatusCompleted ExecutionStatus = "completed"
    ExecutionStatusFailed   ExecutionStatus = "failed"
    ExecutionStatusTimedOut ExecutionStatus = "timed_out"
)

type ExecutionState struct {
    ID                    uuid.UUID
    RuleID                uuid.UUID
    ExecutionID           uuid.UUID
    CurrentActionID       *uuid.UUID
    Status                ExecutionStatus

    // Context
    TriggerEvent          map[string]interface{}
    ActionResults         map[string]interface{}  // action_id -> result
    MergedContext         map[string]interface{}  // Flattened for evaluation

    // Async
    AwaitingActionID      *uuid.UUID
    AwaitingCorrelationKey string

    // Parallel
    ParallelBranchID      *uuid.UUID
    PendingBranches       []uuid.UUID
    CompletedBranches     []uuid.UUID

    // Metadata
    CreatedDate           time.Time
    UpdatedDate           time.Time
    TimeoutAt             *time.Time
}

type NewExecutionState struct {
    RuleID         uuid.UUID
    ExecutionID    uuid.UUID
    TriggerEvent   map[string]interface{}
    TimeoutMinutes int  // 0 = no timeout
}

// MergeActionResult adds an action's result to the context
func (s *ExecutionState) MergeActionResult(actionID uuid.UUID, actionName string, result interface{}) {
    if s.ActionResults == nil {
        s.ActionResults = make(map[string]interface{})
    }
    if s.MergedContext == nil {
        s.MergedContext = make(map[string]interface{})
    }

    // Store by ID
    s.ActionResults[actionID.String()] = result

    // Also store by name for easier template access
    s.MergedContext[actionName] = result

    // Flatten result fields into merged context if it's a map
    if resultMap, ok := result.(map[string]interface{}); ok {
        for k, v := range resultMap {
            // Prefix with action name: "allocation.status"
            s.MergedContext[actionName+"."+k] = v
        }
    }
}
```

### 3. Modified Executor

```go
// business/sdk/workflow/executor.go (modifications)

// ExecuteWithState runs a workflow with state persistence for async continuation
func (e *Executor) ExecuteWithState(ctx context.Context, event TriggerEvent, rule AutomationRule) (*ExecutionState, error) {
    // Create or load execution state
    state, err := e.getOrCreateState(ctx, event, rule)
    if err != nil {
        return nil, err
    }

    // Get starting point
    startActionID := state.CurrentActionID
    if startActionID == nil {
        // Fresh execution - start from beginning
        startEdge, err := e.getStartEdge(ctx, rule.ID)
        if err != nil {
            return nil, err
        }
        startActionID = &startEdge.TargetActionID
    }

    // Execute from current position
    return e.executeFromAction(ctx, state, *startActionID)
}

func (e *Executor) executeFromAction(ctx context.Context, state *ExecutionState, actionID uuid.UUID) (*ExecutionState, error) {
    action, err := e.getAction(ctx, actionID)
    if err != nil {
        return state, err
    }

    // Build execution context with merged data
    execCtx := e.buildExecutionContext(state, action)

    // Execute the action
    handler := e.getHandler(action.ActionType)
    result, err := handler.Execute(ctx, action.Config, execCtx)

    if err != nil {
        state.Status = ExecutionStatusFailed
        return e.saveState(ctx, state)
    }

    // Merge result into context
    state.MergeActionResult(action.ID, action.Name, result)

    // Check if async
    if handler.IsAsync() {
        // Pause execution
        state.Status = ExecutionStatusAwaiting
        state.AwaitingActionID = &action.ID
        state.AwaitingCorrelationKey = e.generateCorrelationKey(state, action)
        state.CurrentActionID = &actionID
        return e.saveState(ctx, state)
    }

    // Get outgoing edges
    edges, err := e.getOutgoingEdges(ctx, actionID, result)
    if err != nil {
        return state, err
    }

    if len(edges) == 0 {
        // End of workflow
        state.Status = ExecutionStatusCompleted
        return e.saveState(ctx, state)
    }

    if len(edges) > 1 {
        // Parallel execution
        return e.executeParallel(ctx, state, edges)
    }

    // Sequential - continue to next
    return e.executeFromAction(ctx, state, edges[0].TargetActionID)
}

func (e *Executor) ResumeExecution(ctx context.Context, correlationKey string, result interface{}) (*ExecutionState, error) {
    // Load paused state
    state, err := e.stateStore.FindByCorrelationKey(ctx, correlationKey)
    if err != nil {
        return nil, fmt.Errorf("no execution state found for correlation key: %s", correlationKey)
    }

    // Merge the async result
    if state.AwaitingActionID != nil {
        action, _ := e.getAction(ctx, *state.AwaitingActionID)
        state.MergeActionResult(*state.AwaitingActionID, action.Name, result)
    }

    // Clear await state
    state.Status = ExecutionStatusRunning
    state.AwaitingActionID = nil
    state.AwaitingCorrelationKey = ""

    // Get next action(s) from the awaited action
    edges, err := e.getOutgoingEdges(ctx, *state.CurrentActionID, result)
    if err != nil || len(edges) == 0 {
        state.Status = ExecutionStatusCompleted
        return e.saveState(ctx, state)
    }

    // Continue execution
    return e.executeFromAction(ctx, state, edges[0].TargetActionID)
}

func (e *Executor) executeParallel(ctx context.Context, state *ExecutionState, edges []ActionEdge) (*ExecutionState, error) {
    // Detect if any branches converge
    convergePoint := e.findConvergencePoint(ctx, edges)

    for _, edge := range edges {
        willConverge := convergePoint != nil && e.pathLeadsTo(edge.TargetActionID, *convergePoint)

        if willConverge {
            state.PendingBranches = append(state.PendingBranches, edge.TargetActionID)
        }

        // Queue branch to RabbitMQ
        e.queueBranchExecution(ctx, state, edge, willConverge)
    }

    if len(state.PendingBranches) > 0 {
        // Wait for convergence
        state.Status = ExecutionStatusAwaiting
        return e.saveState(ctx, state)
    }

    // Fire and forget - mark complete
    state.Status = ExecutionStatusCompleted
    return e.saveState(ctx, state)
}

func (e *Executor) buildExecutionContext(state *ExecutionState, action RuleAction) ActionExecutionContext {
    return ActionExecutionContext{
        ExecutionID: state.ExecutionID.String(),
        RuleID:      &state.RuleID,
        EntityID:    state.TriggerEvent["entity_id"].(uuid.UUID),
        EntityName:  state.TriggerEvent["entity_name"].(string),
        EventType:   state.TriggerEvent["event_type"].(string),

        // Use merged context for evaluation
        RawData:      state.MergedContext,
        FieldChanges: state.TriggerEvent["field_changes"].(map[string]FieldChange),

        // Also expose action results explicitly
        ActionResults: state.ActionResults,
    }
}
```

### 4. Modified Async Handler (Allocation Example)

```go
// business/sdk/workflow/workflowactions/inventory/allocate.go (modifications)

// After ProcessAllocation completes, instead of publishing entity event:
func (h *AllocateInventoryHandler) onAsyncComplete(ctx context.Context, result *InventoryAllocationResult, request AllocationRequest) {
    // Build result map for context merging
    resultData := map[string]interface{}{
        "status":            result.Status,
        "total_allocated":   result.TotalAllocated,
        "allocation_id":     result.AllocationID.String(),
        "reference_id":      request.Config.ReferenceID,
        "reference_type":    request.Config.ReferenceType,
        "allocated_items":   result.AllocatedItems,
        "failed_items":      result.FailedItems,
    }

    // Resume the paused workflow with this result
    correlationKey := request.CorrelationKey  // Passed in from Execute()
    if err := h.executor.ResumeExecution(ctx, correlationKey, resultData); err != nil {
        h.log.Error(ctx, "Failed to resume workflow execution",
            "correlation_key", correlationKey,
            "error", err)
    }

    // Still publish entity event for other listeners (optional)
    h.fireAllocationResultEvent(ctx, result, request, h.publisher)
}
```

### 5. Updated Condition Handler

```go
// business/sdk/workflow/workflowactions/control/condition.go (modifications)

func (h *EvaluateConditionHandler) Execute(ctx context.Context, config json.RawMessage, execCtx ActionExecutionContext) (interface{}, error) {
    var cfg ConditionConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil, err
    }

    // Use merged context (includes action results) for evaluation
    // RawData now contains: original trigger data + all prior action results
    data := execCtx.RawData

    // Also support explicit action result references
    // e.g., condition on "allocate_inventory.status"
    if execCtx.ActionResults != nil {
        for k, v := range execCtx.ActionResults {
            if resultMap, ok := v.(map[string]interface{}); ok {
                for rk, rv := range resultMap {
                    // Support both "action_id.field" and "action_name.field"
                    data[k+"."+rk] = rv
                }
            }
        }
    }

    result := h.evaluateConditions(cfg.Conditions, data, execCtx.FieldChanges, execCtx.EventType, cfg.LogicType)

    // ... rest of handler
}
```

---

## Edge Types

### Current Edge Types
- `start` - Entry point (source_action_id = NULL)
- `sequence` - A → B (linear flow)
- `true_branch` - Condition evaluated true
- `false_branch` - Condition evaluated false
- `always` - Always execute after condition (parallel to branches)

### New/Clarified Edge Types
- `parallel` - Concurrent execution (same as having multiple edges from one node)
- `join` - Convergence point (detected automatically when node has multiple incoming edges)

**Note**: Parallel execution is implicit from the graph structure. If a node has multiple outgoing edges (other than true/false branches), they execute in parallel.

---

## Example Workflows

### Example 1: Line Item Allocation (Fixed)

```
Workflow: "Line Item Created - Allocate Inventory"
Trigger: order_line_items.on_create

Actions:
1. allocate_inventory (async)
   - Config: { source_from_line_item: true, allocation_mode: "reserve" }

2. evaluate_condition
   - Config: { conditions: [{ field: "allocate_inventory.status", operator: "equals", value: "success" }] }

3. update_field (true branch)
   - Config: { target_entity: "order_line_items", target_field: "status", new_value: "ALLOCATED" }

4. create_alert (false branch)
   - Config: { alert_type: "allocation_failed", severity: "high" }

Edges:
- start → action_1 (allocate_inventory)
- action_1 → action_2 (condition) [sequence]
- action_2 → action_3 (update_field) [true_branch]
- action_2 → action_4 (create_alert) [false_branch]
```

### Example 2: Order Creation with Parallel Branches

```
Workflow: "Order Created - Full Processing"
Trigger: orders.on_create

Actions:
1. send_email (notify sales team)
2. allocate_inventory (async)
3. create_alert (order received)
4. evaluate_condition (allocation success?)
5. update_field (set order status to ALLOCATED)
6. create_alert (allocation failed)
7. send_email (order confirmation to customer)

Edges:
- start → action_1 (email) [parallel branch 1]
- start → action_2 (allocate) [parallel branch 2]
- start → action_3 (alert) [parallel branch 3, fire-and-forget]
- action_1 → action_7 (join point)
- action_2 → action_4 (condition)
- action_4 → action_5 (update) [true_branch]
- action_4 → action_6 (alert) [false_branch]
- action_5 → action_7 (join point)
- action_6 → action_7 (join point)

Graph:
        ┌──────► [Email Sales] ────────────────────────┐
        │                                               │
[Trigger]──────► [Allocate] ──► [Condition] ──┬──► [Update] ──┬──► [Email Customer]
        │                                     │              │
        │                                     └──► [Alert]───┘
        │
        └──────► [Alert: Order Received] (fire-and-forget)
```

### Example 3: Complex Order Processing (Production-Scale)

This example demonstrates what real-world workflows might look like with multiple async operations, nested conditions, and complex branching:

```
Workflow: "High-Value Order Processing"
Trigger: orders.on_create
Condition: order.total > 10000

Actions (20+ actions):
1.  evaluate_condition (is_new_customer?)
2.  credit_check (async) - only for new customers
3.  evaluate_condition (credit_approved?)
4.  fraud_detection (async) - third-party API call
5.  evaluate_condition (fraud_score_acceptable?)
6.  manager_approval_request (async) - waits for human approval
7.  evaluate_condition (approved_by_manager?)
8.  allocate_inventory (async)
9.  evaluate_condition (allocation_success?)
10. reserve_shipping (async)
11. evaluate_condition (shipping_reserved?)
12. send_email (customer confirmation)
13. send_email (warehouse notification)
14. create_audit_log
15. update_field (order status → PROCESSING)
16. send_email (sales team notification)
17. create_alert (fraud detected)
18. create_alert (credit denied)
19. create_alert (inventory unavailable)
20. send_email (rejection notification to customer)
21. update_field (order status → REJECTED)
22. escalate_to_supervisor (async)

Graph (simplified view):
                                                    ┌─► [Email Sales]
                                                    │
[Trigger] ─► [New Customer?] ─┬─► [Credit Check] ─► [Credit OK?] ─┬─► [Fraud Check] ─► [Fraud OK?] ─┬─► ...continues
                              │                                    │                                 │
                              │                                    │                                 └─► [Alert: Fraud] ─► [Email Reject] ─► [Status: REJECTED]
                              │                                    │
                              │                                    └─► [Alert: Credit] ─► [Email Reject] ─► [Status: REJECTED]
                              │
                              └─► (skip credit check for existing customers) ─► [Fraud Check] ─► ...

...continues from [Fraud OK?] ─► [Manager Approval?] ─┬─► [Approved?] ─┬─► [Allocate] ─► [Alloc OK?] ─┬─► [Reserve Ship] ─► [Ship OK?] ─┬─► [PARALLEL BRANCH]
                                                      │                │                              │                                 │         │
                                                      │                │                              │                                 │    ┌────┴────┐
                                                      │                │                              │                                 │    ▼         ▼
                                                      │                │                              │                                 │  [Email   [Email
                                                      │                │                              │                                 │  Customer] Warehouse]
                                                      │                │                              │                                 │    │         │
                                                      │                │                              │                                 │    └────┬────┘
                                                      │                │                              │                                 │         ▼
                                                      │                │                              │                                 │    [CONVERGENCE]
                                                      │                │                              │                                 │         │
                                                      │                │                              │                                 │         ▼
                                                      │                │                              │                                 │    [Audit Log]
                                                      │                │                              │                                 │         │
                                                      │                │                              │                                 │         ▼
                                                      │                │                              │                                 │    [Status: PROCESSING]
                                                      │                │                              │                                 │
                                                      │                │                              │                                 └─► [Alert: Ship Failed] ─► [Escalate]
                                                      │                │                              │
                                                      │                │                              └─► [Alert: Inventory] ─► [Escalate]
                                                      │                │
                                                      │                └─► [Email Reject] ─► [Status: REJECTED]
                                                      │
                                                      └─► (timeout after 24h) ─► [Escalate to Supervisor]

Execution States Required:
- After credit_check: await async result, preserve trigger data
- After fraud_detection: await async result, preserve trigger + credit result
- After manager_approval: await human action (may take hours/days), preserve all prior context
- After allocate_inventory: await async result, preserve everything above
- After reserve_shipping: await async result, preserve everything above
- At convergence point: wait for both email branches to complete

Context at Final Step (merged data):
{
  "trigger": { "order_id": "...", "total": 15000, "customer_id": "..." },
  "credit_check": { "status": "approved", "credit_limit": 50000, "checked_at": "..." },
  "fraud_detection": { "score": 0.12, "status": "passed", "provider": "stripe_radar" },
  "manager_approval": { "approved": true, "approved_by": "jane@company.com", "notes": "VIP customer" },
  "allocate_inventory": { "status": "success", "warehouse": "WH-001", "items": [...] },
  "reserve_shipping": { "status": "confirmed", "carrier": "FedEx", "tracking": "..." }
}
```

This example shows:
- **5 async operations** that pause workflow execution
- **6 condition evaluations** with true/false branching
- **Parallel email notifications** that must converge
- **Human-in-the-loop approval** that may take hours or days
- **Cascading error handling** with escalation paths
- **Deep context accumulation** across all async boundaries

---

## Implementation Phases

### Phase 1: Execution State Persistence
**Goal**: Save and restore workflow execution state

**Tasks**:
1. Add `execution_states` migration
2. Create `ExecutionState` model and store
3. Modify executor to create state at start
4. Add `saveState()` and `loadState()` methods
5. Add correlation key generation

**Files**:
- `business/sdk/migrate/sql/migrate.sql`
- `business/sdk/workflow/execution_state.go` (new)
- `business/sdk/workflow/stores/executionstatedb/executionstatedb.go` (new)
- `business/sdk/workflow/executor.go`

**Verification**:
```bash
make migrate
go test -v ./business/sdk/workflow/stores/executionstatedb/...
```

### Phase 2: Async Continuation
**Goal**: Resume workflow after async action completes

**Tasks**:
1. Add `ResumeExecution()` to executor
2. Modify async handlers to call back on completion
3. Pass correlation key through async pipeline
4. Update condition to use merged context

**Files**:
- `business/sdk/workflow/executor.go`
- `business/sdk/workflow/workflowactions/inventory/allocate.go`
- `business/sdk/workflow/workflowactions/control/condition.go`
- `business/sdk/workflow/queue.go`

**Verification**:
```bash
# Create test workflow with async action followed by condition
go test -v ./business/sdk/workflow/... -run TestAsyncContinuation
```

### Phase 3: Parallel Branches
**Goal**: Execute multiple branches concurrently

**Tasks**:
1. Detect multiple outgoing edges
2. Queue each branch to RabbitMQ with parent state reference
3. Each branch creates its own sub-state
4. Fire-and-forget branches complete independently

**Files**:
- `business/sdk/workflow/executor.go`
- `business/sdk/workflow/queue.go`

**Verification**:
```bash
go test -v ./business/sdk/workflow/... -run TestParallelBranches
```

### Phase 4: Convergence (Join)
**Goal**: Wait for all converging branches before continuing

**Tasks**:
1. Implement `findConvergencePoint()` - graph traversal to find nodes with multiple incoming edges
2. Track `pending_branches` and `completed_branches` in state
3. When branch completes, check if all siblings done
4. Resume from convergence point when all complete
5. Merge results from all branches

**Files**:
- `business/sdk/workflow/executor.go`
- `business/sdk/workflow/graph_analysis.go` (new)

**Verification**:
```bash
go test -v ./business/sdk/workflow/... -run TestConvergence
```

### Phase 5: Timeout Handling
**Goal**: Handle workflows that get stuck

**Tasks**:
1. Add `timeout_at` calculation on state creation
2. Create background job to check for timed-out states
3. On timeout: mark state as `timed_out`, optionally execute timeout branch
4. Add `on_timeout` edge type (optional)

**Files**:
- `business/sdk/workflow/timeout_checker.go` (new)
- `api/cmd/services/ichor/main.go` (start timeout checker)

---

## Immediate Fix (Before Full Implementation)

To unblock testing immediately, remove the broken condition from the "Line Item Created" rule:

1. Keep the existing 3-rule structure (it works correctly via event chaining)
2. Remove the condition and alert actions from Rule 1
3. Rule 1 becomes: `Trigger → Allocate Inventory` (done)
4. Rules 2 & 3 handle success/failure (existing behavior)

This is a workaround until the full async continuation architecture is implemented.

---

## Files Summary

### New Files
- `business/sdk/workflow/execution_state.go` - State model
- `business/sdk/workflow/stores/executionstatedb/executionstatedb.go` - DB operations
- `business/sdk/workflow/stores/executionstatedb/model.go` - DB model
- `business/sdk/workflow/graph_analysis.go` - Convergence detection
- `business/sdk/workflow/timeout_checker.go` - Background timeout job

### Modified Files
- `business/sdk/migrate/sql/migrate.sql` - Add execution_states table
- `business/sdk/workflow/executor.go` - Pause/resume/parallel logic
- `business/sdk/workflow/workflowactions/inventory/allocate.go` - Callback on completion
- `business/sdk/workflow/workflowactions/control/condition.go` - Use merged context
- `business/sdk/workflow/queue.go` - Handle continuation messages
- `business/sdk/workflow/engine.go` - State management integration
- `business/sdk/workflow/models.go` - ActionExecutionContext changes

---

## Performance Analysis: Current vs. Proposed Architecture

### Current Architecture Performance

**Characteristics:**
- Each workflow is split into multiple independent rules (e.g., 3 rules for line item allocation)
- Rules trigger via entity events (database writes → event publish → rule evaluation)
- No state persistence between rule executions
- Each rule re-queries context from database

**Overhead per "logical workflow" (current approach):**
| Operation | Count | Latency | Notes |
|-----------|-------|---------|-------|
| Entity events published | N rules | ~1-5ms each | RabbitMQ publish |
| Rule evaluations | N rules | ~5-20ms each | Query rule, check conditions |
| Database context queries | N rules | ~2-10ms each | Each rule re-fetches entity data |
| Action executions | Same | Same | No difference |

**Example - Line Item Allocation (3 rules):**
- 3 entity event publishes: ~5ms
- 3 rule evaluations: ~30ms
- 3 context fetches: ~15ms
- **Overhead: ~50ms** (plus actual action execution time)

**Scaling issues at enterprise scale (10,000+ workflows/minute):**
1. **Event storm**: Each async completion publishes entity event, triggering evaluation of ALL rules with that trigger type
2. **N×M rule matching**: 100 rules × 1000 events/minute = 100,000 evaluations/minute
3. **Database pressure**: Repeated context queries for the same entities
4. **No correlation**: Cannot track workflow execution across rule boundaries
5. **Race conditions**: Multiple rules may react to same event simultaneously

### Proposed Architecture Performance

**Characteristics:**
- Single workflow execution with state persistence
- Correlation key for async continuation (no event matching needed)
- Context preserved in memory/state table (no re-query)
- Direct continuation instead of event-driven chaining

**Overhead per workflow (proposed approach):**
| Operation | Count | Latency | Notes |
|-----------|-------|---------|-------|
| State writes | 1 per async boundary | ~2-5ms | INSERT/UPDATE execution_states |
| State reads | 1 per resume | ~1-3ms | Direct lookup by correlation key |
| Context fetch | 1 (at trigger) | ~2-10ms | Only once, preserved in state |
| Continuation message | 1 per async | ~1ms | Direct queue to executor |

**Same Line Item Allocation (1 workflow):**
- 1 context fetch: ~5ms
- 1 state write (before async): ~3ms
- 1 state read (on resume): ~2ms
- 1 state update (completion): ~2ms
- **Overhead: ~12ms** (vs ~50ms current)

### Enterprise Scale Comparison

**Scenario: 10,000 orders/minute, each triggering complex 20-action workflow**

| Metric | Current (3-rule split) | Proposed (single workflow) |
|--------|------------------------|----------------------------|
| Events published/min | 30,000+ | 10,000 |
| Rule evaluations/min | 100,000+ (all rules × events) | 10,000 (direct continuation) |
| Database context queries/min | 30,000+ | 10,000 |
| State table operations/min | 0 | 40,000 (4 per workflow avg) |
| Total DB operations/min | 30,000+ | 50,000 |
| Memory for correlation | None (stateless) | ~500 bytes × active workflows |

**Net impact:**
- **Reduced event processing**: 3× fewer events, no broadcast matching
- **Eliminated rule matching overhead**: O(1) continuation vs O(N) rule evaluation
- **Tradeoff**: State table adds operations but removes context re-queries

### Latency Analysis

**Current approach - async completion to next action:**
```
T0: Async completes
T1: +2ms  Entity event published
T2: +5ms  Event delivered to rule engine
T3: +15ms All matching rules evaluated (N rules)
T4: +5ms  Matching rule(s) identified
T5: +10ms Context re-fetched from database
T6: +5ms  Action executed
────────────────────────────────
Total: ~40ms minimum
```

**Proposed approach - async completion to next action:**
```
T0: Async completes
T1: +1ms  Continuation message sent to executor
T2: +2ms  State loaded by correlation key
T3: +0ms  Context already in state (no fetch)
T4: +5ms  Action executed
────────────────────────────────
Total: ~8ms
```

**Improvement: ~5× lower latency for async continuation**

### Throughput Bottlenecks

**Current architecture bottlenecks:**
1. **Event bus saturation**: All events broadcast, all rules evaluate
2. **Database hot spots**: Same entities queried repeatedly
3. **Rule matching CPU**: Complex condition evaluation repeated

**Proposed architecture bottlenecks:**
1. **State table writes**: Can be mitigated with:
   - Partitioning by rule_id
   - Redis cache layer for hot states
   - Batch writes for parallel branches
2. **Memory for active workflows**: ~500 bytes × concurrent workflows
   - 10,000 concurrent = ~5MB (negligible)

### Recommendations for Enterprise Deployment

**State table optimization:**
```sql
-- Partition by date for cleanup
CREATE TABLE workflow.execution_states (
    ...
) PARTITION BY RANGE (created_date);

-- Create monthly partitions
CREATE TABLE workflow.execution_states_2024_01
    PARTITION OF workflow.execution_states
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Index for hot path
CREATE INDEX CONCURRENTLY idx_execution_states_correlation
    ON workflow.execution_states(awaiting_correlation_key)
    WHERE status = 'awaiting';
```

**Redis cache layer (optional):**
```go
// Cache frequently accessed states
type CachedStateStore struct {
    redis  *redis.Client
    db     *StateStore
    ttl    time.Duration // 5 minutes for awaiting states
}

func (s *CachedStateStore) FindByCorrelationKey(ctx context.Context, key string) (*ExecutionState, error) {
    // Try cache first
    if cached := s.redis.Get(ctx, "exec_state:"+key); cached != nil {
        return deserialize(cached), nil
    }
    // Fall back to DB
    state, err := s.db.FindByCorrelationKey(ctx, key)
    if err == nil {
        s.redis.Set(ctx, "exec_state:"+key, serialize(state), s.ttl)
    }
    return state, err
}
```

**Metrics to monitor:**
- `workflow_state_write_latency_ms` - State persistence time
- `workflow_continuation_latency_ms` - Time from async complete to next action
- `workflow_active_count` - Number of awaiting workflows
- `workflow_completion_rate` - Workflows completed per minute
- `workflow_timeout_rate` - Workflows that timed out

---

## Open Questions

1. **Correlation Key Strategy**:
   - Option A: `execution_id` (simplest, but need to pass through RabbitMQ)
   - Option B: `rule_id + entity_id + timestamp` (reconstructible)
   - Option C: UUID generated per async action (most flexible)
   - **Recommendation**: Option A with Option C fallback

2. **Branch Result Merging**:
   - When parallel branches converge, how to merge their results?
   - Option A: Last write wins
   - Option B: Namespace by branch/action name
   - Option C: Array of results
   - **Recommendation**: Option B (namespace by action name)

3. **Timeout Behavior**:
   - What happens on timeout?
   - Option A: Mark failed, no further action
   - Option B: Execute a `timeout_branch` if defined
   - Option C: Create alert
   - **Recommendation**: Option A initially, add Option B later

4. **Backward Compatibility**:
   - Existing workflows without state tracking?
   - **Answer**: All workflows get state tracking, just simpler state for sync-only flows

---

## Temporal Integration: Deep Analysis

This section evaluates whether Temporal (or similar workflow orchestration engines like Cadence) should replace the proposed custom continuation architecture.

### What Temporal Provides (That We'd Have to Build)

| Capability | Temporal | Custom (This Doc) |
|------------|----------|-------------------|
| **Durable execution** | Built-in. Survives crashes, restarts, deployments. Automatic recovery. | Must build: `execution_states` table, recovery logic, orphan detection |
| **Retry policies** | Per-activity configurable: exponential backoff, max attempts, non-retryable errors | Must build: retry table, retry worker, error classification |
| **Timeouts** | Activity timeout, workflow timeout, heartbeat timeout, schedule-to-start timeout | Must build: `timeout_checker` background job, timeout state transitions |
| **Visibility** | Built-in dashboard, query by custom attributes, full execution history | Must build: execution history UI, search/filter capabilities |
| **Versioning** | Change workflow definitions while executions are in-flight. Deterministic replay. | **Not addressed in this doc** - major gap |
| **Parallel execution** | Native `workflow.Go()` for concurrent activities, `workflow.Await()` for joins | Must build: branch tracking, convergence detection, state merging |
| **Compensation/Saga** | Built-in saga pattern support for rollbacks | Must build: compensation tracking, rollback orchestration |

### What Temporal Does NOT Provide

#### 1. Visual Graph Editor Integration

**The Problem**: Temporal workflows are *code*. Your visual editor produces a *data structure*.

```go
// Temporal workflow - compiled Go code
func OrderWorkflow(ctx workflow.Context, order Order) error {
    // This is statically defined at compile time
    if order.IsNewCustomer {
        workflow.ExecuteActivity(ctx, CreditCheck, order)
    }
    workflow.ExecuteActivity(ctx, FraudCheck, order)
    // ...
}
```

```json
// Your system - dynamic data structure
{
  "actions": [...],
  "edges": [
    {"source": "action_1", "target": "action_2", "edge_type": "true_branch"},
    {"source": "action_1", "target": "action_3", "edge_type": "false_branch"}
  ]
}
```

**Translation Layer Required**: You'd need a "generic interpreter workflow" that reads your graph and executes it dynamically. This means you're still writing graph traversal logic.

#### 2. Dynamic Workflow Definitions

Temporal workflows are typically deployed as compiled code. Your system allows:
- Users create/modify workflows at runtime via UI
- No deployment required for workflow changes
- Immediate effect on new executions

**Options**:
- **Generic interpreter workflow**: One Temporal workflow that interprets any graph definition
- **Code generation**: Generate Temporal workflow code from graph, deploy dynamically (complex)

#### 3. Your Domain-Specific Action Configs

Your `action_config` JSONB with template variables still needs parsing:

```json
{
  "action_type": "send_email",
  "action_config": {
    "to": "{{entity.customer_email}}",
    "subject": "Order {{entity.order_number}} confirmed",
    "template": "order_confirmation"
  }
}
```

Temporal activities would still call your existing handlers. The template resolution, entity fetching, and config parsing remain your code.

### Architecture Options

#### Option A: Build Custom Continuation (This Document)

```
┌─────────────────────────────────────────────────────────────────┐
│                        Your Codebase                            │
├─────────────────────────────────────────────────────────────────┤
│  Visual Editor → action_edges → ActionExecutor → Handlers      │
│       │              │               │              │           │
│       │              │         execution_states     │           │
│       │              │               │              │           │
│       └──────────────┴───────────────┴──────────────┘           │
│                    Single Codebase                              │
└─────────────────────────────────────────────────────────────────┘
```

**Pros**:
- Tight integration - graph model IS the execution model
- No new infrastructure
- No network hops between systems
- Full control over behavior

**Cons**:
- You own durability, recovery, timeout, versioning
- More code to write and maintain
- Edge cases (crashes mid-execution, schema migrations with in-flight workflows)

#### Option B: Temporal as Execution Backend

```
┌─────────────────────────────────────────────────────────────────┐
│                        Your Codebase                            │
├─────────────────────────────────────────────────────────────────┤
│  Visual Editor → action_edges → Postgres                       │
│                                    │                            │
│                              Graph Config                       │
└────────────────────────────────────┼────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Temporal Server                            │
├─────────────────────────────────────────────────────────────────┤
│  GraphInterpreterWorkflow ──► Activities (your handlers)       │
│         │                            │                          │
│    Workflow State              Activity Results                 │
│    (Temporal owns)             (passed back)                    │
└─────────────────────────────────────────────────────────────────┘
```

**Pros**:
- Durability, retries, timeouts, visibility handled
- Battle-tested at scale (Uber, Netflix, etc.)
- Built-in dashboard for execution inspection

**Cons**:
- Two systems to operate (Postgres + Temporal cluster)
- Network latency between your code and Temporal
- Still writing graph interpreter logic
- State split across two systems (config in Postgres, execution in Temporal)
- Operational complexity (Temporal requires Cassandra/MySQL + Elasticsearch)

#### Option C: Temporal with Generic Workflow Dispatch (Recommended if using Temporal)

```go
// One workflow that interprets ANY graph definition
func ExecuteConfiguredWorkflow(ctx workflow.Context, input WorkflowInput) error {
    // input.GraphDefinition contains action_edges from Postgres
    // input.TriggerData contains the event that started this

    executor := NewGraphExecutor(input.GraphDefinition)
    context := NewMergedContext(input.TriggerData)

    for {
        nextActions := executor.GetNextActions(context)
        if len(nextActions) == 0 {
            break // Workflow complete
        }

        if len(nextActions) > 1 {
            // Parallel execution
            results := executeParallel(ctx, nextActions, context)
            context.MergeResults(results)
        } else {
            // Sequential
            result := workflow.ExecuteActivity(ctx,
                ExecuteAction,
                nextActions[0],
                context,
            )
            context.MergeResult(nextActions[0].ID, result)
        }

        executor.MarkComplete(nextActions)
    }

    return nil
}
```

**What Temporal handles**:
- `execution_states` table → Temporal's internal state
- Crash recovery → Temporal's replay
- Timeout handling → Temporal's timeout configuration
- Correlation keys → Temporal's workflow ID

**What you still build**:
- `GraphExecutor` - traverses your action_edges
- `MergedContext` - accumulates action results
- `ExecuteAction` activity - calls your existing handlers
- Parallel branch tracking logic (though Temporal helps with `workflow.Go()`)

### Critical Gap: Workflow Versioning During Deployment

**Scenario**: 500 workflows are awaiting async completion (e.g., manager approval). You deploy new code that changes the workflow logic.

**This document's approach**: Not addressed. Options:
1. Let in-flight workflows use old logic (how? old code is gone)
2. Force all to use new logic (may break mid-execution state)
3. Fail all in-flight workflows (data loss)

**Temporal's approach**: Deterministic replay with version markers
```go
func MyWorkflow(ctx workflow.Context) error {
    v := workflow.GetVersion(ctx, "new-branch-logic", workflow.DefaultVersion, 1)
    if v == workflow.DefaultVersion {
        // Old logic for in-flight workflows
    } else {
        // New logic for new workflows
    }
}
```

**This is the strongest argument for Temporal** if you expect frequent workflow definition changes during active executions.

### Operational Complexity Comparison

| Aspect | Custom (This Doc) | Temporal |
|--------|-------------------|----------|
| **Infrastructure** | Postgres (existing) | Temporal server + persistence (Cassandra/MySQL) + Elasticsearch |
| **Deployment** | Single binary | Temporal cluster + worker processes |
| **Monitoring** | Build dashboards | Temporal Web UI (included) |
| **Debugging** | Query execution_states | Temporal history inspection |
| **Scaling** | Scale your service | Scale Temporal + scale workers |
| **Expertise required** | Your team's Go skills | Temporal-specific knowledge |

### Cost-Benefit Analysis

**Build custom (this doc) when**:
- Most workflows complete quickly (seconds to minutes)
- Async operations are internal (inventory allocation, not third-party APIs with days latency)
- Workflow definitions rarely change during active executions
- You want minimal infrastructure
- You have capacity to handle the edge cases

**Use Temporal when**:
- Human-in-the-loop workflows that pause for days
- Frequent workflow definition updates during active executions
- Need for built-in visibility/debugging
- Complex retry/compensation requirements
- Willing to operate Temporal infrastructure

### Hybrid Approach: Deferred Temporal Integration

**Phase 1**: Build custom continuation (this doc)
- Solve the immediate problem (conditions evaluating before async completes)
- Prove the model works
- Keep infrastructure simple

**Phase 2**: Evaluate pain points
- Are workflows failing mid-execution during deployments?
- Is debugging execution state painful?
- Are timeout edge cases accumulating?

**Phase 3**: Migrate to Temporal if needed
- Your action handlers become Temporal activities (minimal change)
- Graph interpreter becomes a Temporal workflow
- `execution_states` table becomes historical (Temporal owns state)

This approach de-risks both paths: you're not locked into custom forever, but you're not paying Temporal's operational cost until you need it.

### Temporal Integration: Technical Sketch

If you decide to integrate Temporal, here's how it would map:

```go
// Worker registration (in main.go or separate worker process)
func main() {
    c, _ := client.NewClient(client.Options{})
    w := worker.New(c, "workflow-task-queue", worker.Options{})

    // Register the generic interpreter workflow
    w.RegisterWorkflow(workflows.ExecuteConfiguredWorkflow)

    // Register activities (wrap your existing handlers)
    w.RegisterActivity(activities.AllocateInventory)  // wraps allocate.go
    w.RegisterActivity(activities.SendEmail)          // wraps email.go
    w.RegisterActivity(activities.CreateAlert)        // wraps alert.go
    w.RegisterActivity(activities.UpdateField)        // wraps updatefield.go
    w.RegisterActivity(activities.EvaluateCondition)  // wraps condition.go

    w.Run(worker.InterruptCh())
}

// Trigger integration (in your event handler)
func (e *EventHandler) OnEntityChange(ctx context.Context, event EntityEvent) {
    // Existing: find matching rules
    rules := e.findMatchingRules(event)

    for _, rule := range rules {
        // Load graph definition
        graph := e.loadGraphDefinition(rule.ID)

        // Start Temporal workflow instead of executing directly
        _, err := e.temporalClient.ExecuteWorkflow(ctx,
            client.StartWorkflowOptions{
                ID:        fmt.Sprintf("workflow-%s-%s", rule.ID, event.EntityID),
                TaskQueue: "workflow-task-queue",
            },
            workflows.ExecuteConfiguredWorkflow,
            WorkflowInput{
                RuleID:          rule.ID,
                GraphDefinition: graph,
                TriggerData:     event.Data,
            },
        )
    }
}
```

**Activity wrapper example**:
```go
// activities/allocate.go
func AllocateInventory(ctx context.Context, config json.RawMessage, execCtx ActionContext) (*AllocationResult, error) {
    // Get your existing handler
    handler := inventory.NewAllocateInventoryHandler(...)

    // Call it
    result, err := handler.Execute(ctx, config, execCtx)
    if err != nil {
        return nil, err
    }

    // Return result for context merging
    return result.(*AllocationResult), nil
}
```

### Recommendation

**For your situation** (tightly coupled with value proposition, configurability is key, complexity jump is acceptable):

1. **Start with custom continuation (this doc)** for these reasons:
   - Direct integration with your visual editor model
   - No additional infrastructure
   - Faster to implement initial version
   - You control all behavior

2. **Add explicit versioning strategy** (the gap this doc doesn't address):
   - Store `workflow_definition_version` with execution state
   - On resume, detect if definition changed
   - Either: replay with original definition (store it) or fail gracefully

3. **Instrument for Temporal migration signals**:
   - Track: execution failures during deployment
   - Track: orphaned states count
   - Track: time spent debugging execution issues
   - If these metrics spike, revisit Temporal

4. **Keep action handlers clean** so they can become Temporal activities later:
   - Stateless
   - Idempotent where possible
   - Clear input/output contracts

---

# Phase 2: Temporal Integration Implementation

This section provides a complete Temporal implementation that respects all graphing constraints from Phase 1. Use this as a migration path if the custom continuation approach proves insufficient.

## When to Migrate to Temporal

Monitor these signals after implementing Phase 1:

| Signal | Threshold | Action |
|--------|-----------|--------|
| Orphaned execution states | >1% of total | Investigate recovery gaps |
| Deployment failures affecting in-flight workflows | Any occurrence | Consider Temporal versioning |
| Time debugging execution issues | >4 hours/week | Temporal visibility helps |
| Human-in-the-loop workflows taking >24 hours | Common pattern | Temporal excels here |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Ichor Codebase                           │
├─────────────────────────────────────────────────────────────────┤
│  Visual Editor → action_edges → Postgres                        │
│                                    │                            │
│                              Graph Config                       │
└────────────────────────────────────┼────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Temporal Server                            │
├─────────────────────────────────────────────────────────────────┤
│  ExecuteGraphWorkflow ──► Activities (your existing handlers)  │
│         │                            │                          │
│    Workflow State              Activity Results                 │
│    (Temporal owns)             (passed back)                    │
└─────────────────────────────────────────────────────────────────┘
```

**What Temporal Handles:**
- `execution_states` table → Temporal's internal state
- Crash recovery → Temporal's deterministic replay
- Timeout handling → Temporal's timeout configuration
- Correlation keys → Temporal's workflow ID

**What You Still Build:**
- `GraphExecutor` - traverses your action_edges
- `MergedContext` - accumulates action results
- `ExecuteActionActivity` - calls your existing handlers
- Parallel branch convergence detection

---

## Core Implementation

### 1. Workflow Input & Context Models

**File**: `business/sdk/workflow/temporal/models.go`

```go
package temporal

import (
    "encoding/json"

    "github.com/google/uuid"
)

// WorkflowInput is passed when starting a workflow execution
type WorkflowInput struct {
    RuleID       uuid.UUID              `json:"rule_id"`
    ExecutionID  uuid.UUID              `json:"execution_id"`
    Graph        GraphDefinition        `json:"graph"`
    TriggerData  map[string]interface{} `json:"trigger_data"`
}

// GraphDefinition mirrors your database model (action_edges + rule_actions)
type GraphDefinition struct {
    Actions []ActionNode `json:"actions"`
    Edges   []ActionEdge `json:"edges"`
}

// ActionNode represents a single action in the workflow graph
type ActionNode struct {
    ID         uuid.UUID       `json:"id"`
    Name       string          `json:"name"`
    ActionType string          `json:"action_type"`
    Config     json.RawMessage `json:"action_config"`
}

// ActionEdge represents a directed edge between actions
type ActionEdge struct {
    ID             uuid.UUID  `json:"id"`
    SourceActionID *uuid.UUID `json:"source_action_id"` // nil for start edges
    TargetActionID uuid.UUID  `json:"target_action_id"`
    EdgeType       string     `json:"edge_type"` // start, sequence, true_branch, false_branch, always, parallel
    SortOrder      int        `json:"sort_order"`
}

// MergedContext accumulates results from all executed actions
// This is the key data structure for template variable resolution
type MergedContext struct {
    TriggerData   map[string]interface{}            `json:"trigger_data"`
    ActionResults map[string]map[string]interface{} `json:"action_results"` // action_name -> result
    Flattened     map[string]interface{}            `json:"flattened"`      // For template resolution
}

// NewMergedContext creates a context initialized with trigger data
func NewMergedContext(triggerData map[string]interface{}) *MergedContext {
    ctx := &MergedContext{
        TriggerData:   triggerData,
        ActionResults: make(map[string]map[string]interface{}),
        Flattened:     make(map[string]interface{}),
    }

    // Copy trigger data to flattened for initial template resolution
    for k, v := range triggerData {
        ctx.Flattened[k] = v
    }

    return ctx
}

// MergeResult adds an action's result to the context
// Supports template access patterns:
//   - {{action_name}} -> entire result map
//   - {{action_name.field}} -> specific field
func (c *MergedContext) MergeResult(actionName string, result map[string]interface{}) {
    if c.ActionResults == nil {
        c.ActionResults = make(map[string]map[string]interface{})
    }
    if c.Flattened == nil {
        c.Flattened = make(map[string]interface{})
    }

    // Store full result by action name
    c.ActionResults[actionName] = result

    // Flatten for template access: "action_name.field" -> value
    for k, v := range result {
        c.Flattened[actionName+"."+k] = v
    }

    // Also store action name pointing to full result for {{action_name}} access
    c.Flattened[actionName] = result
}

// Clone creates a deep copy for parallel branch execution
func (c *MergedContext) Clone() *MergedContext {
    clone := &MergedContext{
        TriggerData:   make(map[string]interface{}),
        ActionResults: make(map[string]map[string]interface{}),
        Flattened:     make(map[string]interface{}),
    }

    for k, v := range c.TriggerData {
        clone.TriggerData[k] = v
    }
    for k, v := range c.ActionResults {
        resultCopy := make(map[string]interface{})
        for rk, rv := range v {
            resultCopy[rk] = rv
        }
        clone.ActionResults[k] = resultCopy
    }
    for k, v := range c.Flattened {
        clone.Flattened[k] = v
    }

    return clone
}

// BranchInput is passed to child workflows for parallel branch execution
type BranchInput struct {
    StartAction      ActionNode      `json:"start_action"`
    ConvergencePoint uuid.UUID       `json:"convergence_point"`
    Graph            GraphDefinition `json:"graph"`
    InitialContext   *MergedContext  `json:"initial_context"`
}

// BranchOutput is returned from child workflows
type BranchOutput struct {
    ActionResults map[string]map[string]interface{} `json:"action_results"`
}

// ActionActivityInput is passed to the action execution activity
type ActionActivityInput struct {
    ActionID   uuid.UUID              `json:"action_id"`
    ActionName string                 `json:"action_name"`
    ActionType string                 `json:"action_type"`
    Config     json.RawMessage        `json:"config"`
    Context    map[string]interface{} `json:"context"` // Merged context for template resolution
}

// ActionActivityOutput is returned from the action execution activity
type ActionActivityOutput struct {
    ActionID   uuid.UUID              `json:"action_id"`
    ActionName string                 `json:"action_name"`
    Result     map[string]interface{} `json:"result"`
    Success    bool                   `json:"success"`
}
```

### 2. Graph Executor (Traverses Your action_edges)

**File**: `business/sdk/workflow/temporal/graph_executor.go`

```go
package temporal

import (
    "sort"

    "github.com/google/uuid"
)

// GraphExecutor traverses the workflow graph respecting edge types
type GraphExecutor struct {
    graph         GraphDefinition
    actionsByID   map[uuid.UUID]ActionNode
    edgesBySource map[uuid.UUID][]ActionEdge // source_action_id -> outgoing edges
    incomingCount map[uuid.UUID]int          // For convergence detection
}

// NewGraphExecutor creates an executor from a graph definition
func NewGraphExecutor(graph GraphDefinition) *GraphExecutor {
    e := &GraphExecutor{
        graph:         graph,
        actionsByID:   make(map[uuid.UUID]ActionNode),
        edgesBySource: make(map[uuid.UUID][]ActionEdge),
        incomingCount: make(map[uuid.UUID]int),
    }

    // Index actions by ID
    for _, action := range graph.Actions {
        e.actionsByID[action.ID] = action
    }

    // Index edges by source and count incoming edges
    for _, edge := range graph.Edges {
        if edge.SourceActionID != nil {
            e.edgesBySource[*edge.SourceActionID] = append(e.edgesBySource[*edge.SourceActionID], edge)
        } else {
            // Start edges (source_action_id = nil) stored under nil UUID
            e.edgesBySource[uuid.Nil] = append(e.edgesBySource[uuid.Nil], edge)
        }
        e.incomingCount[edge.TargetActionID]++
    }

    // Sort edges by sort_order for deterministic execution
    for sourceID := range e.edgesBySource {
        edges := e.edgesBySource[sourceID]
        sort.Slice(edges, func(i, j int) bool {
            return edges[i].SortOrder < edges[j].SortOrder
        })
        e.edgesBySource[sourceID] = edges
    }

    return e
}

// GetStartActions returns all actions with start edges (source_action_id = nil)
func (e *GraphExecutor) GetStartActions() []ActionNode {
    startEdges := e.edgesBySource[uuid.Nil]
    actions := make([]ActionNode, 0, len(startEdges))

    for _, edge := range startEdges {
        if edge.EdgeType == "start" {
            if action, ok := e.actionsByID[edge.TargetActionID]; ok {
                actions = append(actions, action)
            }
        }
    }

    return actions
}

// GetNextActions returns the next actions to execute based on edge types and action result
//
// Edge type behavior:
//   - "sequence": Always follow
//   - "parallel": Always follow (concurrent with other parallel edges)
//   - "true_branch": Follow if result["branch_taken"] == "true_branch"
//   - "false_branch": Follow if result["branch_taken"] == "false_branch"
//   - "always": Always follow (parallel to condition branches)
func (e *GraphExecutor) GetNextActions(sourceActionID uuid.UUID, result map[string]interface{}) []ActionNode {
    edges := e.edgesBySource[sourceActionID]
    if len(edges) == 0 {
        return nil // End of path
    }

    var nextActions []ActionNode

    // Determine which edges to follow based on edge type and result
    for _, edge := range edges {
        shouldFollow := false

        switch edge.EdgeType {
        case "sequence", "parallel":
            // Always follow sequential and parallel edges
            shouldFollow = true

        case "true_branch":
            // Follow if condition evaluated to true
            if branch, ok := result["branch_taken"].(string); ok && branch == "true_branch" {
                shouldFollow = true
            }

        case "false_branch":
            // Follow if condition evaluated to false
            if branch, ok := result["branch_taken"].(string); ok && branch == "false_branch" {
                shouldFollow = true
            }

        case "always":
            // Always follow (used for actions that run regardless of condition result)
            shouldFollow = true

        case "start":
            // Start edges are only used for initial dispatch
            continue
        }

        if shouldFollow {
            if action, ok := e.actionsByID[edge.TargetActionID]; ok {
                nextActions = append(nextActions, action)
            }
        }
    }

    return nextActions
}

// FindConvergencePoint detects if multiple branches lead to the same node
// Returns the convergence node if found, nil if branches are fire-and-forget
func (e *GraphExecutor) FindConvergencePoint(branches []ActionNode) *ActionNode {
    if len(branches) <= 1 {
        return nil
    }

    // Track which nodes each branch can reach
    reachable := make(map[uuid.UUID]int) // node_id -> count of branches that can reach it

    for _, branch := range branches {
        visited := e.findReachableNodes(branch.ID, make(map[uuid.UUID]bool))
        for nodeID := range visited {
            reachable[nodeID]++
        }
    }

    // Find first node reachable by ALL branches (ordered by graph structure)
    // We want the closest convergence point
    var convergencePoint *ActionNode
    var minDepth int = -1

    for nodeID, count := range reachable {
        if count == len(branches) {
            // This node is reachable by all branches
            depth := e.calculateMinDepth(branches[0].ID, nodeID)
            if minDepth == -1 || depth < minDepth {
                minDepth = depth
                action := e.actionsByID[nodeID]
                convergencePoint = &action
            }
        }
    }

    return convergencePoint
}

// findReachableNodes performs BFS to find all nodes reachable from start
func (e *GraphExecutor) findReachableNodes(startID uuid.UUID, visited map[uuid.UUID]bool) map[uuid.UUID]bool {
    if visited[startID] {
        return visited
    }
    visited[startID] = true

    // Follow all outgoing edges (assume condition results don't matter for reachability analysis)
    for _, edge := range e.edgesBySource[startID] {
        e.findReachableNodes(edge.TargetActionID, visited)
    }

    return visited
}

// calculateMinDepth calculates minimum edge count from source to target
func (e *GraphExecutor) calculateMinDepth(sourceID, targetID uuid.UUID) int {
    if sourceID == targetID {
        return 0
    }

    visited := make(map[uuid.UUID]bool)
    queue := []struct {
        id    uuid.UUID
        depth int
    }{{sourceID, 0}}

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        if visited[current.id] {
            continue
        }
        visited[current.id] = true

        for _, edge := range e.edgesBySource[current.id] {
            if edge.TargetActionID == targetID {
                return current.depth + 1
            }
            queue = append(queue, struct {
                id    uuid.UUID
                depth int
            }{edge.TargetActionID, current.depth + 1})
        }
    }

    return -1 // Not reachable
}

// PathLeadsTo checks if starting from actionID eventually reaches targetID
func (e *GraphExecutor) PathLeadsTo(actionID, targetID uuid.UUID) bool {
    reachable := e.findReachableNodes(actionID, make(map[uuid.UUID]bool))
    return reachable[targetID]
}

// Graph returns the underlying graph definition
func (e *GraphExecutor) Graph() GraphDefinition {
    return e.graph
}

// GetAction retrieves an action by ID
func (e *GraphExecutor) GetAction(id uuid.UUID) (ActionNode, bool) {
    action, ok := e.actionsByID[id]
    return action, ok
}

// HasMultipleIncoming returns true if the node has more than one incoming edge
// This indicates a potential convergence point
func (e *GraphExecutor) HasMultipleIncoming(actionID uuid.UUID) bool {
    return e.incomingCount[actionID] > 1
}
```

### 3. Main Workflow (Graph Interpreter)

**File**: `business/sdk/workflow/temporal/workflow.go`

```go
package temporal

import (
    "fmt"
    "time"

    "github.com/google/uuid"
    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"
)

const (
    TaskQueue = "ichor-workflow-queue"
)

// ExecuteGraphWorkflow interprets any graph definition dynamically
// This is the core workflow that respects all graphing constraints
func ExecuteGraphWorkflow(ctx workflow.Context, input WorkflowInput) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting graph workflow",
        "rule_id", input.RuleID,
        "execution_id", input.ExecutionID,
    )

    // Initialize execution context
    mergedCtx := NewMergedContext(input.TriggerData)

    // Build graph executor
    executor := NewGraphExecutor(input.Graph)

    // Find start actions (edges with source_action_id = nil)
    startActions := executor.GetStartActions()
    if len(startActions) == 0 {
        logger.Info("Empty workflow - no start actions found")
        return nil
    }

    // Execute from start (may be multiple parallel start actions)
    return executeActions(ctx, executor, startActions, mergedCtx)
}

// executeActions handles both sequential and parallel execution
func executeActions(ctx workflow.Context, executor *GraphExecutor, actions []ActionNode, mergedCtx *MergedContext) error {
    if len(actions) == 0 {
        return nil
    }

    if len(actions) == 1 {
        // Sequential execution
        return executeSingleAction(ctx, executor, actions[0], mergedCtx)
    }

    // Parallel execution - check for convergence
    convergencePoint := executor.FindConvergencePoint(actions)

    if convergencePoint == nil {
        // Fire-and-forget parallel branches - no convergence
        return executeFireAndForget(ctx, executor, actions, mergedCtx)
    }

    // Parallel with convergence - must wait for all branches
    return executeParallelWithConvergence(ctx, executor, actions, convergencePoint, mergedCtx)
}

// executeSingleAction executes one action and continues to next
func executeSingleAction(ctx workflow.Context, executor *GraphExecutor, action ActionNode, mergedCtx *MergedContext) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Executing action",
        "action_id", action.ID,
        "action_name", action.Name,
        "action_type", action.ActionType,
    )

    // Prepare activity input
    activityInput := ActionActivityInput{
        ActionID:   action.ID,
        ActionName: action.Name,
        ActionType: action.ActionType,
        Config:     action.Config,
        Context:    mergedCtx.Flattened,
    }

    // Configure activity options based on action type
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    3,
        },
    }

    // Async actions get longer timeouts
    if isAsyncAction(action.ActionType) {
        ao.StartToCloseTimeout = 30 * time.Minute
        ao.HeartbeatTimeout = time.Minute
    }

    // Human-in-the-loop actions can take days
    if isHumanAction(action.ActionType) {
        ao.StartToCloseTimeout = 7 * 24 * time.Hour // 7 days
        ao.HeartbeatTimeout = time.Hour
    }

    activityCtx := workflow.WithActivityOptions(ctx, ao)

    // Execute the action as a Temporal activity
    var result ActionActivityOutput
    err := workflow.ExecuteActivity(activityCtx, ExecuteActionActivity, activityInput).Get(ctx, &result)
    if err != nil {
        logger.Error("Action failed",
            "action_id", action.ID,
            "action_name", action.Name,
            "error", err,
        )
        return err
    }

    // Merge result into context for subsequent actions
    mergedCtx.MergeResult(action.Name, result.Result)

    logger.Info("Action completed",
        "action_id", action.ID,
        "action_name", action.Name,
        "success", result.Success,
    )

    // Get next actions based on result and edge types
    nextActions := executor.GetNextActions(action.ID, result.Result)

    if len(nextActions) == 0 {
        // End of this path
        return nil
    }

    // Continue execution
    return executeActions(ctx, executor, nextActions, mergedCtx)
}

// executeFireAndForget launches parallel branches without waiting
func executeFireAndForget(ctx workflow.Context, executor *GraphExecutor, branches []ActionNode, mergedCtx *MergedContext) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Executing fire-and-forget parallel branches",
        "branch_count", len(branches),
    )

    // Launch each branch concurrently using workflow.Go
    for i, branch := range branches {
        // Capture for closure
        branchAction := branch
        branchIndex := i

        workflow.Go(ctx, func(gCtx workflow.Context) {
            // Clone context for this branch
            branchCtx := mergedCtx.Clone()

            logger.Info("Starting fire-and-forget branch",
                "branch_index", branchIndex,
                "start_action", branchAction.Name,
            )

            // Execute this branch (errors logged but don't fail parent)
            if err := executeSingleAction(gCtx, executor, branchAction, branchCtx); err != nil {
                logger.Error("Fire-and-forget branch failed",
                    "branch_index", branchIndex,
                    "error", err,
                )
            }
        })
    }

    // Don't wait for fire-and-forget branches
    return nil
}

// executeParallelWithConvergence executes branches and waits for all at convergence
func executeParallelWithConvergence(
    ctx workflow.Context,
    executor *GraphExecutor,
    branches []ActionNode,
    convergencePoint *ActionNode,
    mergedCtx *MergedContext,
) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Executing parallel branches with convergence",
        "branch_count", len(branches),
        "convergence_point", convergencePoint.Name,
    )

    // Create selector for waiting on all branches
    selector := workflow.NewSelector(ctx)
    branchResults := make([]BranchOutput, len(branches))
    branchErrors := make([]error, len(branches))

    for i, branch := range branches {
        // Capture for closure
        branchIndex := i
        branchAction := branch

        // Execute each branch as a child workflow
        childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
            WorkflowID: fmt.Sprintf("%s-branch-%d-%s",
                workflow.GetInfo(ctx).WorkflowExecution.ID,
                branchIndex,
                branchAction.ID,
            ),
        })

        future := workflow.ExecuteChildWorkflow(childCtx, ExecuteBranchUntilConvergence,
            BranchInput{
                StartAction:      branchAction,
                ConvergencePoint: convergencePoint.ID,
                Graph:            executor.Graph(),
                InitialContext:   mergedCtx,
            },
        )

        selector.AddFuture(future, func(f workflow.Future) {
            var output BranchOutput
            branchErrors[branchIndex] = f.Get(ctx, &output)
            branchResults[branchIndex] = output
        })
    }

    // Wait for all branches
    for i := 0; i < len(branches); i++ {
        selector.Select(ctx)
    }

    // Check for errors
    for i, err := range branchErrors {
        if err != nil {
            logger.Error("Branch failed",
                "branch_index", i,
                "error", err,
            )
            return fmt.Errorf("branch %d failed: %w", i, err)
        }
    }

    // Merge all branch results into context
    for _, br := range branchResults {
        for actionName, actionResult := range br.ActionResults {
            mergedCtx.MergeResult(actionName, actionResult)
        }
    }

    logger.Info("All branches converged, continuing from convergence point",
        "convergence_point", convergencePoint.Name,
    )

    // Continue from convergence point
    return executeSingleAction(ctx, executor, *convergencePoint, mergedCtx)
}

// ExecuteBranchUntilConvergence executes actions until reaching the convergence point
// This is a child workflow for each parallel branch
func ExecuteBranchUntilConvergence(ctx workflow.Context, input BranchInput) (BranchOutput, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting branch execution",
        "start_action", input.StartAction.Name,
        "convergence_point", input.ConvergencePoint,
    )

    executor := NewGraphExecutor(input.Graph)
    mergedCtx := input.InitialContext.Clone()

    currentAction := input.StartAction

    for {
        // Check if we've reached convergence point
        if currentAction.ID == input.ConvergencePoint {
            logger.Info("Branch reached convergence point")
            break
        }

        // Execute action
        activityInput := ActionActivityInput{
            ActionID:   currentAction.ID,
            ActionName: currentAction.Name,
            ActionType: currentAction.ActionType,
            Config:     currentAction.Config,
            Context:    mergedCtx.Flattened,
        }

        ao := workflow.ActivityOptions{
            StartToCloseTimeout: 5 * time.Minute,
            RetryPolicy: &temporal.RetryPolicy{
                InitialInterval:    time.Second,
                BackoffCoefficient: 2.0,
                MaximumAttempts:    3,
            },
        }

        if isAsyncAction(currentAction.ActionType) {
            ao.StartToCloseTimeout = 30 * time.Minute
        }

        activityCtx := workflow.WithActivityOptions(ctx, ao)

        var result ActionActivityOutput
        err := workflow.ExecuteActivity(activityCtx, ExecuteActionActivity, activityInput).Get(ctx, &result)
        if err != nil {
            return BranchOutput{}, err
        }

        mergedCtx.MergeResult(currentAction.Name, result.Result)

        // Get next action (should be single path within a branch before convergence)
        nextActions := executor.GetNextActions(currentAction.ID, result.Result)
        if len(nextActions) == 0 {
            // Dead end before convergence - this shouldn't happen in well-formed graphs
            break
        }

        // Take first next action (branches within branches would need nested handling)
        currentAction = nextActions[0]
    }

    return BranchOutput{
        ActionResults: mergedCtx.ActionResults,
    }, nil
}

// isAsyncAction returns true for actions that involve external async operations
func isAsyncAction(actionType string) bool {
    asyncTypes := map[string]bool{
        "allocate_inventory":   true,
        "send_email":           true,
        "credit_check":         true,
        "fraud_detection":      true,
        "third_party_api_call": true,
        "reserve_shipping":     true,
    }
    return asyncTypes[actionType]
}

// isHumanAction returns true for actions that require human interaction
func isHumanAction(actionType string) bool {
    humanTypes := map[string]bool{
        "manager_approval":       true,
        "manual_review":          true,
        "human_verification":     true,
        "approval_request":       true,
    }
    return humanTypes[actionType]
}
```

### 4. Activities (Wrap Existing Handlers)

**File**: `business/sdk/workflow/temporal/activities.go`

```go
package temporal

import (
    "context"
    "encoding/json"
    "fmt"

    "go.temporal.io/sdk/activity"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ActivityDependencies holds all dependencies needed by activities
type ActivityDependencies struct {
    ActionRegistry *workflow.ActionRegistry
    // Add other dependencies as needed
}

var deps *ActivityDependencies

// SetActivityDependencies initializes activity dependencies (called during worker setup)
func SetActivityDependencies(d *ActivityDependencies) {
    deps = d
}

// ExecuteActionActivity dispatches to the appropriate handler based on action type
// This wraps your existing action handlers as Temporal activities
func ExecuteActionActivity(ctx context.Context, input ActionActivityInput) (ActionActivityOutput, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Executing action activity",
        "action_id", input.ActionID,
        "action_name", input.ActionName,
        "action_type", input.ActionType,
    )

    if deps == nil || deps.ActionRegistry == nil {
        return ActionActivityOutput{}, fmt.Errorf("activity dependencies not initialized")
    }

    // Build execution context matching your existing ActionExecutionContext structure
    execCtx := workflow.ActionExecutionContext{
        ActionID:   input.ActionID.String(),
        ActionName: input.ActionName,
        RawData:    input.Context, // Merged context with all prior results
    }

    // Get handler from registry
    handler, err := deps.ActionRegistry.GetHandler(input.ActionType)
    if err != nil {
        return ActionActivityOutput{
            ActionID:   input.ActionID,
            ActionName: input.ActionName,
            Success:    false,
        }, fmt.Errorf("unknown action type %s: %w", input.ActionType, err)
    }

    // Execute the action using your existing handler
    result, err := handler.Execute(ctx, input.Config, execCtx)
    if err != nil {
        logger.Error("Action execution failed",
            "action_id", input.ActionID,
            "error", err,
        )
        return ActionActivityOutput{
            ActionID:   input.ActionID,
            ActionName: input.ActionName,
            Success:    false,
        }, err
    }

    // Convert result to map for context merging
    resultMap := toResultMap(result)

    logger.Info("Action execution succeeded",
        "action_id", input.ActionID,
        "action_name", input.ActionName,
    )

    return ActionActivityOutput{
        ActionID:   input.ActionID,
        ActionName: input.ActionName,
        Result:     resultMap,
        Success:    true,
    }, nil
}

// toResultMap converts any result type to a map for context merging
func toResultMap(result interface{}) map[string]interface{} {
    if result == nil {
        return map[string]interface{}{}
    }

    // If already a map, return directly
    if m, ok := result.(map[string]interface{}); ok {
        return m
    }

    // Marshal/unmarshal for struct types
    data, err := json.Marshal(result)
    if err != nil {
        return map[string]interface{}{"raw": result}
    }

    var m map[string]interface{}
    if err := json.Unmarshal(data, &m); err != nil {
        return map[string]interface{}{"raw": result}
    }

    return m
}
```

### 5. Trigger Integration

**File**: `business/sdk/workflow/temporal/trigger.go`

```go
package temporal

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "go.temporal.io/sdk/client"

    "github.com/timmaaaz/ichor/business/domain/workflow/automationrulesbus"
    "github.com/timmaaaz/ichor/foundation/logger"
)

// EntityEvent represents an entity change event
type EntityEvent struct {
    EntityID   uuid.UUID
    EntityName string
    EventType  string // on_create, on_update, on_delete
    Data       map[string]interface{}
}

// EdgeStore interface for loading graph definitions
type EdgeStore interface {
    QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionNode, error)
    QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]ActionEdge, error)
}

// WorkflowTrigger handles entity events and starts Temporal workflows
type WorkflowTrigger struct {
    log            *logger.Logger
    temporalClient client.Client
    rulesBus       *automationrulesbus.Business
    edgeStore      EdgeStore
}

// NewWorkflowTrigger creates a new trigger handler
func NewWorkflowTrigger(
    log *logger.Logger,
    tc client.Client,
    rb *automationrulesbus.Business,
    es EdgeStore,
) *WorkflowTrigger {
    return &WorkflowTrigger{
        log:            log,
        temporalClient: tc,
        rulesBus:       rb,
        edgeStore:      es,
    }
}

// OnEntityEvent is called when an entity event fires
// This replaces your current event-driven rule matching with Temporal workflow dispatch
func (t *WorkflowTrigger) OnEntityEvent(ctx context.Context, event EntityEvent) error {
    t.log.Info(ctx, "Processing entity event",
        "entity_name", event.EntityName,
        "event_type", event.EventType,
        "entity_id", event.EntityID,
    )

    // Find matching automation rules (your existing logic)
    filter := automationrulesbus.QueryFilter{
        TriggerEntity: &event.EntityName,
        TriggerEvent:  &event.EventType,
        IsActive:      boolPtr(true),
    }

    rules, err := t.rulesBus.Query(ctx, filter, automationrulesbus.DefaultOrderBy, page.MustParse("1", "100"))
    if err != nil {
        return fmt.Errorf("query rules: %w", err)
    }

    t.log.Info(ctx, "Found matching rules",
        "count", len(rules),
    )

    for _, rule := range rules {
        // Load the graph definition from action_edges
        graph, err := t.loadGraphDefinition(ctx, rule.ID)
        if err != nil {
            t.log.Error(ctx, "Failed to load graph definition",
                "rule_id", rule.ID,
                "error", err,
            )
            continue
        }

        // Skip empty workflows
        if len(graph.Actions) == 0 {
            continue
        }

        // Generate unique execution ID
        executionID := uuid.New()

        // Create unique workflow ID
        workflowID := fmt.Sprintf("workflow-%s-%s-%s",
            rule.ID,
            event.EntityID,
            executionID,
        )

        // Start Temporal workflow
        workflowOptions := client.StartWorkflowOptions{
            ID:        workflowID,
            TaskQueue: TaskQueue,
        }

        we, err := t.temporalClient.ExecuteWorkflow(ctx, workflowOptions,
            ExecuteGraphWorkflow,
            WorkflowInput{
                RuleID:      rule.ID,
                ExecutionID: executionID,
                Graph:       graph,
                TriggerData: event.Data,
            },
        )

        if err != nil {
            t.log.Error(ctx, "Failed to start workflow",
                "rule_id", rule.ID,
                "workflow_id", workflowID,
                "error", err,
            )
            continue
        }

        t.log.Info(ctx, "Started workflow",
            "rule_id", rule.ID,
            "workflow_id", workflowID,
            "run_id", we.GetRunID(),
        )
    }

    return nil
}

func (t *WorkflowTrigger) loadGraphDefinition(ctx context.Context, ruleID uuid.UUID) (GraphDefinition, error) {
    // Load actions for this rule
    actions, err := t.edgeStore.QueryActionsByRule(ctx, ruleID)
    if err != nil {
        return GraphDefinition{}, fmt.Errorf("query actions: %w", err)
    }

    // Load edges for this rule
    edges, err := t.edgeStore.QueryEdgesByRule(ctx, ruleID)
    if err != nil {
        return GraphDefinition{}, fmt.Errorf("query edges: %w", err)
    }

    return GraphDefinition{
        Actions: actions,
        Edges:   edges,
    }, nil
}

func boolPtr(b bool) *bool {
    return &b
}
```

### 6. Worker Setup

**File**: `api/cmd/services/workflow-worker/main.go`

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/ardanlabs/conf/v3"
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"

    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
    "github.com/timmaaaz/ichor/foundation/logger"
)

func main() {
    log := logger.New(os.Stdout, logger.LevelInfo, "WORKFLOW-WORKER", func(context.Context) string { return "" })

    if err := run(log); err != nil {
        log.Error(context.Background(), "startup", "error", err)
        os.Exit(1)
    }
}

func run(log *logger.Logger) error {
    // Configuration
    cfg := struct {
        Temporal struct {
            HostPort  string `conf:"default:localhost:7233"`
            Namespace string `conf:"default:default"`
        }
        DB struct {
            User     string `conf:"default:postgres"`
            Password string `conf:"default:postgres,mask"`
            Host     string `conf:"default:localhost:5432"`
            Name     string `conf:"default:postgres"`
        }
    }{}

    if _, err := conf.Parse("ICHOR", &cfg); err != nil {
        return fmt.Errorf("parsing config: %w", err)
    }

    // Create Temporal client
    c, err := client.Dial(client.Options{
        HostPort:  cfg.Temporal.HostPort,
        Namespace: cfg.Temporal.Namespace,
        Logger:    newTemporalLogger(log),
    })
    if err != nil {
        return fmt.Errorf("creating temporal client: %w", err)
    }
    defer c.Close()

    // Initialize database connection and action registry
    // (Similar to your existing main.go setup)
    actionRegistry := workflow.NewActionRegistry()
    // Register all your action handlers...
    // actionRegistry.Register("allocate_inventory", inventory.NewAllocateInventoryHandler(...))
    // actionRegistry.Register("send_email", notification.NewSendEmailHandler(...))
    // etc.

    // Set activity dependencies
    temporal.SetActivityDependencies(&temporal.ActivityDependencies{
        ActionRegistry: actionRegistry,
    })

    // Create worker
    w := worker.New(c, temporal.TaskQueue, worker.Options{
        MaxConcurrentActivityExecutionSize:     100,
        MaxConcurrentWorkflowTaskExecutionSize: 100,
    })

    // Register workflows
    w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
    w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)

    // Register activities
    w.RegisterActivity(temporal.ExecuteActionActivity)

    log.Info(context.Background(), "Starting workflow worker",
        "task_queue", temporal.TaskQueue,
        "temporal_host", cfg.Temporal.HostPort,
    )

    // Handle shutdown
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Run worker
    errCh := make(chan error, 1)
    go func() {
        errCh <- w.Run(worker.InterruptCh())
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        log.Info(context.Background(), "Shutting down worker")
        return nil
    }
}

// temporalLogger adapts our logger to Temporal's logger interface
type temporalLogger struct {
    log *logger.Logger
}

func newTemporalLogger(log *logger.Logger) *temporalLogger {
    return &temporalLogger{log: log}
}

func (l *temporalLogger) Debug(msg string, keyvals ...interface{}) {
    l.log.Debug(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Info(msg string, keyvals ...interface{}) {
    l.log.Info(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Warn(msg string, keyvals ...interface{}) {
    l.log.Warn(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Error(msg string, keyvals ...interface{}) {
    l.log.Error(context.Background(), msg, keyvals...)
}
```

---

## Migration Path from Phase 1 to Phase 2

### Step 1: Deploy Temporal Infrastructure

```bash
# Using Docker Compose for development
docker-compose -f docker-compose-temporal.yml up -d

# Or using Kubernetes (helm)
helm install temporal temporalio/temporal \
  --set cassandra.enabled=true \
  --set elasticsearch.enabled=true
```

### Step 2: Create EdgeStore Adapter

Implement the `EdgeStore` interface to read from your existing tables:

```go
// business/sdk/workflow/temporal/stores/edgedb/edgedb.go

package edgedb

import (
    "context"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

type Store struct {
    db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
    return &Store{db: db}
}

func (s *Store) QueryActionsByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionNode, error) {
    const q = `
        SELECT id, name, action_type, action_config
        FROM workflow.rule_actions
        WHERE rule_id = $1
        ORDER BY sequence_order
    `
    // Implementation...
}

func (s *Store) QueryEdgesByRule(ctx context.Context, ruleID uuid.UUID) ([]temporal.ActionEdge, error) {
    const q = `
        SELECT id, source_action_id, target_action_id, edge_type, sort_order
        FROM workflow.action_edges
        WHERE rule_id = $1
        ORDER BY sort_order
    `
    // Implementation...
}
```

### Step 3: Dual-Write During Migration

Run both systems in parallel during migration:

```go
func (t *Trigger) OnEntityEvent(ctx context.Context, event EntityEvent) error {
    // Feature flag or percentage rollout
    if shouldUseTemporalForRule(rule.ID) {
        return t.temporalTrigger.OnEntityEvent(ctx, event)
    }
    return t.customTrigger.OnEntityEvent(ctx, event)
}
```

### Step 4: Deprecate execution_states Table

Once all workflows use Temporal, the `execution_states` table from Phase 1 is no longer needed. Temporal maintains its own execution history.

---

## Comparison: Phase 1 vs Phase 2

| Aspect | Phase 1 (Custom) | Phase 2 (Temporal) |
|--------|------------------|-------------------|
| **State persistence** | `execution_states` table | Temporal internal |
| **Crash recovery** | Manual orphan detection | Automatic replay |
| **Timeouts** | `timeout_checker` job | Native configuration |
| **Retries** | Manual implementation | Policy-based |
| **Visibility** | Build dashboards | Temporal Web UI |
| **Versioning** | Store definition version | `GetVersion()` API |
| **Infrastructure** | Postgres only | Temporal cluster |
| **Graph execution** | Same `GraphExecutor` | Same `GraphExecutor` |
| **Action handlers** | Direct calls | Wrapped as activities |
| **Context merging** | Same `MergedContext` | Same `MergedContext` |

---

## Files Summary for Phase 2

### New Files
- `business/sdk/workflow/temporal/models.go` - Input/output types
- `business/sdk/workflow/temporal/graph_executor.go` - Graph traversal
- `business/sdk/workflow/temporal/workflow.go` - Main workflow logic
- `business/sdk/workflow/temporal/activities.go` - Activity wrappers
- `business/sdk/workflow/temporal/trigger.go` - Event to workflow dispatch
- `business/sdk/workflow/temporal/stores/edgedb/edgedb.go` - Edge store adapter
- `api/cmd/services/workflow-worker/main.go` - Worker entry point
- `zarf/docker/dockerfile.workflow-worker` - Worker container
- `zarf/k8s/dev/workflow-worker/` - K8s deployment

### Modified Files
- `api/cmd/services/ichor/build/all/all.go` - Initialize Temporal client
- `business/sdk/workflow/engine.go` - Add Temporal dispatch option
- `Makefile` - Add workflow-worker targets

---

## Temporal Infrastructure Requirements

### Development (Docker Compose)

```yaml
# zarf/compose/docker-compose-temporal.yml
version: "3.5"
services:
  temporal:
    image: temporalio/auto-setup:1.22
    ports:
      - "7233:7233"
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=temporal
      - POSTGRES_PWD=temporal
      - POSTGRES_SEEDS=postgres
    depends_on:
      - postgres

  temporal-ui:
    image: temporalio/ui:2.21
    ports:
      - "8080:8080"
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
    depends_on:
      - temporal
```

### Production (Kubernetes)

See [Temporal Helm Charts](https://github.com/temporalio/helm-charts) for production deployment with:
- Cassandra or PostgreSQL persistence
- Elasticsearch for visibility
- Multiple history/matching/frontend shards
- Prometheus metrics

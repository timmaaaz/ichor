# Progress Summary: workflow-engine.md

## Overview
Complete workflow execution engine using Temporal. Implements trigger matching, graph-based action execution, parallel branching, sync/async activities, and history continuation.

## State Machine

```
trigger-event → rule-match → workflow-dispatch → graph-traverse → action-execute
                                                                   ├─ sync: activity returns result
                                                                   └─ async: StartAsync + CompleteActivity
```

### Invariants
- Each rule execution gets unique executionID
- Workflow ID deterministic prefix: "workflow-{ruleID}-{entityID}-{executionID}"

## Pipeline

```
All domain [bus] Create/Update/Delete
  ↓
delegate.Call()
  ↓
DelegateHandler[sdk]
  ↓
WorkflowTrigger[sdk]
  ↓
Temporal[⚡]
  ↓
Worker
  ↓
Activities[sdk]
```

## DelegateHandler [sdk] — `business/sdk/workflow/temporal/delegatehandler.go`

**Responsibility:** Convert domain events to workflow triggers.

### Key Facts
- **Implements delegate.Func** — registered in all.go at startup
- **delegate.Data → workflow.TriggerEvent conversion** via reflection (extractEntityData)
- **All domain creates/updates/deletes fire through here automatically**

## WorkflowTrigger [sdk] — `business/sdk/workflow/temporal/trigger.go`

**Responsibility:** Match domain events to rules and dispatch workflows.

### Key Facts
- **OnEntityEvent(ctx, TriggerEvent)** — entry point from DelegateHandler
- **RuleMatcher.ProcessEvent()** → []MatchedRule
- **Loads GraphDefinition** per matched rule from EdgeStore
- **Fails open** — individual rule failure logged + skipped; other rules proceed
- **Workflow ID format:** "workflow-{ruleID}-{entityID}-{executionID}" (deterministic prefix)
- **Task queue:** "ichor-workflow-queue" (models.go:18); tests: "test-workflow-{t.Name()}"
- **⚡ Temporal.ExecuteWorkflow** — dispatches workflow to Temporal

## TriggerProcessor / RuleMatcher [sdk] — `business/sdk/workflow/trigger.go`

**Responsibility:** Match trigger events against rule conditions.

### TriggerProcessor
- **Initialize(ctx context.Context) error** — loads active rules (NOT LoadRules())
- **ProcessEvent(ctx, TriggerEvent) → MatchResult{MatchedRules[]}**
- **Condition evaluation** — TriggerConditions JSON matched against TriggerEvent.FieldChanges

### RuleMatcher Interface
```go
type RuleMatcher interface {
    ProcessEvent(ctx context.Context, event TriggerEvent) (MatchResult, error)
}
```

Extracted for unit test isolation.

## TemporalWorkflow [sdk] — `business/sdk/workflow/temporal/workflow.go`

**Responsibility:** Execute graph via Temporal workflow.

### Key Facts
- **Receives WorkflowInput** — delegates execution to GraphExecutor.Execute()
- **Continue-As-New triggered** at ~10,000 history events
- **ContinuationState (*MergedContext)** preserves accumulated results across CAN

### WorkflowInput
```go
type WorkflowInput struct {
    RuleID            uuid.UUID
    RuleName          string
    ExecutionID       uuid.UUID
    Graph             GraphDefinition
    TriggerData       map[string]any
    ContinuationState *MergedContext
}

type GraphDefinition struct {
    Actions []ActionNode
    Edges   []ActionEdge
}

type ActionNode struct {
    ID            uuid.UUID
    Name          string
    Description   string
    ActionType    string
    Config        json.RawMessage
    IsActive      bool
    DeactivatedBy uuid.UUID
}

type ActionEdge struct {
    ID             uuid.UUID
    SourceActionID *uuid.UUID        // nil for start edge
    TargetActionID uuid.UUID
    EdgeType       string            // start, sequence, always
    SourceOutput   *string           // for conditional routing
    SortOrder      int
}

type MergedContext struct {
    TriggerData     map[string]any
    ActionResults   map[string]map[string]any  // action_name → result map
    Flattened       map[string]any             // flat {{key}} access
}
```

## GraphExecutor [sdk] — `business/sdk/workflow/temporal/graph_executor.go`

**Responsibility:** Traverse action graph with BFS, handle sync/async activities, manage parallel branches.

### Key Facts
- **BFS traversal** of ActionNode/ActionEdge graph
- **Edge types:** start, sequence, always (only 3 types — no true_branch/false_branch)
- **Parallel branches** — fire concurrent activities, merge at convergence point
- **Fire-and-forget** — BranchInput.ConvergencePoint = uuid.Nil signals abandon on parent close
- **Activity options** — MaxAttempts=3 sync, MaxAttempts=1 async/human (prevents duplicate queue)

### BranchInput
```go
type BranchInput struct {
    StartAction       ActionNode
    ConvergencePoint  uuid.UUID         // uuid.Nil = fire-and-forget
    Graph             GraphDefinition
    InitialContext    *MergedContext
    RuleID            uuid.UUID
    ExecutionID       uuid.UUID
    RuleName          string
}
```

## Activities [sdk] — `business/sdk/workflow/temporal/activities.go`

**Responsibility:** Execute individual workflow actions.

### Key Facts
- **Activities struct** holds Registry (sync) + AsyncRegistry (async)
- **ExecuteActionActivity** — sync handler path
- **ExecuteAsyncActionActivity** — async handler StartAsync path
- **selectActivityFunc** — returns string name ("ExecuteActionActivity" / "ExecuteAsyncActionActivity")
- **toResultMap** — handles nil/map/struct via JSON roundtrip (int64→float64 lossy >2^53)

### ActionActivityInput
```go
type ActionActivityInput struct {
    ActionID    uuid.UUID
    ActionName  string
    ActionType  string
    Config      json.RawMessage
    Context     map[string]any
    RuleID      uuid.UUID
    ExecutionID uuid.UUID
    RuleName    string
}

type ActionActivityOutput struct {
    ActionID   uuid.UUID
    ActionName string
    Result     map[string]any
    Success    bool
}
```

## ActionHandler Interface [sdk] — `business/sdk/workflow/interfaces.go`

**Responsibility:** Define contract for all action implementations.

```go
type ActionHandler interface {
    // Execute performs the action with given configuration and context.
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)

    // Validate validates the action configuration before execution.
    Validate(config json.RawMessage) error

    // GetType returns the unique identifier for this action type.
    GetType() string

    // SupportsManualExecution returns true if this action can be triggered manually.
    SupportsManualExecution() bool

    // IsAsync returns true if this action queues work for async processing.
    IsAsync() bool

    // GetDescription returns a human-readable description for discovery APIs.
    GetDescription() string
}
```

### Registered Action Types (21 Production Handlers)

**Control:** seek_approval, evaluate_condition, delay

**Data:** update_field, create_entity, lookup_entity, transition_status, log_audit_entry

**Inventory:** check_inventory, allocate_inventory, reserve_inventory, release_reservation, check_reorder_point, receive_inventory, commit_allocation, create_put_away_task

**Communication:** send_email, send_notification, create_alert

**Integration:** call_webhook, create_purchase_order

## EdgeStore [db] — `business/sdk/workflow/temporal/stores/edgedb/edgedb.go`

**Responsibility:** Load workflow graph from database.

### Key Facts
- **Read-only** — 2 methods (LoadActions, LoadEdges)
- **Custom query** — NOT rule_actions_view (view lacks deactivated_by column)
- **LEFT JOIN action_templates** — for ActionType resolution (NULL template_id → empty string)
- **sql.NullString** — for nullable UUIDs (deactivated_by, source_action_id)
- **NamedQuerySlice returns nil** — for empty (NOT ErrDBNotFound)

### Data Sources
- ⊗ workflow.rule_actions
- ⊗ workflow.action_edges
- ⊗ workflow.action_templates

## WorkflowBus [bus] — `business/sdk/workflow/workflowbus/`

**Responsibility:** CRUD for workflow metadata (rules, actions, edges, templates, executions).

### Key Facts
- **CRUD for** — AutomationRules, RuleActions, ActionEdges, ActionTemplates, Executions
- **Does NOT handle trigger/dispatch** — that is TriggerProcessor + WorkflowTrigger

### Data Sources
- ⊗⊕ workflow.automation_rules
- ⊗⊕ workflow.rule_actions
- ⊗⊕ workflow.action_edges
- ⊗⊕ workflow.action_templates
- ⊗⊕ workflow.automation_executions

## DBSchema

| Table | Purpose |
|-------|---------|
| workflow.automation_rules | Rule definitions (entity_id, trigger_type, conditions) |
| workflow.rule_actions | Action nodes attached to rules |
| workflow.action_edges | Directed edges (type: start / sequence / always) |
| workflow.action_templates | Reusable action type configs |
| workflow.automation_executions | Execution history |

## Critical Warnings

### ⚠ evaluate_condition action config JSON tags

**File:** `business/sdk/workflow/workflowactions/control/condition.go`

Read config struct json tags before writing any seed/test data.

**Silent fail pattern:** wrong json key → field evaluates to zero value → condition = false, no error logged

### ⚠ alert source_rule_id propagation

**File:** `business/sdk/workflow/workflowactions/communication/alert.go`

- **Execute()** assigns SourceRuleID
- **SourceRuleID = uuid.Nil** for manual executions (nil execCtx.RuleID)
- **Test isolation:** always scope alert queries — alertbus.QueryFilter{SourceRuleID: &ruleID}
- **Never count global alert totals** in workflow tests — concurrent subtests pollute the count

### ⚠ Execute() MUST return map[string]any — never a typed struct

**Why:**
1. Temporal deserializes activity results to map — concrete types erased at SDK boundary
2. MergedContext.ActionResults is map[string]map[string]any — template resolution needs {{action_name.field}}
3. GraphExecutor reads result["output"] for edge routing — must coexist with result data

**Required key "output":** string matched against ActionEdge.SourceOutput
- If missing: activities.go injects "success" default (silent misroute risk)

**Pattern:** typed structs fine internally — only Execute() return must be map

**Testing:** assert to map[string]any, never concrete struct

## Change Patterns

### ⚠ Adding a New ActionHandler

Affects 6 areas:
1. `business/sdk/workflow/interfaces.go` — confirm ActionHandler interface satisfied
2. `business/sdk/workflow/temporal/activities.go` — AsyncRegistry vs Registry decision
3. `api/cmd/services/ichor/build/all/all.go` — Register() call in ActionRegistry setup
4. `business/sdk/dbtest/seedmodels/` — new test seed if handler needs domain data
5. `docs/workflow/README.md` — update handler catalog
6. **Verify:** goToImplementation — confirm 21 existing implementors; register alongside them

### ⚠ Adding a New Edge Type

Affects 5 areas:
1. `business/sdk/workflow/temporal/models.go` — new EdgeType const
2. `business/sdk/workflow/temporal/graph_executor.go` — handle in BFS traversal
3. `app/domain/workflow/workflowsaveapp/graph.go` — cycle/validation logic
4. `app/domain/workflow/workflowsaveapp/model.go` — allowed edge types
5. `api/cmd/services/ichor/tests/workflow/` — integration test update

### ⚠ Changing WorkflowInput Shape

Affects 5 areas:
1. `business/sdk/workflow/temporal/models.go` — WorkflowInput struct
2. `business/sdk/workflow/temporal/workflow.go` — unpack new fields
3. `business/sdk/workflow/temporal/trigger.go` — populate new fields when dispatching
4. `business/sdk/workflow/temporal/graph_executor.go` — consume new fields if needed
5. `apitest/workflow.go` — update test infra if WorkflowInfra changes

## Critical Points
- **Only 3 edge types** — start, sequence, always (not conditional branching)
- **Fail-open rule matching** — individual rule failures don't block other rules
- **Fire-and-forget branches** — uuid.Nil convergence point abandons on parent close
- **Execute() returns map only** — enables template resolution and edge routing
- **Task queue per test** — prevents cross-test activity routing in parallel execution

## Notes for Future Development
The workflow engine is sophisticated with careful Temporal integration. Most changes will be:
- Adding new action types (moderate, requires handler + registry entry + testing)
- Adjusting activity options (low-risk if MaxAttempts logic understood)
- Changing edge types (risky, affects validation + execution engine)

The distinction between sync (MaxAttempts=3) and async (MaxAttempts=1) activities is critical — changing it could cause duplicate queue publishes.

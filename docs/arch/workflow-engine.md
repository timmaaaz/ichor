# workflow-engine

[bus]=business layer [sdk]=SDK [app]=app layer
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction

---

## StateMachine

trigger-event → rule-match → workflow-dispatch → graph-traverse → action-execute
                                                                  ├─ sync: activity returns result
                                                                  └─ async: StartAsync + CompleteActivity
invariant: each rule execution gets unique executionID; workflow ID deterministic prefix

---

## Pipeline

DelegateHandler[sdk] → WorkflowTrigger[sdk] → Temporal[⚡] → Worker → Activities[sdk]
        ↑
all domain [bus] Create/Update/Delete → delegate.Call()

---

## DelegateHandler [sdk]

file: business/sdk/workflow/temporal/delegatehandler.go
imports: delegate.Handler interface, WorkflowTrigger
key facts:
  - Implements delegate.Handler — registered in all.go at startup
  - delegate.Data → workflow.TriggerEvent conversion (extractEntityData via reflection)
  - All domain creates/updates/deletes fire through here automatically

---

## WorkflowTrigger [sdk]

file: business/sdk/workflow/temporal/trigger.go
imports: RuleMatcher[sdk], WorkflowStarter (narrow client.Client interface), EdgeStore[db]
key facts:
  - OnEntityEvent(ctx, TriggerEvent) — entry point from DelegateHandler
  - RuleMatcher.ProcessEvent() → []MatchedRule
  - Loads GraphDefinition per matched rule from EdgeStore
  - Fails open: individual rule failure logged + skipped; other rules proceed
  - Workflow ID: "workflow-{ruleID}-{entityID}-{executionID}" (deterministic prefix)
  - Task queue: "ichor-workflow-queue" (models.go:18) (tests: "test-workflow-{t.Name()}")

⚡ Temporal.ExecuteWorkflow

---

## TriggerProcessor / RuleMatcher [sdk]

file: business/sdk/workflow/temporal/ (TriggerProcessor)
imports: workflow.Business[bus]
key facts:
  - RuleMatcher interface extracted for unit test isolation
  - Initialize() loads active rules (NOT LoadRules())
  - ProcessEvent(ctx, TriggerEvent) → MatchResult{MatchedRules[]}
  - Condition evaluation: TriggerConditions JSON matched against TriggerEvent.FieldChanges

---

## Temporal Workflow [sdk]

file: business/sdk/workflow/temporal/workflow.go
imports: GraphExecutor[sdk], temporal.workflow SDK
key facts:
  - Receives WorkflowInput, delegates execution to GraphExecutor.Execute()
  - Continue-As-New triggered at ~10,000 history events
  - ContinuationState (*MergedContext) preserves accumulated results across CAN

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
    SourceActionID *uuid.UUID
    TargetActionID uuid.UUID
    EdgeType       string
    SourceOutput   *string
    SortOrder      int
}

type MergedContext struct {
    TriggerData   map[string]any
    ActionResults map[string]map[string]any
    Flattened     map[string]any
}
```

---

## GraphExecutor [sdk]

file: business/sdk/workflow/temporal/graph_executor.go
key facts:
  - BFS traversal of ActionNode/ActionEdge graph
  - Edge types: start, sequence, always (only 3 types — no true_branch/false_branch)
  - Parallel branches: fire concurrent activities, merge at convergence point
  - BranchInput.ConvergencePoint = uuid.Nil → fire-and-forget (parent close = abandon)
  - activityOptions(): MaxAttempts=3 sync, MaxAttempts=1 async/human (no duplicate queue)

```go
type BranchInput struct {
    StartAction       ActionNode
    ConvergencePoint  uuid.UUID
    Graph             GraphDefinition
    InitialContext    *MergedContext
    RuleID            uuid.UUID
    ExecutionID       uuid.UUID
    RuleName          string
}
```

---

## Activities [sdk]

file: business/sdk/workflow/temporal/activities.go
imports: workflow.ActionRegistry[sdk], AsyncRegistry[sdk]
key facts:
  - Activities struct holds Registry (sync) + AsyncRegistry (async)
  - ExecuteActionActivity: sync handler path
  - ExecuteAsyncActionActivity: async handler StartAsync path
  - selectActivityFunc returns string name ("ExecuteActionActivity" / "ExecuteAsyncActionActivity")
  - toResultMap: handles nil/map/struct via JSON roundtrip (int64→float64 lossy >2^53)

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

---

## ActionHandler Interface [sdk]

file: business/sdk/workflow/interfaces.go
```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string
    IsAsync() bool
    SupportsManualExecution() bool
    GetDescription() string
}
```

registered action types (20):
  seek_approval, evaluate_condition, delay, update_field, create_entity,
  lookup_entity, transition_status, log_audit_entry, check_inventory,
  allocate_inventory, reserve_inventory, release_reservation, check_reorder_point,
  receive_inventory, commit_allocation, send_email, send_notification,
  create_alert, create_purchase_order, call_webhook

---

## EdgeStore [db]

file: business/sdk/workflow/temporal/stores/edgedb/edgedb.go
key facts:
  - Read-only: 2 methods (LoadActions, LoadEdges)
  - Custom query (NOT rule_actions_view — view lacks deactivated_by column)
  - LEFT JOIN action_templates for ActionType (NULL template_id → empty string)
  - sql.NullString for nullable UUIDs (deactivated_by, source_action_id)
  - NamedQuerySlice returns nil for empty (NOT ErrDBNotFound)
⊗ workflow.rule_actions
⊗ workflow.action_edges
⊗ workflow.action_templates

---

## WorkflowBus [bus]

file: business/sdk/workflow/workflowbus/ (workflow.Business)
key facts:
  - CRUD for: AutomationRules, RuleActions, ActionEdges, ActionTemplates, Executions
  - Does NOT handle trigger/dispatch — that is TriggerProcessor + WorkflowTrigger
⊗⊕ workflow.automation_rules
⊗⊕ workflow.rule_actions
⊗⊕ workflow.action_edges
⊗⊕ workflow.action_templates
⊗⊕ workflow.automation_executions

---

## DB Schema

workflow.automation_rules       — rule definitions (entity_id, trigger_type, conditions)
workflow.rule_actions           — action nodes attached to rules
workflow.action_edges           — directed edges (type: start / sequence / always)
workflow.action_templates       — reusable action type configs
workflow.automation_executions  — execution history

---

## ⚠ Adding a new ActionHandler

  business/sdk/workflow/interfaces.go                              (confirm ActionHandler interface satisfied)
  business/sdk/workflow/temporal/activities.go                     (AsyncRegistry vs Registry decision)
  api/cmd/services/ichor/build/all/all.go                          (Register() call in ActionRegistry setup)
  business/sdk/dbtest/seedmodels/                                   (new test seed if handler needs domain data)
  docs/workflow/README.md                                           (update handler catalog)

## ⚠ Adding a new Edge type

  business/sdk/workflow/temporal/models.go                         (new EdgeType const — task queue file)
  business/sdk/workflow/temporal/graph_executor.go                 (handle in BFS traversal)
  app/domain/workflow/workflowsaveapp/graph.go                     (cycle/validation logic)
  app/domain/workflow/workflowsaveapp/model.go                     (allowed edge types — see workflow-save.md)
  api/cmd/services/ichor/tests/workflow/                            (integration test update)

## ⚠ Changing WorkflowInput shape

  business/sdk/workflow/temporal/models.go                         (WorkflowInput struct)
  business/sdk/workflow/temporal/workflow.go                       (unpack new fields)
  business/sdk/workflow/temporal/trigger.go                        (populate new fields when dispatching)
  business/sdk/workflow/temporal/graph_executor.go                 (consume new fields if needed)
  apitest/workflow.go                                               (update test infra if WorkflowInfra changes)

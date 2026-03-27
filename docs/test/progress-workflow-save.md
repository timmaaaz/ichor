# Progress Summary: workflow-save.md

## Overview
Architecture for workflow validation and persistence. Handles DAG validation via Kahn's topological sort, action configuration validation, and atomic save operations.

## WorkflowSaveApp [app] — `app/domain/workflow/workflowsaveapp/workflowsaveapp.go`

**Responsibility:** Validate and persist workflow rules, actions, and edges atomically.

### Struct
```go
type App struct {
    log         *logger.Logger
    db          *sqlx.DB
    workflowBus *workflow.Business
    delegate    *delegate.Delegate
    registry    *workflow.ActionRegistry
}
```

### Methods
- `NewApp(log, db, workflowBus, del, registry) *App`
- `DryRunValidate(req SaveWorkflowRequest) ValidationResult` — validation only, no DB writes
- `SaveWorkflow(ctx, ruleID uuid.UUID, req SaveWorkflowRequest) (SaveWorkflowResponse, error)` — validate then persist
- `CreateWorkflow(ctx, userID uuid.UUID, req SaveWorkflowRequest) (SaveWorkflowResponse, error)` — new workflow
- `DuplicateWorkflow(ctx, ruleID uuid.UUID, userID uuid.UUID) (SaveWorkflowResponse, error)` — clone existing

### Key Facts
- **Validates and persists as atomic operation** — rule + actions + edges
- **Two entry points:** DryRunValidate (validation only) and SaveWorkflow (validation + DB write)
- **DryRunValidate** — full validation only, no DB changes, returns ValidationResult with all errors collected
- **SaveWorkflow** — prepareRequest validates, then ReadCommitted [tx]: update rule + sync actions (create/update/delete) + recreate edges + delegate.Call() after commit
- **Default workflows protection** — workflows with IsDefault=true cannot be modified via SaveWorkflow; use DuplicateWorkflow to modify

## Models — `app/domain/workflow/workflowsaveapp/model.go`

### SaveWorkflowRequest
```go
type SaveWorkflowRequest struct {
    Name              string
    Description       string
    IsActive          bool
    EntityID          uuid.UUID
    TriggerTypeID     uuid.UUID
    TriggerConditions json.RawMessage
    Actions           []SaveActionRequest
    Edges             []SaveEdgeRequest
    CanvasLayout      json.RawMessage
}
```

### SaveActionRequest
```go
type SaveActionRequest struct {
    ID           uuid.UUID
    Name         string
    Description  string
    ActionType   string
    ActionConfig json.RawMessage
    IsActive     bool
}
```

### SaveEdgeRequest
```go
type SaveEdgeRequest struct {
    SourceActionID uuid.UUID
    TargetActionID uuid.UUID
    EdgeType       string   // "start" | "sequence" | "always"
    SourceOutput   *string
    EdgeOrder      int
}
```

### SaveWorkflowResponse
```go
type SaveWorkflowResponse struct {
    ID, Name, Description string
    IsActive              bool
    EntityID, TriggerTypeID uuid.UUID
    TriggerConditions     json.RawMessage
    Actions               []SaveActionResponse
    Edges                 []SaveEdgeResponse
    CanvasLayout          json.RawMessage
    CreatedDate, UpdatedDate time.Time
}
```

### ValidationResult
```go
type ValidationResult struct {
    Valid       bool
    Errors      []string
    ActionCount int
    EdgeCount   int
}
```

### Action Config Structs (Per ActionType)
```go
CreateAlertConfig       { AlertType, Severity, Title, Message string }
SendEmailConfig         { Recipients []string, Subject, Body string }
SendNotificationConfig  { Recipients, Channels []string }
UpdateFieldConfig       { TargetEntity, TargetField string }
SeekApprovalConfig      { Approvers []string, ApprovalType string }
AllocateInventoryConfig { InventoryItems []any, AllocationMode string, SourceFromLineItem bool }
EvaluateConditionConfig { Conditions []any }
```

## GraphValidation [app] — `app/domain/workflow/workflowsaveapp/graph.go`

**Algorithm:** Kahn's topological sort (in-degree tracking)

### Validation Checks (In Order)
1. **Empty workflows valid** — workflows with no actions pass (draft mode, skip remaining checks)
2. **At least one edge required** — workflows with actions must have at least one edge
3. **Exactly one start edge required** — workflow entry point
4. **No cycles** — Kahn's algorithm detects cycles
5. **All actions reachable from start** — BFS reachability check
6. **start and always edges** — must NOT have source_output specified
7. **sequence edges with source_output** — must reference valid output ports from ActionRegistry
8. **Action config validation** — per type (required fields validated)
9. **Existing action references** — must match actions belonging to the rule being updated

### Edge Types
- **start** — workflow entry point (exactly one required)
- **sequence** — normal control flow
- **always** — unconditional execution (independent of previous action result)

## Change Patterns

### ⚠ Adding a New Edge Type
Affects 5 areas:
1. `app/domain/workflow/workflowsaveapp/model.go` — add to EdgeType allowed values
2. `app/domain/workflow/workflowsaveapp/graph.go` — handle in validation (source_output rules)
3. `business/sdk/workflow/temporal/models.go` — add EdgeType const
4. `business/sdk/workflow/temporal/graph_executor.go` — BFS traversal handling
5. `api/cmd/services/ichor/tests/workflow/` — integration tests

### ⚠ Adding a New Action Type with Required Config Fields
Affects 6 areas:
1. `business/sdk/workflow/interfaces.go` — new ActionHandler implementing Validate()
2. `app/domain/workflow/workflowsaveapp/model.go` — new *Config struct
3. `app/domain/workflow/workflowsaveapp/graph.go` — add config validation case
4. `business/sdk/workflow/temporal/activities.go` — Registry vs AsyncRegistry registration
5. `api/cmd/services/ichor/build/all/all.go` — Register() in ActionRegistry setup
6. `docs/workflow/README.md` — update handler catalog

### ⚠ Changing ValidationResult Shape
Affects 3 areas:
1. `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` — DryRunValidate return
2. `app/domain/workflow/workflowsaveapp/model.go` — ValidationResult struct
3. All callers of DryRunValidate — `app/domain/workflow/...` API layer

## Critical Points
- **Exactly one start edge** — enforced; no workflow can have multiple entry points
- **No cycles** — Kahn's algorithm catches cycles; prevents infinite loops at execution time
- **Reachability required** — all actions must be reachable from start (dead code detection)
- **Output port validation** — sequence edges with source_output must reference real action outputs
- **Action config validation** — per-type validation happens before persistence
- **Default workflows protected** — IsDefault=true workflows cannot be modified directly

## Notes for Future Development
Workflow validation is sophisticated and critical for execution safety. Most changes will be:
- Adding new action types (moderate, requires config validation + registry entry)
- Adding new edge types (moderate, requires graph traversal updates)
- Changing validation rules (risky, affects all existing workflows)

The Kahn's algorithm + BFS reachability combo ensures workflows are always DAGs with no dead code — essential for the Temporal execution engine.

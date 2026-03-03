# workflow-save

[app]=application layer [bus]=business layer [sdk]=shared
→=depends on ⊕=writes ⊗=reads [tx]=transaction

---

## Overview

Validates and persists workflow rule + actions + edges as a single atomic operation.
Two entry points: DryRunValidate (no DB writes, returns all errors) and SaveWorkflow (DB write).
Graph validation uses Kahn's algorithm for cycle detection.

---

## WorkflowSaveApp [app]

file: app/domain/workflow/workflowsaveapp/workflowsaveapp.go
```go
type App struct {
    log         *logger.Logger
    db          *sqlx.DB
    workflowBus *workflow.Business
    delegate    *delegate.Delegate
    registry    *workflow.ActionRegistry
}
```

api:
  NewApp(log, db, workflowBus, del, registry) *App
  DryRunValidate(req SaveWorkflowRequest) ValidationResult
  SaveWorkflow(ctx, ruleID uuid.UUID, req SaveWorkflowRequest) (SaveWorkflowResponse, error)
  CreateWorkflow(ctx, userID uuid.UUID, req SaveWorkflowRequest) (SaveWorkflowResponse, error)
  DuplicateWorkflow(ctx, ruleID uuid.UUID, userID uuid.UUID) (SaveWorkflowResponse, error)

path differences:
  DryRunValidate — full validation only, no DB changes, returns ValidationResult with all errors collected
  SaveWorkflow   — prepareRequest validates, then ReadCommitted [tx]: update rule + sync actions (create/update/delete) + recreate edges + delegate.Call() after commit

constraints:
  - Default workflows (IsDefault=true) cannot be modified via SaveWorkflow — use DuplicateWorkflow

---

## Models

file: app/domain/workflow/workflowsaveapp/model.go

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

type SaveActionRequest struct {
    ID           uuid.UUID
    Name         string
    Description  string
    ActionType   string
    ActionConfig json.RawMessage
    IsActive     bool
}

type SaveEdgeRequest struct {
    SourceActionID uuid.UUID
    TargetActionID uuid.UUID
    EdgeType       string   // "start" | "sequence" | "always"
    SourceOutput   *string
    EdgeOrder      int
}

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

type ValidationResult struct {
    Valid       bool
    Errors      []string
    ActionCount int
    EdgeCount   int
}
```

### Action config structs (per ActionType)

```go
CreateAlertConfig       { AlertType, Severity, Title, Message string }
SendEmailConfig         { Recipients []string, Subject, Body string }
SendNotificationConfig  { Recipients, Channels []string }
UpdateFieldConfig       { TargetEntity, TargetField string }
SeekApprovalConfig      { Approvers []string, ApprovalType string }
AllocateInventoryConfig { InventoryItems []any, AllocationMode string, SourceFromLineItem bool }
EvaluateConditionConfig { Conditions []any }
```

---

## Graph Validation

file: app/domain/workflow/workflowsaveapp/graph.go
algorithm: Kahn's topological sort (in-degree tracking)

validation checks (in order):
  1. Workflows with no actions are valid (draft mode — skip remaining checks)
  2. Workflows with actions must have at least one edge
  3. Exactly one start edge required
  4. No cycles (Kahn's algorithm)
  5. All actions reachable from start edge (BFS reachability check)
  6. start and always edges must NOT have source_output specified
  7. sequence edges with source_output must reference valid output ports from ActionRegistry
  8. Action config validation per type (required fields, see config structs above)
  9. Existing action references must match actions belonging to the rule being updated

---

## ⚠ Adding a new edge type

  app/domain/workflow/workflowsaveapp/model.go              (add to EdgeType allowed values)
  app/domain/workflow/workflowsaveapp/graph.go              (handle in validation — source_output rules)
  business/sdk/workflow/temporal/models.go                  (add EdgeType const)
  business/sdk/workflow/temporal/graph_executor.go          (BFS traversal handling)
  api/cmd/services/ichor/tests/workflow/                    (integration tests)

## ⚠ Adding a new action type with required config fields

  business/sdk/workflow/interfaces.go                       (new ActionHandler implementing Validate())
  app/domain/workflow/workflowsaveapp/model.go              (new *Config struct)
  app/domain/workflow/workflowsaveapp/graph.go              (add config validation case)
  business/sdk/workflow/temporal/activities.go              (Registry vs AsyncRegistry registration)
  api/cmd/services/ichor/build/all/all.go                   (Register() in ActionRegistry setup)
  docs/workflow/README.md                                    (handler catalog)

## ⚠ Changing ValidationResult shape

  app/domain/workflow/workflowsaveapp/workflowsaveapp.go    (DryRunValidate return)
  app/domain/workflow/workflowsaveapp/model.go              (ValidationResult struct)
  All callers of DryRunValidate                              (app/domain/workflow/... API layer)

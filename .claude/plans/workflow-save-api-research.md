# Workflow Save API - Research & Implementation Plan

## Overview

This document contains research findings and an implementation plan for a **transactional workflow save API** for the Ichor backend. The frontend workflow editor needs to save complete workflows (triggers, actions, edges, conditions) atomically.

---

## Research Findings Summary

### Phase 1: Workflow Domain Analysis

**Architecture Discovery**: The workflow domain uses a **single monolithic `workflow.Business`** struct (not separate `*bus` packages like other domains).

**Location**: `business/sdk/workflow/workflowbus.go`

**Key Characteristics**:
- All workflow operations in one Business struct
- Full transaction support via `NewWithTx(tx sqldb.CommitRollbacker)`
- Storer interface abstracts database operations

**Available CRUD Methods**:
```
Rules:     CreateRule, UpdateRule, QueryRuleByID, DeactivateRule, ActivateRule
Actions:   CreateRuleAction, UpdateRuleAction, QueryActionsByRule, DeactivateRuleAction
Edges:     CreateActionEdge, QueryEdgesByRuleID, DeleteActionEdge, DeleteEdgesByRuleID
Templates: CreateActionTemplate, UpdateActionTemplate
Executions: CreateExecution, QueryExecutionHistory
```

**Current API Endpoints** (in `api/domain/http/workflow/`):
- `ruleapi` - Rule CRUD (creates rule + embedded actions but NOT transactional)
- `edgeapi` - Edge CRUD
- `alertapi` - Alert management
- `executionapi` - Execution history
- `referenceapi` - Reference data

**Current Gap**: Rule create endpoint creates rule first, then actions sequentially - NOT in a transaction. If action creation fails, rule is orphaned.

---

### Phase 2: FormData System Analysis

**Key Insight**: FormData already solves multi-entity transactional saves and is production-tested with 45+ entity types.

**Pattern Summary**:
```go
tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
defer tx.Rollback()

// Execute operations in order with template variable resolution
for _, step := range plan {
    result, err := executeOperation(ctx, step, templateContext)
    if err != nil {
        return FormDataResponse{}, err  // Auto-rollback via defer
    }
    templateContext[step.EntityName] = result  // Available for FK templates
}

tx.Commit()
```

**Template Variable Resolution**: Supports `{{entity.field}}` syntax for FK references between entities created in same transaction.

**Array Support**: Can create multiple children (like line items) in single operation.

**Error Handling**: Typed errors (`errs.Error`) propagate with appropriate HTTP status codes.

**Applicability to Workflows**: HIGHLY applicable - same pattern (parent + children + relationships).

---

### Phase 3: Transaction Patterns Analysis

**Core Interfaces** (`business/sdk/sqldb/tran.go`):
```go
type Beginner interface {
    Begin() (CommitRollbacker, error)
}

type CommitRollbacker interface {
    Commit() error
    Rollback() error
}
```

**NewWithTx Pattern** (used throughout codebase):
```go
// Business layer
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
    storer, err := b.storer.NewWithTx(tx)
    if err != nil {
        return nil, err
    }
    return &Business{log: b.log, storer: storer, del: b.del}, nil
}

// Store layer
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (Storer, error) {
    ec, err := sqldb.GetExtContext(tx)  // Extract sqlx.ExtContext
    if err != nil {
        return nil, err
    }
    return &Store{log: s.log, db: ec}, nil
}
```

**Middleware Option** (`app/sdk/mid/transaction.go`): `BeginCommitRollback()` can wrap handlers with automatic transaction management.

---

## Implementation Approach Decision

### Option A: Register in FormData (Faster)
- Register `workflow.rules`, `workflow.rule_actions`, `workflow.edges` in formdata_registry.go
- Use existing FormData endpoint `/v1/formdata/{form_id}/upsert`
- **Pros**: Minimal code, reuses production-tested system
- **Cons**: Generic error messages, no workflow-specific validation, frontend must build FormData requests

### Option B: Dedicated Workflow Save Endpoint (Recommended)
- Create `PUT /v1/workflow/rules/{id}/full` for atomic saves
- Create `POST /v1/workflow/rules/full` for new workflows
- Follow FormData transaction patterns but with workflow-specific logic
- **Pros**: Clear API, workflow-specific validation, better error messages, graph validation
- **Cons**: More code to write

**Recommendation**: Option B - Dedicated endpoint provides better UX and enables workflow-specific features like graph cycle detection.

---

## API Contract Design

### Endpoint: `PUT /v1/workflow/rules/{id}/full`

Updates an existing rule with all its actions and edges atomically.

**Request Body**:
```json
{
  "name": "Order Status Notification",
  "description": "Send notifications when order status changes",
  "is_active": true,
  "entity_id": "uuid",
  "trigger_type_id": "uuid",
  "trigger_conditions": {
    "field_conditions": [
      {
        "field_name": "status",
        "operator": "changed_to",
        "value": "shipped"
      }
    ]
  },
  "actions": [
    {
      "id": "existing-action-uuid",
      "name": "Send Email",
      "action_config": {
        "recipients": ["{{customer_email}}"],
        "subject": "Order {{number}} Shipped",
        "body": "Your order has shipped!"
      },
      "execution_order": 1
    },
    {
      "id": null,
      "name": "Create Alert",
      "action_config": {
        "alert_type": "order_shipped",
        "severity": "low",
        "title": "Order shipped",
        "message": "Order {{number}} shipped to {{customer_name}}"
      },
      "execution_order": 1
    }
  ],
  "edges": [
    {
      "source_action_id": null,
      "target_action_id": "temp:0",
      "edge_type": "start"
    },
    {
      "source_action_id": "temp:0",
      "target_action_id": "temp:1",
      "edge_type": "sequence"
    }
  ],
  "canvas_layout": {
    "viewport": {"x": 0, "y": 0, "zoom": 1},
    "node_positions": {
      "trigger": {"x": 100, "y": 100},
      "temp:0": {"x": 300, "y": 100},
      "temp:1": {"x": 300, "y": 250}
    }
  }
}
```

**Key Design Decisions**:
1. **Action IDs**: Existing actions have UUID, new actions have `null` ID
2. **Edge References**: Use `"temp:N"` syntax to reference new actions by array index
3. **Canvas Layout**: Store in rule metadata (JSONB column) or separate table
4. **Start Edges**: Have `null` source_action_id by convention

**Response (Success)**:
```json
{
  "id": "rule-uuid",
  "name": "Order Status Notification",
  "description": "...",
  "is_active": true,
  "entity_id": "uuid",
  "trigger_type_id": "uuid",
  "trigger_conditions": {...},
  "actions": [
    {
      "id": "existing-action-uuid",
      "name": "Send Email",
      "action_config": {...},
      "execution_order": 1
    },
    {
      "id": "newly-created-action-uuid",
      "name": "Create Alert",
      "action_config": {...},
      "execution_order": 1
    }
  ],
  "edges": [
    {
      "id": "edge-uuid-1",
      "source_action_id": null,
      "target_action_id": "existing-action-uuid",
      "edge_type": "start"
    },
    {
      "id": "edge-uuid-2",
      "source_action_id": "existing-action-uuid",
      "target_action_id": "newly-created-action-uuid",
      "edge_type": "sequence"
    }
  ],
  "canvas_layout": {...},
  "created_date": "...",
  "updated_date": "..."
}
```

**Response (Validation Error)**:
```json
{
  "error": "invalid_argument",
  "message": "workflow validation failed",
  "details": {
    "validation_errors": [
      {"field": "actions[1].action_config.recipients", "message": "at least one recipient required"},
      {"field": "edges", "message": "cycle detected between actions"}
    ]
  }
}
```

### Endpoint: `POST /v1/workflow/rules/full`

Creates a new rule with all its actions and edges atomically. Same request/response format as PUT.

---

## Validation Requirements

### Pre-Save Validation (App Layer)

1. **Rule Validation**:
   - `name` is required and non-empty
   - `entity_id` must exist in `workflow.entities`
   - `trigger_type_id` must exist in `workflow.trigger_types`
   - `trigger_conditions` JSON must be valid structure

2. **Action Validation**:
   - Each action must have `name` and `action_config`
   - `action_config` must match schema for action type
   - `execution_order` must be positive integer

3. **Edge Validation**:
   - `start` edges must have `null` source_action_id
   - Non-start edges must have valid source_action_id
   - `target_action_id` must reference valid action (existing or temp)
   - `edge_type` must be one of: `start`, `sequence`, `true_branch`, `false_branch`, `always`

4. **Graph Validation**:
   - No cycles in the action graph
   - All actions must be reachable from a start edge
   - No orphaned actions (actions not connected to graph)

---

## Implementation Plan

### Files to Create

```
app/domain/workflow/workflowsaveapp/
├── workflowsaveapp.go      # Main save orchestration
├── model.go                 # Request/Response models
├── validation.go            # Workflow-specific validation
├── graph.go                 # Graph cycle detection, reachability

api/domain/http/workflow/workflowsaveapi/
├── workflowsaveapi.go       # HTTP handlers
├── route.go                 # Route definitions
├── model.go                 # API-specific models (if different from app)
```

### Files to Modify

1. **`api/cmd/services/ichor/build/all/all.go`**:
   - Wire up new `workflowsaveapi.Routes()`

2. **`business/sdk/workflow/workflowbus.go`** (optional):
   - Add `SaveWorkflow()` method if we want to encapsulate the transaction logic in business layer
   - Or handle transaction at app layer (FormData pattern)

3. **`business/sdk/migrate/sql/migrate.sql`** (if needed):
   - Add `canvas_layout JSONB` column to `workflow.automation_rules` for storing node positions

### Implementation Steps

#### Step 1: Database Schema (if storing canvas layout)

```sql
-- Version: X.XX
-- Description: Add canvas_layout to automation_rules for visual editor state
ALTER TABLE workflow.automation_rules
ADD COLUMN canvas_layout JSONB DEFAULT '{}';
```

#### Step 2: App Layer Models (`app/domain/workflow/workflowsaveapp/model.go`)

```go
type SaveWorkflowRequest struct {
    Name              string             `json:"name" validate:"required"`
    Description       string             `json:"description"`
    IsActive          bool               `json:"is_active"`
    EntityID          string             `json:"entity_id" validate:"required,uuid"`
    TriggerTypeID     string             `json:"trigger_type_id" validate:"required,uuid"`
    TriggerConditions json.RawMessage    `json:"trigger_conditions"`
    Actions           []SaveActionRequest `json:"actions" validate:"required,min=1"`
    Edges             []SaveEdgeRequest   `json:"edges"`
    CanvasLayout      json.RawMessage    `json:"canvas_layout"`
}

type SaveActionRequest struct {
    ID             *string         `json:"id"`  // null for new
    Name           string          `json:"name" validate:"required"`
    ActionConfig   json.RawMessage `json:"action_config" validate:"required"`
    ExecutionOrder int             `json:"execution_order" validate:"required,min=1"`
}

type SaveEdgeRequest struct {
    SourceActionID string `json:"source_action_id"`  // null/"" for start edges, or "temp:N"
    TargetActionID string `json:"target_action_id" validate:"required"`  // UUID or "temp:N"
    EdgeType       string `json:"edge_type" validate:"required,oneof=start sequence true_branch false_branch always"`
    EdgeOrder      int    `json:"edge_order"`
}

type SaveWorkflowResponse struct {
    ID               string                   `json:"id"`
    Name             string                   `json:"name"`
    // ... full rule representation with resolved IDs
    Actions          []SaveActionResponse     `json:"actions"`
    Edges            []SaveEdgeResponse       `json:"edges"`
}
```

#### Step 3: Graph Validation (`app/domain/workflow/workflowsaveapp/graph.go`)

```go
func ValidateGraph(actions []SaveActionRequest, edges []SaveEdgeRequest) error {
    // 1. Build adjacency list
    // 2. Check for cycles using DFS
    // 3. Check all nodes reachable from start edges
    // 4. Check no orphaned actions
    return nil
}
```

#### Step 4: Main Save Logic (`app/domain/workflow/workflowsaveapp/workflowsaveapp.go`)

```go
func (a *App) SaveWorkflow(ctx context.Context, ruleID uuid.UUID, req SaveWorkflowRequest) (SaveWorkflowResponse, error) {
    // 1. Validate request
    if err := req.Validate(); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.InvalidArgument, "validation: %s", err)
    }

    // 2. Validate graph structure
    if err := ValidateGraph(req.Actions, req.Edges); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.InvalidArgument, "graph: %s", err)
    }

    // 3. Begin transaction
    tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
    if err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "begin tx: %s", err)
    }
    defer tx.Rollback()

    // 4. Get transaction-aware business layer
    txBus, err := a.workflowBus.NewWithTx(tx)
    if err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "new with tx: %s", err)
    }

    // 5. Update rule metadata
    rule, err := txBus.UpdateRule(ctx, ruleID, toBusUpdateRule(req))
    if err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "update rule: %s", err)
    }

    // 6. Sync actions (diff: create new, update existing, delete removed)
    actionIDMap, err := a.syncActions(ctx, txBus, ruleID, req.Actions)
    if err != nil {
        return SaveWorkflowResponse{}, err
    }

    // 7. Delete all existing edges, recreate from request
    if err := txBus.DeleteEdgesByRuleID(ctx, ruleID); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "delete edges: %s", err)
    }

    edges, err := a.createEdges(ctx, txBus, ruleID, req.Edges, actionIDMap)
    if err != nil {
        return SaveWorkflowResponse{}, err
    }

    // 8. Commit transaction
    if err := tx.Commit(); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "commit: %s", err)
    }

    // 9. Build and return response
    return buildResponse(rule, actionIDMap, edges), nil
}
```

#### Step 5: API Handler (`api/domain/http/workflow/workflowsaveapi/workflowsaveapi.go`)

```go
func (api *api) save(ctx context.Context, r *http.Request) web.Encoder {
    ruleID, err := uuid.Parse(web.Param(r, "id"))
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    var req workflowsaveapp.SaveWorkflowRequest
    if err := web.Decode(r, &req); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    resp, err := api.app.SaveWorkflow(ctx, ruleID, req)
    if err != nil {
        return errs.NewError(err)
    }

    return resp
}
```

#### Step 6: Route Registration (`api/domain/http/workflow/workflowsaveapi/route.go`)

```go
func Routes(app *web.App, cfg Config) {
    const version = "v1"
    api := newAPI(workflowsaveapp.NewApp(cfg.DB, cfg.WorkflowBus))
    authen := mid.Authenticate(cfg.AuthClient)

    app.HandlerFunc(http.MethodPut, version, "/workflow/rules/{id}/full", api.save, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "workflow.automation_rules", permissionsbus.Actions.Update, auth.RuleAdminOnly))

    app.HandlerFunc(http.MethodPost, version, "/workflow/rules/full", api.create, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "workflow.automation_rules", permissionsbus.Actions.Create, auth.RuleAdminOnly))
}
```

#### Step 7: Wire Up in all.go

```go
// In api/cmd/services/ichor/build/all/all.go

import "github.com/timmaaaz/ichor/api/domain/http/workflow/workflowsaveapi"

// In Routes() function:
workflowsaveapi.Routes(app, workflowsaveapi.Config{
    Log:            cfg.Log,
    DB:             cfg.DB,
    WorkflowBus:    workflowBus,  // existing workflow business instance
    AuthClient:     cfg.AuthClient,
    PermissionsBus: permissionsBus,
})
```

---

## Testing Plan

### Unit Tests

1. **Graph validation tests**:
   - Detect cycles
   - Detect orphaned nodes
   - Detect unreachable nodes
   - Valid graph passes

2. **Action sync tests**:
   - Create new actions
   - Update existing actions
   - Delete removed actions
   - Preserve action order

3. **Edge creation tests**:
   - Resolve temp IDs to real UUIDs
   - Start edge validation
   - Edge type validation

### Integration Tests

1. **Full save flow**:
   - Create new workflow with actions and edges
   - Update existing workflow
   - Partial update (only change some actions)

2. **Transaction rollback**:
   - Invalid action config → entire save rolls back
   - Cycle in graph → entire save rolls back
   - Database error → entire save rolls back

3. **Permission tests**:
   - Admin can save
   - Non-admin cannot save
   - User can only save their own rules (if applicable)

---

## Open Questions for User

1. **Canvas Layout Storage**: Should we add a `canvas_layout` JSONB column to `automation_rules`, or create a separate `workflow.rule_canvas` table?

2. **Action Type Validation**: Should we validate that `action_config` matches the expected schema for each action type (e.g., `send_email` must have `recipients`, `subject`, `body`)? This adds complexity but catches errors early.

3. **Delete vs Soft-Delete**: When actions are removed from a workflow, should we hard-delete them or soft-delete (set `is_active = false`)? Soft-delete preserves history but leaves orphaned records.

4. **Create Endpoint**: Do you need `POST /v1/workflow/rules/full` for creating new workflows, or will workflows always be created empty first and then saved?

---

## Summary

The implementation follows established patterns from the FormData system:
- Transaction orchestration at app layer
- `NewWithTx()` pattern for transaction-aware business layers
- Typed error propagation
- Comprehensive validation before save

The key difference is workflow-specific features:
- Graph cycle detection
- Temp ID resolution for new actions in edges
- Action sync (diff-based create/update/delete)

Estimated implementation effort: 3-5 files, ~500-800 lines of Go code, plus tests.

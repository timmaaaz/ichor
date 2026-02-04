# Workflow Save API Implementation Plan

## Goal
Create transactional API endpoints to save complete workflows atomically (rule + actions + edges).

## Design Decisions
- Canvas layout: JSONB column on `automation_rules`
- Action config: Strict schema validation per action type
- Removed actions: Hard delete
- Endpoints: Both PUT (update) and POST (create)

---

## Phase 1: Database Schema Update

**File**: `business/sdk/migrate/sql/migrate.sql`

Modify the existing `automation_rules` CREATE TABLE statement (Version 1.66) to add the `canvas_layout` column:

```sql
-- Version: 1.66
-- Description: Create table automation_rules
CREATE TABLE workflow.automation_rules (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   name VARCHAR(100) NOT NULL,
   description TEXT,
   entity_id UUID NOT NULL, -- table or view name, maybe others in the future
   entity_type_id UUID NOT NULL REFERENCES workflow.entity_types(id),

   -- Trigger conditions
   trigger_type_id UUID NOT NULL REFERENCES workflow.trigger_types(id),

   trigger_conditions JSONB NULL, -- When to trigger

   -- Visual editor state
   canvas_layout JSONB DEFAULT '{}',

   -- Control
   is_active BOOLEAN NOT NULL DEFAULT TRUE,

   created_date TIMESTAMP NOT NULL DEFAULT NOW(),
   updated_date TIMESTAMP NOT NULL DEFAULT NOW(),
   created_by UUID NOT NULL REFERENCES core.users(id),
   updated_by UUID NOT NULL REFERENCES core.users(id),

   deactivated_by UUID NULL REFERENCES core.users(id)
);
```

---

## Phase 2: Propagate canvas_layout Through Existing Domain

Adding `canvas_layout` to the schema requires updates across the existing workflow domain layers.

### 2a. Business Layer Models

**File**: `business/sdk/workflow/models.go`

| Struct | Add Field | Type |
|--------|-----------|------|
| `AutomationRule` | `CanvasLayout` | `json.RawMessage` |
| `NewAutomationRule` | `CanvasLayout` | `json.RawMessage` |
| `UpdateAutomationRule` | `CanvasLayout` | `*json.RawMessage` (pointer for optional) |
| `AutomationRuleView` | `CanvasLayout` | `json.RawMessage` |

### 2b. Database Store Models

**File**: `business/sdk/workflow/stores/workflowdb/models.go`

| Item | Change |
|------|--------|
| `automationRule` struct | Add `CanvasLayout sql.NullString` with `db:"canvas_layout"` tag |
| `automationRulesView` struct | Add `CanvasLayout sql.NullString` with `db:"canvas_layout"` tag |
| `toCoreAutomationRule()` | Add unmarshal: `if db.CanvasLayout.Valid { rule.CanvasLayout = json.RawMessage(db.CanvasLayout.String) }` |
| `toDBAutomationRule()` | Add marshal: set `CanvasLayout` to `sql.NullString{String: string(bus.CanvasLayout), Valid: len(bus.CanvasLayout) > 0}` |
| `toCoreAutomationRuleView()` | Add canvas_layout conversion (same pattern) |

### 2c. Database Store Queries

**File**: `business/sdk/workflow/stores/workflowdb/workflowdb.go`

Update SQL statements to include `canvas_layout`:

| Function | Change |
|----------|--------|
| `CreateRule()` | Add `canvas_layout` to INSERT columns and `:canvas_layout` to VALUES |
| `UpdateRule()` | Add `canvas_layout = :canvas_layout` to SET clause |
| `QueryRuleByID()` | Add `canvas_layout` to SELECT columns |
| `QueryRulesByEntity()` | Add `canvas_layout` to SELECT columns |
| `QueryActiveRules()` | Add `canvas_layout` to SELECT columns |
| `QueryAutomationRulesView()` | Add `canvas_layout` to SELECT columns |
| `QueryAutomationRulesViewPaginated()` | Add `canvas_layout` to SELECT columns |

### 2d. API Layer Models (Existing ruleapi)

**File**: `api/domain/http/workflow/ruleapi/model.go`

| Struct | Add Field | JSON Tag |
|--------|-----------|----------|
| `CreateRuleRequest` | `CanvasLayout json.RawMessage` | `json:"canvas_layout,omitempty"` |
| `UpdateRuleRequest` | `CanvasLayout *json.RawMessage` | `json:"canvas_layout,omitempty"` |
| `RuleResponse` | `CanvasLayout json.RawMessage` | `json:"canvas_layout"` |

**File**: `api/domain/http/workflow/ruleapi/converters.go`

| Function | Change |
|----------|--------|
| `toRuleResponse()` | Add `CanvasLayout: view.CanvasLayout` |
| `toNewAutomationRule()` | Add `CanvasLayout: req.CanvasLayout` |
| `toUpdateAutomationRule()` | Add `CanvasLayout: req.CanvasLayout` |

### 2e. Test Files

Update all `NewAutomationRule` instantiations to include `CanvasLayout: nil`:

| File | Change |
|------|--------|
| `api/cmd/services/ichor/tests/workflow/ruleapi/seed_test.go` | Add field to seed data |
| `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_seed_test.go` | Add field to all rule instantiations |
| `api/cmd/services/ichor/tests/workflow/ruleapi/create_test.go` | Add field to test requests |
| `api/cmd/services/ichor/tests/workflow/ruleapi/update_test.go` | Add field to test requests |

---

## Phase 3: Create App Layer

### `app/domain/workflow/workflowsaveapp/model.go`

```go
package workflowsaveapp

import (
    "encoding/json"
    "github.com/timmaaaz/ichor/app/sdk/errs"
)

type SaveWorkflowRequest struct {
    Name              string               `json:"name" validate:"required,min=1,max=255"`
    Description       string               `json:"description" validate:"max=1000"`
    IsActive          bool                 `json:"is_active"`
    EntityID          string               `json:"entity_id" validate:"required,uuid"`
    TriggerTypeID     string               `json:"trigger_type_id" validate:"required,uuid"`
    TriggerConditions json.RawMessage      `json:"trigger_conditions"`
    Actions           []SaveActionRequest  `json:"actions" validate:"required,min=1,dive"`
    Edges             []SaveEdgeRequest    `json:"edges" validate:"dive"`
    CanvasLayout      json.RawMessage      `json:"canvas_layout"`
}

func (r *SaveWorkflowRequest) Decode(data []byte) error {
    return json.Unmarshal(data, r)
}

func (r SaveWorkflowRequest) Validate() error {
    if err := errs.Check(r); err != nil {
        return errs.Newf(errs.InvalidArgument, "validate: %s", err)
    }
    return nil
}

type SaveActionRequest struct {
    ID             *string         `json:"id"`
    Name           string          `json:"name" validate:"required,min=1,max=255"`
    Description    string          `json:"description" validate:"max=1000"`
    ActionType     string          `json:"action_type" validate:"required,oneof=create_alert send_email send_notification update_field seek_approval allocate_inventory evaluate_condition"`
    ActionConfig   json.RawMessage `json:"action_config" validate:"required"`
    ExecutionOrder int             `json:"execution_order" validate:"required,min=1"`
    IsActive       bool            `json:"is_active"`
}

type SaveEdgeRequest struct {
    SourceActionID string `json:"source_action_id"`
    TargetActionID string `json:"target_action_id" validate:"required"`
    EdgeType       string `json:"edge_type" validate:"required,oneof=start sequence true_branch false_branch always"`
    EdgeOrder      int    `json:"edge_order" validate:"min=0"`
}

type SaveWorkflowResponse struct {
    ID                string                `json:"id"`
    Name              string                `json:"name"`
    Description       string                `json:"description"`
    IsActive          bool                  `json:"is_active"`
    EntityID          string                `json:"entity_id"`
    TriggerTypeID     string                `json:"trigger_type_id"`
    TriggerConditions json.RawMessage       `json:"trigger_conditions"`
    Actions           []SaveActionResponse  `json:"actions"`
    Edges             []SaveEdgeResponse    `json:"edges"`
    CanvasLayout      json.RawMessage       `json:"canvas_layout"`
    CreatedDate       string                `json:"created_date"`
    UpdatedDate       string                `json:"updated_date"`
}

func (r SaveWorkflowResponse) Encode() ([]byte, string, error) {
    data, err := json.Marshal(r)
    return data, "application/json", err
}

type SaveActionResponse struct {
    ID             string          `json:"id"`
    Name           string          `json:"name"`
    Description    string          `json:"description"`
    ActionType     string          `json:"action_type"`
    ActionConfig   json.RawMessage `json:"action_config"`
    ExecutionOrder int             `json:"execution_order"`
    IsActive       bool            `json:"is_active"`
}

type SaveEdgeResponse struct {
    ID             string `json:"id"`
    SourceActionID string `json:"source_action_id"`
    TargetActionID string `json:"target_action_id"`
    EdgeType       string `json:"edge_type"`
    EdgeOrder      int    `json:"edge_order"`
}
```

### `app/domain/workflow/workflowsaveapp/graph.go`

Implement cycle detection and reachability validation using Kahn's algorithm (topological sort).

Key functions:
- `ValidateGraph(actions []SaveActionRequest, edges []SaveEdgeRequest) error`
- `detectCycles()` - Uses BFS topological sort
- `checkReachability()` - BFS from start edges to verify all actions reachable

### `app/domain/workflow/workflowsaveapp/validation.go`

Validate action configs per type:

| Action Type | Required Fields |
|-------------|-----------------|
| `create_alert` | `alert_type`, `severity`, `title`, `message` |
| `send_email` | `recipients`, `subject`, `body` |
| `send_notification` | `recipients`, `channels` |
| `update_field` | `target_entity`, `target_field` |
| `seek_approval` | `approvers`, `approval_type` |
| `allocate_inventory` | `inventory_items`, `allocation_mode` |
| `evaluate_condition` | `conditions` |

### `app/domain/workflow/workflowsaveapp/workflowsaveapp.go`

```go
package workflowsaveapp

import (
    "context"
    "database/sql"
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/timmaaaz/ichor/app/sdk/errs"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/foundation/logger"
)

type App struct {
    log         *logger.Logger
    db          *sqlx.DB
    workflowBus *workflow.Business
}

func NewApp(log *logger.Logger, db *sqlx.DB, workflowBus *workflow.Business) *App {
    return &App{log: log, db: db, workflowBus: workflowBus}
}

func (a *App) SaveWorkflow(ctx context.Context, ruleID uuid.UUID, req SaveWorkflowRequest) (SaveWorkflowResponse, error) {
    // 1. Validate request
    if err := req.Validate(); err != nil {
        return SaveWorkflowResponse{}, err
    }

    // 2. Validate action configs
    if err := ValidateActionConfigs(req.Actions); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.InvalidArgument, "action config: %s", err)
    }

    // 3. Validate graph
    if err := ValidateGraph(req.Actions, req.Edges); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.InvalidArgument, "graph: %s", err)
    }

    // 4. Begin transaction
    tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
    if err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "begin tx: %s", err)
    }
    defer tx.Rollback()

    // 5. Get transaction-aware business
    txBus, err := a.workflowBus.NewWithTx(tx)
    if err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "new with tx: %s", err)
    }

    // 6. Update rule
    rule, err := a.updateRule(ctx, txBus, ruleID, req)
    if err != nil {
        return SaveWorkflowResponse{}, err
    }

    // 7. Sync actions (create/update/delete)
    actionIDMap, savedActions, err := a.syncActions(ctx, txBus, ruleID, req.Actions)
    if err != nil {
        return SaveWorkflowResponse{}, err
    }

    // 8. Delete and recreate edges
    if err := txBus.DeleteEdgesByRuleID(ctx, ruleID); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "delete edges: %s", err)
    }

    savedEdges, err := a.createEdges(ctx, txBus, ruleID, req.Edges, actionIDMap)
    if err != nil {
        return SaveWorkflowResponse{}, err
    }

    // 9. Commit
    if err := tx.Commit(); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.Internal, "commit: %s", err)
    }

    return buildResponse(rule, savedActions, savedEdges, req.CanvasLayout), nil
}

func (a *App) CreateWorkflow(ctx context.Context, req SaveWorkflowRequest) (SaveWorkflowResponse, error) {
    // Similar to SaveWorkflow but creates rule first
    // ...
}
```

Helper methods needed:
- `updateRule()` - Updates rule metadata including canvas_layout
- `syncActions()` - Diff existing vs request: create new, update existing, delete removed
- `createEdges()` - Create edges with temp ID resolution (`temp:N` -> real UUID)
- `resolveActionID()` - Convert `temp:N` references to actual UUIDs
- `buildResponse()` - Convert business models to response

---

## Phase 4: Create API Layer

### `api/domain/http/workflow/workflowsaveapi/workflowsaveapi.go`

```go
package workflowsaveapi

import (
    "context"
    "net/http"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
    "github.com/timmaaaz/ichor/app/sdk/errs"
    "github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
    app *workflowsaveapp.App
}

func newAPI(app *workflowsaveapp.App) *api {
    return &api{app: app}
}

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

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
    var req workflowsaveapp.SaveWorkflowRequest
    if err := web.Decode(r, &req); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    resp, err := api.app.CreateWorkflow(ctx, req)
    if err != nil {
        return errs.NewError(err)
    }

    return resp
}
```

### `api/domain/http/workflow/workflowsaveapi/route.go`

```go
package workflowsaveapi

import (
    "net/http"

    "github.com/jmoiron/sqlx"
    "github.com/timmaaaz/ichor/api/sdk/http/mid"
    "github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
    "github.com/timmaaaz/ichor/app/sdk/auth"
    "github.com/timmaaaz/ichor/app/sdk/authclient"
    "github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
    Log            *logger.Logger
    DB             *sqlx.DB
    WorkflowBus    *workflow.Business
    AuthClient     *authclient.Client
    PermissionsBus *permissionsbus.Business
}

const RouteTable = "workflow.automation_rules"

func Routes(app *web.App, cfg Config) {
    const version = "v1"

    workflowApp := workflowsaveapp.NewApp(cfg.Log, cfg.DB, cfg.WorkflowBus)
    api := newAPI(workflowApp)
    authen := mid.Authenticate(cfg.AuthClient)

    app.HandlerFunc(http.MethodPut, version, "/workflow/rules/{id}/full", api.save, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAdminOnly))

    app.HandlerFunc(http.MethodPost, version, "/workflow/rules/full", api.create, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAdminOnly))
}
```

---

## Phase 5: Wire Up Routes

**File**: `api/cmd/services/ichor/build/all/all.go`

Add import:
```go
"github.com/timmaaaz/ichor/api/domain/http/workflow/workflowsaveapi"
```

In Routes() function, add:
```go
workflowsaveapi.Routes(app, workflowsaveapi.Config{
    Log:            cfg.Log,
    DB:             cfg.DB,
    WorkflowBus:    workflowBus,
    AuthClient:     cfg.AuthClient,
    PermissionsBus: permissionsBus,
})
```

---

## Phase 6: Save API Integration Tests

**Location**: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/`

### File Structure

```
api/cmd/services/ichor/tests/workflow/workflowsaveapi/
├── save_test.go       # Main orchestrator: Test_WorkflowSaveAPI()
├── seed_test.go       # SaveSeedData struct + insertSeedData()
├── create_test.go     # POST /workflow/rules/full tests
├── update_test.go     # PUT /workflow/rules/{id}/full tests
└── validation_test.go # Graph validation, action config validation
```

### Seed Data Structure

```go
type SaveSeedData struct {
    apitest.SeedData
    TriggerTypes []workflow.TriggerType
    EntityTypes  []workflow.EntityType
    Entities     []workflow.Entity

    // Pre-existing rules for update tests
    ExistingRule    workflow.AutomationRule
    ExistingActions []workflow.RuleAction
    ExistingEdges   []workflow.RuleEdge
}
```

**Seed Requirements**:
- Admin user with token + full permissions on `workflow.automation_rules`
- Regular user with token (for 401 tests)
- 3+ trigger types (on_create, on_update, on_delete)
- 2+ entity types
- 2+ entities
- 1 existing rule with 3 actions + 2 edges (for update tests)

### Test Cases

#### `create_test.go` - POST /workflow/rules/full

| Test Name | Status | Description |
|-----------|--------|-------------|
| `basic` | 200 | Create rule + 1 action + 1 start edge |
| `with-sequence` | 200 | Create rule + 3 actions in sequence (start → a1 → a2 → a3) |
| `with-branch` | 200 | Create rule with evaluate_condition branching |
| `with-canvas-layout` | 200 | Verify canvas_layout JSON saved and returned |
| `temp-id-resolution` | 200 | Verify `temp:0`, `temp:1` edges resolve to real UUIDs |
| `missing-name` | 400 | Validation error: name required |
| `missing-trigger-type` | 400 | Validation error: trigger_type_id required |
| `invalid-action-type` | 400 | Validation error: unknown action_type |
| `invalid-action-config` | 400 | Action config missing required fields |
| `graph-cycle` | 400 | Reject: action1 → action2 → action1 |
| `graph-unreachable` | 400 | Reject: action with no incoming edge |
| `no-auth` | 401 | Missing token |
| `wrong-user` | 401 | User without permissions |

#### `update_test.go` - PUT /workflow/rules/{id}/full

| Test Name | Status | Description |
|-----------|--------|-------------|
| `update-rule-only` | 200 | Change name/description, keep actions same |
| `add-action` | 200 | Add new action (id: null), verify created |
| `update-action` | 200 | Update existing action (id: uuid), verify updated |
| `delete-action` | 200 | Remove action from request, verify hard-deleted |
| `replace-edges` | 200 | All edges deleted and recreated |
| `update-canvas-layout` | 200 | Modify canvas_layout, verify persisted |
| `rule-not-found` | 404 | Invalid rule UUID |
| `invalid-action-id` | 400 | Action ID doesn't belong to this rule |

#### `validation_test.go` - Graph & Config Validation

| Test Name | Status | Description |
|-----------|--------|-------------|
| `create-alert-valid` | 200 | Valid create_alert config |
| `create-alert-missing-severity` | 400 | Missing required field |
| `send-email-valid` | 200 | Valid send_email config |
| `send-email-missing-recipients` | 400 | Missing required field |
| `evaluate-condition-valid` | 200 | Valid conditions array |
| `multi-start-edges` | 400 | Multiple start edges (only one allowed) |
| `no-start-edge` | 400 | No start edge defined |

### CmpFunc Patterns

Phase 6 tests use three distinct patterns based on what's being verified:

#### Pattern 1: Hybrid cmp.Diff for Create/Update Success (200)

For successful CREATE and UPDATE responses where we need to compare field values but must handle server-generated fields (IDs, timestamps):

```go
// create_test.go - basic, with-sequence, with-branch, with-canvas-layout
CmpFunc: func(got any, exp any) string {
    gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
    if !ok {
        return "failed to cast got response"
    }
    expResp, ok := exp.(*workflowsaveapp.SaveWorkflowResponse)
    if !ok {
        return "failed to cast exp response"
    }

    // Sync server-generated fields from got to exp
    expResp.ID = gotResp.ID
    expResp.CreatedDate = gotResp.CreatedDate
    expResp.UpdatedDate = gotResp.UpdatedDate

    // Sync action IDs (server-generated)
    for i := range gotResp.Actions {
        if i < len(expResp.Actions) {
            expResp.Actions[i].ID = gotResp.Actions[i].ID
        }
    }

    // Sync edge IDs and resolved action references
    for i := range gotResp.Edges {
        if i < len(expResp.Edges) {
            expResp.Edges[i].ID = gotResp.Edges[i].ID
            // temp:N references are resolved to real UUIDs
            expResp.Edges[i].SourceActionID = gotResp.Edges[i].SourceActionID
            expResp.Edges[i].TargetActionID = gotResp.Edges[i].TargetActionID
        }
    }

    return cmp.Diff(gotResp, expResp)
}
```

#### Pattern 2: Manual Checks for Temp ID Resolution

For tests that specifically verify temp ID resolution logic (not struct equality):

```go
// create_test.go - temp-id-resolution
CmpFunc: func(got any, exp any) string {
    gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
    if !ok {
        return "failed to cast response"
    }

    // Verify all IDs are assigned (not empty)
    if gotResp.ID == "" {
        return "rule ID should not be empty"
    }

    // Build map of action indices to their assigned UUIDs
    actionIDMap := make(map[int]string)
    for i, action := range gotResp.Actions {
        if action.ID == "" {
            return fmt.Sprintf("action[%d] ID should not be empty", i)
        }
        actionIDMap[i] = action.ID
    }

    // Verify edges have resolved UUIDs (no temp: prefix)
    for i, edge := range gotResp.Edges {
        if edge.ID == "" {
            return fmt.Sprintf("edge[%d] ID should not be empty", i)
        }
        if strings.HasPrefix(edge.TargetActionID, "temp:") {
            return fmt.Sprintf("edge[%d] target not resolved: %s", i, edge.TargetActionID)
        }
        if edge.SourceActionID != "" && strings.HasPrefix(edge.SourceActionID, "temp:") {
            return fmt.Sprintf("edge[%d] source not resolved: %s", i, edge.SourceActionID)
        }

        // Verify the resolved UUID matches an actual action
        if edge.TargetActionID != "" {
            found := false
            for _, actionID := range actionIDMap {
                if edge.TargetActionID == actionID {
                    found = true
                    break
                }
            }
            if !found {
                return fmt.Sprintf("edge[%d] target %s doesn't match any action", i, edge.TargetActionID)
            }
        }
    }

    return ""
}
```

#### Pattern 3: cmp.Diff for Error Responses (400/401/404)

For error responses, use `cmp.Diff` to compare the full error structure:

```go
// create_test.go - missing-name, invalid-action-type, graph-cycle, etc.
CmpFunc: func(got any, exp any) string {
    gotResp, ok := got.(*errs.Error)
    if !ok {
        return "failed to cast got to error"
    }
    expResp, ok := exp.(*errs.Error)
    if !ok {
        return "failed to cast exp to error"
    }

    return cmp.Diff(gotResp, expResp)
}
```

#### Pattern 4: Manual Checks for Update with Deletion Verification

For update tests that need to verify actions were deleted:

```go
// update_test.go - delete-action
CmpFunc: func(got any, exp any) string {
    gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
    if !ok {
        return "failed to cast response"
    }

    // Verify the deleted action is NOT in the response
    deletedActionID := sd.ExistingActions[2].ID.String() // action we removed from request
    for _, action := range gotResp.Actions {
        if action.ID == deletedActionID {
            return fmt.Sprintf("action %s should have been deleted but still exists", deletedActionID)
        }
    }

    // Verify remaining actions count
    expectedCount := len(sd.ExistingActions) - 1
    if len(gotResp.Actions) != expectedCount {
        return fmt.Sprintf("expected %d actions, got %d", expectedCount, len(gotResp.Actions))
    }

    return ""
}
```

### Example Test Tables

#### Create 200 Tests (Hybrid cmp.Diff)

```go
func create200(sd SaveSeedData) []apitest.Table {
    return []apitest.Table{
        {
            Name:       "basic",
            URL:        "/v1/workflow/rules/full",
            Token:      sd.Users[0].Token,
            Method:     http.MethodPost,
            StatusCode: http.StatusOK,
            Input: workflowsaveapp.SaveWorkflowRequest{
                Name:          "Test Workflow",
                Description:   "A test workflow",
                IsActive:      true,
                EntityID:      sd.Entities[0].ID.String(),
                TriggerTypeID: sd.TriggerTypes[0].ID.String(),
                Actions: []workflowsaveapp.SaveActionRequest{
                    {
                        Name:           "Create Alert",
                        ActionType:     "create_alert",
                        ExecutionOrder: 1,
                        IsActive:       true,
                        ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"Test","message":"Test message"}`),
                    },
                },
                Edges: []workflowsaveapp.SaveEdgeRequest{
                    {TargetActionID: "temp:0", EdgeType: "start"},
                },
            },
            GotResp: &workflowsaveapp.SaveWorkflowResponse{},
            ExpResp: &workflowsaveapp.SaveWorkflowResponse{
                Name:          "Test Workflow",
                Description:   "A test workflow",
                IsActive:      true,
                EntityID:      sd.Entities[0].ID.String(),
                TriggerTypeID: sd.TriggerTypes[0].ID.String(),
                Actions: []workflowsaveapp.SaveActionResponse{
                    {
                        Name:           "Create Alert",
                        ActionType:     "create_alert",
                        ExecutionOrder: 1,
                        IsActive:       true,
                        ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"Test","message":"Test message"}`),
                    },
                },
                Edges: []workflowsaveapp.SaveEdgeResponse{
                    {EdgeType: "start", EdgeOrder: 0},
                },
            },
            CmpFunc: func(got any, exp any) string {
                gotResp := got.(*workflowsaveapp.SaveWorkflowResponse)
                expResp := exp.(*workflowsaveapp.SaveWorkflowResponse)

                // Sync server-generated fields
                expResp.ID = gotResp.ID
                expResp.CreatedDate = gotResp.CreatedDate
                expResp.UpdatedDate = gotResp.UpdatedDate

                for i := range gotResp.Actions {
                    if i < len(expResp.Actions) {
                        expResp.Actions[i].ID = gotResp.Actions[i].ID
                    }
                }
                for i := range gotResp.Edges {
                    if i < len(expResp.Edges) {
                        expResp.Edges[i].ID = gotResp.Edges[i].ID
                        expResp.Edges[i].SourceActionID = gotResp.Edges[i].SourceActionID
                        expResp.Edges[i].TargetActionID = gotResp.Edges[i].TargetActionID
                    }
                }

                return cmp.Diff(gotResp, expResp)
            },
        },
    }
}
```

#### Create 400 Tests (cmp.Diff for Errors)

```go
func create400(sd SaveSeedData) []apitest.Table {
    return []apitest.Table{
        {
            Name:       "missing-name",
            URL:        "/v1/workflow/rules/full",
            Token:      sd.Users[0].Token,
            Method:     http.MethodPost,
            StatusCode: http.StatusBadRequest,
            Input: workflowsaveapp.SaveWorkflowRequest{
                Name:          "", // Missing required field
                EntityID:      sd.Entities[0].ID.String(),
                TriggerTypeID: sd.TriggerTypes[0].ID.String(),
                Actions: []workflowsaveapp.SaveActionRequest{
                    {Name: "Action", ActionType: "create_alert", ExecutionOrder: 1, IsActive: true,
                        ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
                },
                Edges: []workflowsaveapp.SaveEdgeRequest{{TargetActionID: "temp:0", EdgeType: "start"}},
            },
            GotResp: &errs.Error{},
            ExpResp: &errs.Error{Code: errs.InvalidArgument, Message: "validate: name is required"},
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
        {
            Name:       "graph-cycle",
            URL:        "/v1/workflow/rules/full",
            Token:      sd.Users[0].Token,
            Method:     http.MethodPost,
            StatusCode: http.StatusBadRequest,
            Input: workflowsaveapp.SaveWorkflowRequest{
                Name:          "Cycle Test",
                EntityID:      sd.Entities[0].ID.String(),
                TriggerTypeID: sd.TriggerTypes[0].ID.String(),
                Actions: []workflowsaveapp.SaveActionRequest{
                    {Name: "Action1", ActionType: "create_alert", ExecutionOrder: 1, IsActive: true,
                        ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
                    {Name: "Action2", ActionType: "create_alert", ExecutionOrder: 2, IsActive: true,
                        ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
                },
                Edges: []workflowsaveapp.SaveEdgeRequest{
                    {TargetActionID: "temp:0", EdgeType: "start"},
                    {SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence"},
                    {SourceActionID: "temp:1", TargetActionID: "temp:0", EdgeType: "sequence"}, // Creates cycle
                },
            },
            GotResp: &errs.Error{},
            ExpResp: &errs.Error{Code: errs.InvalidArgument, Message: "graph: cycle detected in workflow graph"},
            CmpFunc: func(got any, exp any) string {
                return cmp.Diff(got, exp)
            },
        },
    }
}
```

### Required Imports for Phase 6 Tests

```go
import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "testing"

    "github.com/google/go-cmp/cmp"
    "github.com/timmaaaz/ichor/api/sdk/http/apitest"
    "github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
    "github.com/timmaaaz/ichor/app/sdk/errs"
)
```

### Test Pattern Summary by Test Case

| Test File | Test Name | Pattern | Reason |
|-----------|-----------|---------|--------|
| `create_test.go` | `basic` | Hybrid cmp.Diff | Compare all fields, sync IDs/timestamps |
| `create_test.go` | `with-sequence` | Hybrid cmp.Diff | Compare all fields, sync IDs/timestamps |
| `create_test.go` | `with-branch` | Hybrid cmp.Diff | Compare all fields, sync IDs/timestamps |
| `create_test.go` | `with-canvas-layout` | Hybrid cmp.Diff | Compare all fields, sync IDs/timestamps |
| `create_test.go` | `temp-id-resolution` | Manual Check | Verify temp→UUID resolution logic specifically |
| `create_test.go` | `missing-name` | cmp.Diff | Error struct comparison |
| `create_test.go` | `missing-trigger-type` | cmp.Diff | Error struct comparison |
| `create_test.go` | `invalid-action-type` | cmp.Diff | Error struct comparison |
| `create_test.go` | `invalid-action-config` | cmp.Diff | Error struct comparison |
| `create_test.go` | `graph-cycle` | cmp.Diff | Error struct comparison |
| `create_test.go` | `graph-unreachable` | cmp.Diff | Error struct comparison |
| `create_test.go` | `no-auth` | cmp.Diff | Error struct comparison |
| `create_test.go` | `wrong-user` | cmp.Diff | Error struct comparison |
| `update_test.go` | `update-rule-only` | Hybrid cmp.Diff | Compare all fields, sync timestamps |
| `update_test.go` | `add-action` | Hybrid cmp.Diff | Compare all fields, sync new action ID |
| `update_test.go` | `update-action` | Hybrid cmp.Diff | Compare all fields |
| `update_test.go` | `delete-action` | Manual Check | Verify deleted action is absent |
| `update_test.go` | `replace-edges` | Hybrid cmp.Diff | Compare all fields, sync edge IDs |
| `update_test.go` | `update-canvas-layout` | Hybrid cmp.Diff | Compare all fields |
| `update_test.go` | `rule-not-found` | cmp.Diff | Error struct comparison |
| `update_test.go` | `invalid-action-id` | cmp.Diff | Error struct comparison |
| `validation_test.go` | `create-alert-valid` | Hybrid cmp.Diff | Valid response comparison |
| `validation_test.go` | `create-alert-missing-severity` | cmp.Diff | Error struct comparison |
| `validation_test.go` | `send-email-valid` | Hybrid cmp.Diff | Valid response comparison |
| `validation_test.go` | `send-email-missing-recipients` | cmp.Diff | Error struct comparison |
| `validation_test.go` | `evaluate-condition-valid` | Hybrid cmp.Diff | Valid response comparison |
| `validation_test.go` | `multi-start-edges` | cmp.Diff | Error struct comparison |
| `validation_test.go` | `no-start-edge` | cmp.Diff | Error struct comparison |

### Transaction Rollback Test

```go
// Test that partial failure rolls back everything
{
    Name:       "rollback-on-edge-failure",
    URL:        "/v1/workflow/rules/full",
    Token:      sd.Users[0].Token,
    StatusCode: http.StatusBadRequest,
    Method:     http.MethodPost,
    Input: workflowsaveapp.SaveWorkflowRequest{
        Name:          "Rollback Test",
        EntityID:      sd.Entities[0].ID.String(),
        TriggerTypeID: sd.TriggerTypes[0].ID.String(),
        Actions: []workflowsaveapp.SaveActionRequest{
            {Name: "Action", ActionType: "create_alert", ExecutionOrder: 1, IsActive: true,
                ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
        },
        Edges: []workflowsaveapp.SaveEdgeRequest{
            {TargetActionID: "temp:999", EdgeType: "start"}, // Invalid temp ID - will fail
        },
    },
    GotResp: &errs.Error{},
    ExpResp: &errs.Error{Code: errs.InvalidArgument, Message: "edge references invalid action index: temp:999"},
    CmpFunc: func(got any, exp any) string {
        return cmp.Diff(got, exp)
    },
}
```

---

## API Contract

### Request Format
```json
{
  "name": "Rule Name",
  "description": "Optional",
  "is_active": true,
  "entity_id": "uuid",
  "trigger_type_id": "uuid",
  "trigger_conditions": {"field_conditions": [...]},
  "actions": [
    {"id": "existing-uuid", "name": "...", "action_type": "send_email", "action_config": {...}, "execution_order": 1},
    {"id": null, "name": "...", "action_type": "create_alert", "action_config": {...}, "execution_order": 2}
  ],
  "edges": [
    {"source_action_id": null, "target_action_id": "temp:0", "edge_type": "start"},
    {"source_action_id": "temp:0", "target_action_id": "temp:1", "edge_type": "sequence"}
  ],
  "canvas_layout": {"viewport": {...}, "node_positions": {...}}
}
```

### Key Rules
- Actions with `null` ID = new (will be created)
- Actions with UUID = existing (will be updated)
- Missing existing actions = deleted
- Edges use `temp:N` to reference new actions by array index
- Start edges have `null` source_action_id

---

## Trigger System Reference

### Trigger Types (Pre-defined Lookup Data)

The system has 4 pre-seeded trigger types. Users select one by ID:

| Name | Description | Use Case |
|------|-------------|----------|
| `on_create` | Fired when entity is created | New order notifications |
| `on_update` | Fired when entity is updated | Status change alerts |
| `on_delete` | Fired when entity is deleted | Archive notifications |
| `scheduled` | Fired on a schedule | Daily reports, reminders |

**Frontend**: Query `/v1/workflow/trigger-types` to get the list of available trigger types with their UUIDs.

### trigger_conditions Schema

The `trigger_conditions` JSONB defines **when** a rule should fire within the selected trigger type.

**Structure**:
```json
{
  "field_conditions": [
    {
      "field_name": "status",
      "operator": "changed_to",
      "value": "shipped",
      "previous_value": "processing"
    }
  ]
}
```

**Supported Operators**:

| Operator | Description | Example |
|----------|-------------|---------|
| `equals` | Field equals value | `{"field_name": "priority", "operator": "equals", "value": "high"}` |
| `not_equals` | Field doesn't equal value | `{"field_name": "status", "operator": "not_equals", "value": "cancelled"}` |
| `changed_to` | Field changed to specific value | `{"field_name": "status", "operator": "changed_to", "value": "shipped"}` |
| `changed_from` | Field changed from specific value | `{"field_name": "status", "operator": "changed_from", "value": "draft"}` |
| `greater_than` | Numeric comparison | `{"field_name": "amount", "operator": "greater_than", "value": 1000}` |
| `less_than` | Numeric comparison | `{"field_name": "quantity", "operator": "less_than", "value": 10}` |
| `contains` | String contains | `{"field_name": "name", "operator": "contains", "value": "urgent"}` |
| `in` | Value in list | `{"field_name": "category", "operator": "in", "value": ["A", "B", "C"]}` |

**Behavior**:
- **Empty/null conditions**: Rule fires on ALL events of that trigger type
- **Multiple conditions**: AND logic (all conditions must match)
- **Change detection**: `changed_to` and `changed_from` compare old vs new values

**Example: Order Shipped Notification**
```json
{
  "trigger_type_id": "uuid-of-on_update",
  "trigger_conditions": {
    "field_conditions": [
      {
        "field_name": "status",
        "operator": "changed_to",
        "value": "shipped"
      },
      {
        "field_name": "is_priority",
        "operator": "equals",
        "value": true
      }
    ]
  }
}
```
This rule fires when: an order is updated AND status changed to "shipped" AND is_priority is true.

**Example: Fire on Any Create**
```json
{
  "trigger_type_id": "uuid-of-on_create",
  "trigger_conditions": null
}
```
This rule fires on every create event for the entity.

---

## Files to Create

| File | Purpose |
|------|---------|
| `app/domain/workflow/workflowsaveapp/model.go` | Request/Response models |
| `app/domain/workflow/workflowsaveapp/validation.go` | Action config validation |
| `app/domain/workflow/workflowsaveapp/graph.go` | Cycle/reachability checks |
| `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` | Main logic |
| `api/domain/http/workflow/workflowsaveapi/workflowsaveapi.go` | HTTP handlers |
| `api/domain/http/workflow/workflowsaveapi/route.go` | Routes |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/save_test.go` | Test orchestrator |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/seed_test.go` | Test seed data |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/create_test.go` | Create tests |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/update_test.go` | Update tests |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/validation_test.go` | Validation tests |

## Files to Modify

### Schema & Business Layer (Phase 1-2)

| File | Change |
|------|--------|
| `business/sdk/migrate/sql/migrate.sql` | Modify CREATE TABLE to add `canvas_layout` column |
| `business/sdk/workflow/models.go` | Add `CanvasLayout` to 4 structs |
| `business/sdk/workflow/stores/workflowdb/models.go` | Add DB field + update 3 converters |
| `business/sdk/workflow/stores/workflowdb/workflowdb.go` | Update 7 SQL functions |

### Existing API Layer (Phase 2)

| File | Change |
|------|--------|
| `api/domain/http/workflow/ruleapi/model.go` | Add `CanvasLayout` to 3 structs |
| `api/domain/http/workflow/ruleapi/converters.go` | Update 3 converter functions |

### Tests (Phase 2)

| File | Change |
|------|--------|
| `api/cmd/services/ichor/tests/workflow/ruleapi/seed_test.go` | Add `CanvasLayout: nil` |
| `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_seed_test.go` | Add `CanvasLayout: nil` to ~6 instantiations |
| `api/cmd/services/ichor/tests/workflow/ruleapi/create_test.go` | Add field to test requests |
| `api/cmd/services/ichor/tests/workflow/ruleapi/update_test.go` | Add field to test requests |

### New Save API (Phases 3-5)

| File | Change |
|------|--------|
| `api/cmd/services/ichor/build/all/all.go` | Wire new routes |

---

## Verification

```bash
# Run all workflow tests
make test

# Run specific test suites
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...
go test -v ./api/cmd/services/ichor/tests/workflow/ruleapi/...

# Manual test
curl -X PUT http://localhost:3000/v1/workflow/rules/{id}/full \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","entity_id":"...","actions":[...]}'
```

---

## Phase 7: Workflow Execution Integration Tests

**Location**: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_test.go`

### Purpose

Verify that workflows created via the Save API actually execute correctly when their triggers fire.

### Test Infrastructure - Use Existing Utilities

The codebase has comprehensive workflow testing infrastructure. Use these existing utilities:

| Utility | Location | Purpose |
|---------|----------|---------|
| `apitest.InitWorkflowInfra(t, db)` | `api/sdk/http/apitest/workflow.go` | Returns `*WorkflowInfra` with Engine, QueueManager, WorkflowBus, Client |
| `workflow.TestSeedFullWorkflow()` | `business/sdk/workflow/testutil.go` | Seeds trigger types, entity types, rules, actions |
| `workflow.NewEventPublisher()` | `business/sdk/workflow/eventpublisher.go` | Creates event publisher |
| `workflow.NewDelegateHandler()` | `business/sdk/workflow/delegatehandler.go` | Bridges domain events to workflow |
| `qm.GetMetrics()` | `business/sdk/workflow/queue.go` | Track TotalEnqueued, TotalProcessed, TotalFailed |
| `engine.GetExecutionHistory(n)` | `business/sdk/workflow/engine.go` | Get recent workflow executions |
| `workflow.ResetEngineForTesting()` | `business/sdk/workflow/engine.go` | Reset singleton for test isolation |
| `qm.ResetMetrics()` | `business/sdk/workflow/queue.go` | Clear metrics between tests |

### Seed Setup Pattern

```go
// ExecutionTestData uses existing infrastructure
type ExecutionTestData struct {
    SaveSeedData
    WF *apitest.WorkflowInfra  // Engine, QueueManager, WorkflowBus, Client

    // Created workflows for testing (via Save API)
    SimpleWorkflow    SaveWorkflowResponse
    SequenceWorkflow  SaveWorkflowResponse
    BranchingWorkflow SaveWorkflowResponse
    ParallelWorkflow  SaveWorkflowResponse
}

func insertExecutionSeedData(t *testing.T, test *apitest.Test) ExecutionTestData {
    ctx := context.Background()

    // 1. Get base save seed data
    sd := insertSeedData(t, test)

    // 2. Use existing workflow infrastructure helper
    wf := apitest.InitWorkflowInfra(t, test.DB)

    // 3. Seed workflow lookup data (trigger types, entity types)
    workflowData, err := workflow.TestSeedFullWorkflow(ctx, sd.Users[0].ID, wf.WorkflowBus)
    if err != nil {
        t.Fatalf("seed workflow data: %v", err)
    }

    // 4. Re-initialize engine to pick up seeded data
    wf.Engine.Initialize(ctx, wf.WorkflowBus)

    // 5. Reset for clean test state
    workflow.ResetEngineForTesting()
    wf.QueueManager.ResetMetrics()
    wf.QueueManager.ResetCircuitBreaker()

    // 6. Create test workflows via the Save API
    simpleWorkflow := createSimpleWorkflow(t, test, sd, workflowData)
    sequenceWorkflow := createSequenceWorkflow(t, test, sd, workflowData)
    branchingWorkflow := createBranchingWorkflow(t, test, sd, workflowData)
    parallelWorkflow := createParallelWorkflow(t, test, sd, workflowData)

    // 7. Re-initialize engine after creating workflows
    wf.Engine.Initialize(ctx, wf.WorkflowBus)

    return ExecutionTestData{
        SaveSeedData:      sd,
        WF:                wf,
        SimpleWorkflow:    simpleWorkflow,
        SequenceWorkflow:  sequenceWorkflow,
        BranchingWorkflow: branchingWorkflow,
        ParallelWorkflow:  parallelWorkflow,
    }
}
```

### Helper: Wait for Async Processing

Use this pattern from existing tests (`api/cmd/services/ichor/tests/sales/ordersapi/workflow_test.go`):

```go
func waitForProcessing(t *testing.T, qm *workflow.QueueManager, initialProcessed int64, timeout time.Duration) bool {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        metrics := qm.GetMetrics()
        if metrics.TotalProcessed > initialProcessed {
            return true
        }
        time.Sleep(50 * time.Millisecond)
    }
    return false
}
```

### Test Cases

#### 7a. Single Action Execution

| Test Name | Description | Verification |
|-----------|-------------|--------------|
| `execute_single_create_alert` | Fire trigger → 1 action (create_alert) | Alert created in workflow.alerts |
| `execute_single_update_field` | Fire trigger → 1 action (update_field) | Target entity field updated |
| `execute_single_send_email` | Fire trigger → 1 action (send_email) | Email queued/mock verified |

**Test Pattern**:

```go
func executeSingleCreateAlert(sd ExecutionTestData) []apitest.Table {
    return []apitest.Table{
        {
            Name: "single-action-create-alert",
            // This is a custom execution test, not HTTP test
            // Use CmpFunc to orchestrate and verify
            CmpFunc: func(got, exp any) string {
                ctx := context.Background()

                // 1. Create trigger event
                event := workflow.TriggerEvent{
                    EventType:  "on_create",
                    EntityName: sd.Entities[0].TableName,
                    EntityID:   uuid.New(),
                    Timestamp:  time.Now(),
                    RawData:    map[string]any{"status": "new"},
                    UserID:     sd.Users[0].ID,
                }

                // 2. Get initial metrics
                initialMetrics := sd.WF.QueueManager.GetMetrics()

                // 3. Execute workflow via engine
                execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
                if err != nil {
                    return fmt.Sprintf("execution failed: %v", err)
                }

                // 4. Verify execution completed
                if execution.Status != workflow.ExecutionStatusCompleted {
                    return fmt.Sprintf("expected completed, got %s", execution.Status)
                }

                // 5. Verify action executed
                if len(execution.BatchResults) == 0 {
                    return "no batch results"
                }
                if len(execution.BatchResults[0].RuleResults) == 0 {
                    return "no rule results"
                }
                ruleResult := execution.BatchResults[0].RuleResults[0]
                if len(ruleResult.ActionResults) == 0 {
                    return "no action results"
                }
                if ruleResult.ActionResults[0].Status != "success" {
                    return fmt.Sprintf("action failed: %s", ruleResult.ActionResults[0].ErrorMessage)
                }

                // 6. Verify alert created in database
                ruleID, _ := uuid.Parse(sd.SimpleWorkflow.ID)
                alerts, err := sd.WF.WorkflowBus.QueryAlerts(ctx, workflow.AlertFilter{
                    SourceRuleID: &ruleID,
                })
                if err != nil {
                    return fmt.Sprintf("query alerts: %v", err)
                }
                if len(alerts) == 0 {
                    return "alert not created in database"
                }

                // 7. Verify execution history
                history := sd.WF.Engine.GetExecutionHistory(10)
                if len(history) == 0 {
                    return "no execution in history"
                }

                return ""
            },
        },
    }
}
```

#### 7b. Sequence Execution

| Test Name | Description | Verification |
|-----------|-------------|--------------|
| `execute_sequence_3_actions` | Fire trigger → action1 → action2 → action3 | All 3 actions executed in order |
| `execute_sequence_partial_failure` | action2 fails | action1 succeeded, action2 failed, action3 skipped |
| `execute_sequence_first_failure` | action1 fails | action1 failed, action2+3 skipped, status=failed |

**Test Pattern**:

```go
func executeSequence3Actions(sd ExecutionTestData) []apitest.Table {
    return []apitest.Table{
        {
            Name: "sequence-3-actions",
            CmpFunc: func(got, exp any) string {
                ctx := context.Background()

                event := workflow.TriggerEvent{
                    EventType:  "on_update",
                    EntityName: sd.Entities[0].TableName,
                    EntityID:   uuid.New(),
                    Timestamp:  time.Now(),
                    RawData:    map[string]any{"status": "processing"},
                    UserID:     sd.Users[0].ID,
                }

                execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
                if err != nil {
                    return fmt.Sprintf("execution failed: %v", err)
                }

                // Verify all 3 actions executed
                ruleResult := execution.BatchResults[0].RuleResults[0]
                if len(ruleResult.ActionResults) != 3 {
                    return fmt.Sprintf("expected 3 actions, got %d", len(ruleResult.ActionResults))
                }

                // Verify execution order
                for i, ar := range ruleResult.ActionResults {
                    if ar.Status != "success" {
                        return fmt.Sprintf("action[%d] failed: %s", i, ar.ErrorMessage)
                    }
                }

                return ""
            },
        },
    }
}
```

#### 7c. Graph-Based (Branching) Execution

| Test Name | Description | Verification |
|-----------|-------------|--------------|
| `execute_branch_true` | Condition evaluates true → true_branch | Only true_branch action executed |
| `execute_branch_false` | Condition evaluates false → false_branch | Only false_branch action executed |
| `execute_branch_convergent` | Both branches converge to common action | Common action executed once |
| `execute_multi_condition` | Nested conditions | Correct branch path taken |

**Test Pattern**:

```go
func executeBranchTrue(sd ExecutionTestData) []apitest.Table {
    return []apitest.Table{
        {
            Name: "branch-true-path",
            CmpFunc: func(got, exp any) string {
                ctx := context.Background()

                // Event data that will make condition evaluate to TRUE
                event := workflow.TriggerEvent{
                    EventType:  "on_update",
                    EntityName: sd.Entities[0].TableName,
                    EntityID:   uuid.New(),
                    Timestamp:  time.Now(),
                    RawData: map[string]any{
                        "amount": 1500, // condition: amount > 1000
                        "status": "approved",
                    },
                    UserID: sd.Users[0].ID,
                }

                execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
                if err != nil {
                    return fmt.Sprintf("execution failed: %v", err)
                }

                // Find the evaluate_condition action result
                var conditionResult *workflow.ActionResult
                var trueBranchResult *workflow.ActionResult
                var falseBranchResult *workflow.ActionResult

                for _, br := range execution.BatchResults {
                    for _, rr := range br.RuleResults {
                        for i := range rr.ActionResults {
                            ar := &rr.ActionResults[i]
                            switch ar.ActionType {
                            case "evaluate_condition":
                                conditionResult = ar
                            case "create_alert":
                                // Assuming true_branch creates alert
                                trueBranchResult = ar
                            case "send_email":
                                // Assuming false_branch sends email
                                falseBranchResult = ar
                            }
                        }
                    }
                }

                // Verify condition evaluated
                if conditionResult == nil {
                    return "condition action not found"
                }
                if conditionResult.BranchTaken != "true_branch" {
                    return fmt.Sprintf("expected true_branch, got %s", conditionResult.BranchTaken)
                }

                // Verify true branch executed
                if trueBranchResult == nil || trueBranchResult.Status != "success" {
                    return "true branch action did not execute"
                }

                // Verify false branch NOT executed (should be skipped or nil)
                if falseBranchResult != nil && falseBranchResult.Status == "success" {
                    return "false branch should not have executed"
                }

                return ""
            },
        },
    }
}
```

#### 7d. Parallel Execution

| Test Name | Description | Verification |
|-----------|-------------|--------------|
| `execute_parallel_2_actions` | Same execution_order → parallel | Both actions executed (verify timing) |
| `execute_parallel_one_fails` | Parallel actions, one fails | Other completes, status=partial |

#### 7e. Execution History Verification

| Test Name | Description | Verification |
|-----------|-------------|--------------|
| `execution_record_created` | Any execution | Record in automation_executions |
| `execution_trigger_data_stored` | Any execution | trigger_data JSONB contains event |
| `execution_actions_stored` | Any execution | actions_executed JSONB contains results |
| `execution_time_tracked` | Any execution | execution_time_ms > 0 |

**Test Pattern**:

```go
func executionRecordCreated(sd ExecutionTestData) []apitest.Table {
    return []apitest.Table{
        {
            Name: "execution-record-created",
            CmpFunc: func(got, exp any) string {
                ctx := context.Background()

                event := workflow.TriggerEvent{
                    EventType:  "on_create",
                    EntityName: sd.Entities[0].TableName,
                    EntityID:   uuid.New(),
                    Timestamp:  time.Now(),
                    RawData:    map[string]any{},
                    UserID:     sd.Users[0].ID,
                }

                execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
                if err != nil {
                    return fmt.Sprintf("execution failed: %v", err)
                }

                // Query execution record from database
                execRecord, err := sd.WF.WorkflowBus.QueryExecutionByID(ctx, execution.ExecutionID)
                if err != nil {
                    return fmt.Sprintf("execution not found in DB: %v", err)
                }

                // Verify fields
                if execRecord.Status != "success" && execRecord.Status != "completed" {
                    return fmt.Sprintf("expected success status, got %s", execRecord.Status)
                }
                if len(execRecord.TriggerData) == 0 {
                    return "trigger_data should not be empty"
                }
                if execRecord.ExecutionTimeMs == 0 {
                    return "execution_time_ms should be > 0"
                }

                // Also verify via GetExecutionHistory
                history := sd.WF.Engine.GetExecutionHistory(10)
                found := false
                for _, h := range history {
                    if h.ExecutionID == execution.ExecutionID {
                        found = true
                        break
                    }
                }
                if !found {
                    return "execution not in GetExecutionHistory"
                }

                return ""
            },
        },
    }
}
```

---

## Phase 8: End-to-End Trigger Integration Tests

**Location**: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go`

### Purpose

Test complete scenarios where real entity CRUD operations trigger workflow execution through the delegate/event system.

### Test Infrastructure - Using Existing Patterns

Reference: `api/cmd/services/ichor/tests/sales/ordersapi/workflow_test.go` (TestWorkflow_OrdersDelegateEvents)

```go
type TriggerTestData struct {
    ExecutionTestData

    // Domain business layers (from test.DB.BusDomain)
    OrdersBus    *ordersbus.Business
    CustomersBus *customersbus.Business

    // Event bridge
    EventPublisher  *workflow.EventPublisher
    DelegateHandler *workflow.DelegateHandler
    Delegate        *delegate.Delegate

    // Test entities
    TestCustomer customersbus.Customer
}
```

### Seed Setup

```go
func insertTriggerSeedData(t *testing.T, test *apitest.Test) TriggerTestData {
    ctx := context.Background()

    // 1. Get execution seed data (includes WorkflowInfra)
    sd := insertExecutionSeedData(t, test)

    // 2. Create delegate for domain event registration
    del := delegate.New(test.DB.Log)

    // 3. Create event publisher and delegate handler
    eventPublisher := workflow.NewEventPublisher(test.DB.Log, sd.WF.QueueManager)
    delegateHandler := workflow.NewDelegateHandler(test.DB.Log, eventPublisher)

    // 4. Register domains for event bridging
    delegateHandler.RegisterDomain(del, ordersbus.DomainName, "sales.orders")
    delegateHandler.RegisterDomain(del, customersbus.DomainName, "sales.customers")

    // 5. Create domain business layers with the delegate
    ordersBus := ordersbus.NewBusiness(test.DB.Log, del, ordersdb.NewStore(test.DB.Log, test.DB.DB))
    customersBus := customersbus.NewBusiness(test.DB.Log, del, customersdb.NewStore(test.DB.Log, test.DB.DB))

    // 6. Seed test customer
    testCustomer, err := customersBus.Create(ctx, customersbus.NewCustomer{
        Name:  "Test Customer",
        Email: "test@example.com",
    })
    if err != nil {
        t.Fatalf("create test customer: %v", err)
    }

    // 7. Start queue manager to process events
    sd.WF.QueueManager.Start(ctx)

    return TriggerTestData{
        ExecutionTestData: sd,
        OrdersBus:         ordersBus,
        CustomersBus:      customersBus,
        EventPublisher:    eventPublisher,
        DelegateHandler:   delegateHandler,
        Delegate:          del,
        TestCustomer:      testCustomer,
    }
}
```

### Test Cases

#### 8a. Entity Create Triggers

| Test Name | Trigger | Workflow | Verification |
|-----------|---------|----------|--------------|
| `customer_create_alert` | Create customer | on_create customer → create_alert | Alert created with customer info |
| `order_create_notification` | Create order | on_create order → send_notification | Notification queued |
| `product_create_audit` | Create product | on_create product → update_field (audit log) | Audit record created |

**Test Pattern**:

```go
func customerCreateAlert(sd TriggerTestData) []apitest.Table {
    return []apitest.Table{
        {
            Name: "customer-create-triggers-alert",
            CmpFunc: func(got, exp any) string {
                ctx := context.Background()

                // 1. First, create a workflow that triggers on customer creation
                workflowReq := workflowsaveapp.SaveWorkflowRequest{
                    Name:          "Customer Create Alert",
                    EntityID:      sd.Entities[0].ID.String(), // customers entity
                    TriggerTypeID: sd.TriggerTypes[0].ID.String(), // on_create
                    IsActive:      true,
                    Actions: []workflowsaveapp.SaveActionRequest{
                        {
                            Name:           "Create Welcome Alert",
                            ActionType:     "create_alert",
                            ExecutionOrder: 1,
                            IsActive:       true,
                            ActionConfig: json.RawMessage(`{
                                "alert_type": "customer_welcome",
                                "severity": "info",
                                "title": "New Customer Created",
                                "message": "Customer {{customer_name}} has been created"
                            }`),
                        },
                    },
                    Edges: []workflowsaveapp.SaveEdgeRequest{
                        {TargetActionID: "temp:0", EdgeType: "start"},
                    },
                }

                // Create workflow via API
                resp, err := createWorkflowViaAPI(ctx, sd, workflowReq)
                if err != nil {
                    return fmt.Sprintf("create workflow: %v", err)
                }

                // Re-initialize engine to pick up new workflow
                sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus)

                // 2. Get initial metrics
                initialMetrics := sd.WF.QueueManager.GetMetrics()

                // 3. Create a customer (this should trigger the workflow via delegate)
                newCustomer := customersbus.NewCustomer{
                    Name:  "Test Customer",
                    Email: "test@example.com",
                }
                customer, err := sd.CustomersBus.Create(ctx, newCustomer)
                if err != nil {
                    return fmt.Sprintf("create customer: %v", err)
                }

                // 4. Wait for async workflow execution
                if !waitForProcessing(t, sd.WF.QueueManager, initialMetrics.TotalProcessed, 5*time.Second) {
                    return "workflow did not process in time"
                }

                // 5. Verify alert was created
                alerts, err := sd.WF.WorkflowBus.QueryAlerts(ctx, workflow.AlertFilter{
                    SourceEntityID: &customer.ID,
                })
                if err != nil {
                    return fmt.Sprintf("query alerts: %v", err)
                }
                if len(alerts) == 0 {
                    return "no alert created for customer"
                }

                alert := alerts[0]
                if alert.AlertType != "customer_welcome" {
                    return fmt.Sprintf("wrong alert type: %s", alert.AlertType)
                }
                if !strings.Contains(alert.Message, "Test Customer") {
                    return "alert message should contain customer name"
                }

                // 6. Verify execution record via GetExecutionHistory
                history := sd.WF.Engine.GetExecutionHistory(10)
                if len(history) == 0 {
                    return "no execution record created"
                }

                return ""
            },
        },
    }
}
```

#### 8b. Entity Update Triggers with Conditions

| Test Name | Trigger | Condition | Workflow | Verification |
|-----------|---------|-----------|----------|--------------|
| `order_status_shipped` | Update order | status changed_to shipped | send_email to customer | Email queued with order details |
| `order_amount_threshold` | Update order | amount greater_than 1000 | seek_approval | Approval request created |
| `inventory_low_stock` | Update inventory | quantity less_than 10 | create_alert | Low stock alert created |
| `order_priority_flag` | Update order | is_priority equals true | send_notification | Priority notification sent |

**Test Pattern**:

```go
func orderStatusShipped(sd TriggerTestData) []apitest.Table {
    return []apitest.Table{
        {
            Name: "order-status-shipped-triggers-email",
            CmpFunc: func(got, exp any) string {
                ctx := context.Background()

                // 1. Create workflow with condition
                workflowReq := workflowsaveapp.SaveWorkflowRequest{
                    Name:          "Order Shipped Email",
                    EntityID:      sd.OrdersEntity.ID.String(),
                    TriggerTypeID: sd.TriggerTypes[1].ID.String(), // on_update
                    TriggerConditions: json.RawMessage(`{
                        "field_conditions": [{
                            "field_name": "status",
                            "operator": "changed_to",
                            "value": "shipped"
                        }]
                    }`),
                    IsActive: true,
                    Actions: []workflowsaveapp.SaveActionRequest{
                        {
                            Name:           "Send Shipped Email",
                            ActionType:     "send_email",
                            ExecutionOrder: 1,
                            IsActive:       true,
                            ActionConfig: json.RawMessage(`{
                                "recipients": ["{{customer_email}}"],
                                "subject": "Your order has shipped!",
                                "body": "Order #{{order_id}} is on its way."
                            }`),
                        },
                    },
                    Edges: []workflowsaveapp.SaveEdgeRequest{
                        {TargetActionID: "temp:0", EdgeType: "start"},
                    },
                }

                resp, err := createWorkflowViaAPI(ctx, sd, workflowReq)
                if err != nil {
                    return fmt.Sprintf("create workflow: %v", err)
                }

                // Re-initialize engine to pick up new workflow
                sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus)

                // 2. Create an order with status = "processing"
                order, err := sd.OrdersBus.Create(ctx, ordersbus.NewOrder{
                    CustomerID: sd.TestCustomer.ID,
                    Status:     "processing",
                    Amount:     decimal.NewFromFloat(100.00),
                })
                if err != nil {
                    return fmt.Sprintf("create order: %v", err)
                }

                // 3. Get metrics before update
                initialMetrics := sd.WF.QueueManager.GetMetrics()

                // 4. Update order status to "shipped" (this should trigger)
                _, err = sd.OrdersBus.Update(ctx, order, ordersbus.UpdateOrder{
                    Status: ptr("shipped"),
                })
                if err != nil {
                    return fmt.Sprintf("update order: %v", err)
                }

                // 5. Wait for async execution
                if !waitForProcessing(t, sd.WF.QueueManager, initialMetrics.TotalProcessed, 5*time.Second) {
                    return "workflow did not process in time"
                }

                // 6. Verify execution triggered via history
                history := sd.WF.Engine.GetExecutionHistory(10)
                if len(history) == 0 {
                    return "workflow not triggered on status change"
                }

                // 7. Verify email action was successful (check action results)
                latestExec := history[0]
                foundEmailSuccess := false
                for _, br := range latestExec.BatchResults {
                    for _, rr := range br.RuleResults {
                        for _, ar := range rr.ActionResults {
                            if ar.ActionType == "send_email" && ar.Status == "success" {
                                foundEmailSuccess = true
                            }
                        }
                    }
                }
                if !foundEmailSuccess {
                    return "email action did not succeed"
                }

                return ""
            },
        },
    }
}
```

#### 8c. Condition NOT Matching (Negative Tests)

| Test Name | Trigger | Condition | Verification |
|-----------|---------|-----------|--------------|
| `order_status_wrong_value` | Update order status to "cancelled" | status changed_to shipped | Workflow NOT triggered |
| `amount_below_threshold` | Update order with amount=500 | amount greater_than 1000 | Workflow NOT triggered |
| `no_field_change` | Update order (no status change) | status changed_to shipped | Workflow NOT triggered |

**Test Pattern**:

```go
func orderStatusWrongValue(sd TriggerTestData) []apitest.Table {
    return []apitest.Table{
        {
            Name: "condition-not-matched-no-trigger",
            CmpFunc: func(got, exp any) string {
                ctx := context.Background()

                // Create workflow with shipped condition (same as previous test)
                // ... workflow creation code ...

                // Re-initialize engine
                sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus)

                // Create order
                order, _ := sd.OrdersBus.Create(ctx, ordersbus.NewOrder{
                    CustomerID: sd.TestCustomer.ID,
                    Status:     "processing",
                })

                // Get initial metrics and history count
                initialMetrics := sd.WF.QueueManager.GetMetrics()
                initialHistory := len(sd.WF.Engine.GetExecutionHistory(100))

                // Update order to "cancelled" (NOT shipped - should NOT trigger)
                _, err := sd.OrdersBus.Update(ctx, order, ordersbus.UpdateOrder{
                    Status: ptr("cancelled"),
                })
                if err != nil {
                    return fmt.Sprintf("update order: %v", err)
                }

                // Wait a bit for any potential processing
                time.Sleep(500 * time.Millisecond)

                // Verify NO new execution created
                finalHistory := len(sd.WF.Engine.GetExecutionHistory(100))
                if finalHistory > initialHistory {
                    return "workflow should NOT have triggered for cancelled status"
                }

                // Also verify metrics didn't increase (or only enqueued, not processed)
                finalMetrics := sd.WF.QueueManager.GetMetrics()
                if finalMetrics.TotalProcessed > initialMetrics.TotalProcessed {
                    // Check if any processed event was for our rule
                    history := sd.WF.Engine.GetExecutionHistory(10)
                    for _, h := range history {
                        // If we find an execution that matched our rule, fail
                        for _, br := range h.BatchResults {
                            for _, rr := range br.RuleResults {
                                if rr.RuleID == workflowID {
                                    return "workflow executed when condition should not match"
                                }
                            }
                        }
                    }
                }

                return ""
            },
        },
    }
}
```

#### 8d. Entity Delete Triggers

| Test Name | Trigger | Workflow | Verification |
|-----------|---------|----------|--------------|
| `customer_delete_archive` | Delete customer | on_delete customer → update_field (archive) | Archive record created |
| `order_delete_notify` | Delete order | on_delete order → send_notification | Notification sent |

#### 8e. Inactive Rule Tests

| Test Name | Description | Verification |
|-----------|-------------|--------------|
| `inactive_rule_no_trigger` | Create entity, workflow is_active=false | Workflow NOT triggered |
| `reactivate_rule_triggers` | Set is_active=true, then create entity | Workflow triggered |

---

## Phase 9: Action-Specific Integration Tests

**Location**: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go`

### Purpose

Verify each action type executes correctly with proper configuration and produces expected side effects.

### Test Cases by Action Type

#### 9a. create_alert Action

| Test Name | Config | Verification |
|-----------|--------|--------------|
| `create_alert_basic` | type, severity, title, message | Alert in workflow.alerts |
| `create_alert_with_recipients` | + user recipients | alert_recipients populated |
| `create_alert_template_vars` | message with {{entity_id}} | Variables substituted |
| `create_alert_severity_levels` | info, warning, error, critical | Correct severity stored |

#### 9b. update_field Action

| Test Name | Config | Verification |
|-----------|--------|--------------|
| `update_field_string` | target_field=status, value="updated" | Entity field updated in DB |
| `update_field_numeric` | target_field=priority, value=5 | Numeric field updated |
| `update_field_template` | value="Processed by {{user_name}}" | Template resolved |
| `update_field_cascade` | Update triggers another workflow | Downstream workflow fired |

#### 9c. send_email Action

| Test Name | Config | Verification |
|-----------|--------|--------------|
| `send_email_basic` | recipients, subject, body | Email handler invoked (mock) |
| `send_email_multiple_recipients` | 3 recipients | All recipients processed |
| `send_email_template_body` | body with {{order_total}} | Variables substituted |

#### 9d. evaluate_condition Action

| Test Name | Config | Verification |
|-----------|--------|--------------|
| `condition_equals_true` | field equals value | Returns "true_branch" |
| `condition_equals_false` | field equals other_value | Returns "false_branch" |
| `condition_greater_than` | amount > 1000 | Correct branch |
| `condition_multiple_and` | 2 conditions (AND logic) | Both must match |
| `condition_nested` | Condition → condition → action | Correct path through |

#### 9e. seek_approval Action

| Test Name | Config | Verification |
|-----------|--------|--------------|
| `seek_approval_basic` | approvers, approval_type | Approval request created |
| `seek_approval_with_timeout` | + timeout_hours | Timeout tracked |

#### 9f. allocate_inventory Action (Async)

| Test Name | Config | Verification |
|-----------|--------|--------------|
| `allocate_basic` | inventory_items, allocation_mode | Queued to RabbitMQ |
| `allocate_result_stored` | After processing | allocation_results populated |

---

## Phase 10: Error Handling & Edge Case Tests

**Location**: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go`

### Purpose

Verify proper error handling, rollback behavior, and edge cases.

### Test Cases

#### 10a. Action Failures

| Test Name | Scenario | Verification |
|-----------|----------|--------------|
| `action_fails_sequence_stops` | Action 2 of 3 fails | Action 3 skipped, status=failed |
| `action_fails_partial_rollback` | Transactional action fails | Prior changes rolled back |
| `action_timeout` | Action exceeds timeout | Marked as failed with timeout error |
| `invalid_action_config_runtime` | Config valid at save, fails at runtime | Error logged, execution fails |

#### 10b. Trigger Condition Errors

| Test Name | Scenario | Verification |
|-----------|----------|--------------|
| `condition_field_not_found` | Condition references non-existent field | Rule skipped (not error) |
| `condition_type_mismatch` | Numeric operator on string field | Graceful handling |

#### 10c. Concurrency & Race Conditions

| Test Name | Scenario | Verification |
|-----------|----------|--------------|
| `concurrent_triggers_same_rule` | 10 entities created simultaneously | All executions recorded, no duplicates |
| `rule_updated_during_execution` | Rule modified while executing | Original version used |

#### 10d. Queue Failures

| Test Name | Scenario | Verification |
|-----------|----------|--------------|
| `queue_unavailable` | RabbitMQ down | Event published returns error, doesn't block |
| `queue_retry_success` | First attempt fails, retry succeeds | Eventually processed |

---

## CmpFunc Pattern Summary for Phases 7-10

Unlike Phase 6 (HTTP CRUD tests) which use `cmp.Diff` for response comparison, Phases 7-10 use **manual checks** because they test **behavior and side effects**, not struct equality.

| Phase | Test Type | Pattern | Reason |
|-------|-----------|---------|--------|
| 7 | Execution | Manual Check | Verifies workflow executed, actions ran, alerts created in DB |
| 8 | E2E Triggers | Manual Check | Verifies async events fired, workflows triggered by entity CRUD |
| 9 | Action-Specific | Manual Check | Verifies side effects (alerts, emails, field updates) |
| 10 | Error Handling | Manual Check | Verifies failure states, rollbacks, retry behavior |

**Key difference:**
- **Phase 6**: "Did the API return the correct response struct?" → Use `cmp.Diff`
- **Phases 7-10**: "Did the workflow engine execute correctly and produce the right side effects?" → Use manual checks

All Phase 7-10 tests return `""` on success or a descriptive error string on failure. They do NOT compare response structs - they verify behavior by:
1. Triggering workflows via engine or entity operations
2. Querying database for expected records (alerts, executions)
3. Checking metrics and execution history
4. Verifying action results and branch paths

---

## Files to Create (Testing Phases)

| File | Purpose |
|------|---------|
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_test.go` | Phase 7 execution tests |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_seed_test.go` | Phase 7 seed data |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go` | Phase 8 E2E trigger tests |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_seed_test.go` | Phase 8 seed data |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go` | Phase 9 action-specific tests |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go` | Phase 10 error handling tests |

---

## Test Orchestrator Update

**File**: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/save_test.go`

```go
func Test_WorkflowSaveAPI(t *testing.T) {
    test := apitest.StartTest(t, "workflowsaveapi_test")

    // Phase 6: Save API CRUD Tests (existing)
    sd := insertSeedData(t, test)
    test.Run(t, create200(sd), "create-200")
    test.Run(t, create400(sd), "create-400")
    test.Run(t, update200(sd), "update-200")
    test.Run(t, update400(sd), "update-400")
    test.Run(t, validation400(sd), "validation-400")

    // Phase 7: Execution Tests
    esd := insertExecutionSeedData(t, test)
    test.Run(t, executeSingleCreateAlert(esd), "exec-single-alert")
    test.Run(t, executeSequence3Actions(esd), "exec-sequence")
    test.Run(t, executeBranchTrue(esd), "exec-branch-true")
    test.Run(t, executeBranchFalse(esd), "exec-branch-false")
    test.Run(t, executionRecordCreated(esd), "exec-record")

    // Phase 8: End-to-End Trigger Tests
    tsd := insertTriggerSeedData(t, test)
    test.Run(t, customerCreateAlert(tsd), "trigger-customer-create")
    test.Run(t, orderStatusShipped(tsd), "trigger-order-shipped")
    test.Run(t, orderStatusWrongValue(tsd), "trigger-no-match")
    test.Run(t, inactiveRuleNoTrigger(tsd), "trigger-inactive")

    // Phase 9: Action-Specific Tests
    test.Run(t, createAlertBasic(esd), "action-alert")
    test.Run(t, updateFieldString(esd), "action-update-field")
    test.Run(t, evaluateConditionTrue(esd), "action-condition")

    // Phase 10: Error Handling Tests
    test.Run(t, actionFailsSequenceStops(esd), "error-sequence-stop")
    test.Run(t, concurrentTriggersSameRule(tsd), "error-concurrent")
}
```

---

## Verification Commands

```bash
# Run all workflow save API tests
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...

# Run specific test phases
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -run "exec-"      # Phase 7
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -run "trigger-"   # Phase 8
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -run "action-"    # Phase 9
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -run "error-"     # Phase 10

# Run with race detector
go test -race -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...

# Run with verbose logging
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -args -test.v
```

---

## Dependencies

These testing phases require:

1. **Phase 1-6 completed** (schema, business layer, app layer, API layer, wiring, basic tests)
2. **Existing test infrastructure** - all handled by `apitest.InitWorkflowInfra()`
3. **Test entities seeded** for each domain being tested (customers, orders, etc.)

---

## Existing Infrastructure (Already Implemented)

All infrastructure is handled by existing utilities:

| What | How | Location |
|------|-----|----------|
| RabbitMQ container | `rabbitmq.GetTestContainer(t)` | `foundation/rabbitmq/rabbitmq.go` |
| Test RabbitMQ client | `rabbitmq.NewTestClient(url)` | `foundation/rabbitmq/rabbitmq.go` |
| Test workflow queue | `rabbitmq.NewTestWorkflowQueue(client, log)` | `foundation/rabbitmq/rabbitmq.go` |
| Complete workflow stack | `apitest.InitWorkflowInfra(t, db)` | `api/sdk/http/apitest/workflow.go` |
| Workflow seeding | `workflow.TestSeedFullWorkflow()` | `business/sdk/workflow/testutil.go` |
| Database + all business layers | `dbtest.NewDatabase(t, name)` | `business/sdk/dbtest/dbtest.go` |

**Key point**: `apitest.InitWorkflowInfra()` handles all of:
- Getting shared RabbitMQ container
- Creating test client and queue
- Creating workflow business layer
- Initializing engine with action handlers (SendEmail, SendNotification, CreateAlert)
- Creating queue manager
- Registering cleanup via `t.Cleanup()`

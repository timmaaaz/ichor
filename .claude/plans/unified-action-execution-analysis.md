# Unified Action Execution Analysis

## User Goal

Any button that "does stuff" (non-CRUD operations like inventory allocation, approval requests, etc.) should use the **same action execution mechanism** regardless of whether:
- A **user clicks a button** manually, or
- The **workflow engine** executes it automatically via an automation rule

This ensures consistency, auditability, and allows the workflow system to automate anything a user can do manually.

---

## Current Architecture Assessment

### Does the current architecture support this principle?

**Short Answer: Partially, but with a gap.**

The current architecture has a **well-designed action handler interface** that is decoupled from the automation trigger system. However, there is **no built-in mechanism for manual action execution**—actions can only be invoked through the workflow engine when automation rules match events.

---

## Architecture Deep Dive

### What Works Well

#### 1. ActionHandler Interface (Supports Unified Execution)

**Location**: [interfaces.go](business/sdk/workflow/interfaces.go)

```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string
}
```

This interface is **completely decoupled** from the automation system:
- Takes a generic `config` (JSON configuration)
- Takes an `ActionExecutionContext` (execution metadata)
- Returns `any` for flexibility

This means the same handler could theoretically be called from:
- Workflow engine (current path)
- Manual API endpoint (missing path)

#### 2. ActionRegistry (Supports Discovery)

**Location**: [interfaces.go:44-74](business/sdk/workflow/interfaces.go#L44-L74)

```go
type ActionRegistry struct {
    handlers map[string]ActionHandler
}

func (ar *ActionRegistry) Get(actionType string) (ActionHandler, bool)
func (ar *ActionRegistry) GetAll() []string
```

The registry already supports:
- Looking up handlers by type
- Listing all available action types

This could be exposed for manual execution.

#### 3. Existing Handlers (Ready for Dual Use)

All 6 registered handlers implement the same interface:

| Action Type | Handler | Manual Execution Support |
|------------|---------|-------------------------|
| `allocate_inventory` | AllocateInventoryHandler | Yes |
| `create_alert` | CreateAlertHandler | Yes |
| `send_email` | SendEmailHandler | Yes |
| `send_notification` | SendNotificationHandler | Yes |
| `seek_approval` | SeekApprovalHandler | Yes |
| `update_field` | UpdateFieldHandler | **No** (use entity CRUD) |

---

### What's Missing (The Gap)

#### Current Execution Flow (Automation-Only)

```
Business Entity Change (Create/Update/Delete)
       │
       ▼
Delegate fires event
       │
       ▼
EventPublisher queues to RabbitMQ
       │
       ▼
QueueManager consumer picks up
       │
       ▼
WorkflowEngine.ExecuteWorkflow()
       │
       ▼
TriggerProcessor.ProcessEvent() ← Finds matching rules
       │
       ▼
ActionExecutor.ExecuteRuleActions() ← Executes handler
       │
       ▼
handler.Execute(ctx, config, context)
```

**The gap**: There is no path from "User clicks button" → `handler.Execute()`.

#### What Would Be Needed

```
User clicks "Allocate Inventory" button
       │
       ▼
POST /v1/workflow/actions/{actionType}/execute
       │
       ▼
actionapi.execute() ← NEW
       │
       ▼
actionapp.Execute() ← NEW
       │
       ▼
ActionService.Execute() ← NEW (or direct registry access)
       │
       ▼
handler.Execute(ctx, config, context) ← SAME handler as automation
```

---

## Recommendation

**Option C (Hybrid — ActionService Layer)** is the best balance of:
- Code reuse
- Consistency
- Auditability
- Minimal disruption to existing code

---

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Permission model | Role-based action permissions (new table) | Clean semantics, constraint support, auditability |
| Async support | Full async support | Long-running operations like allocation need tracking |
| Constraints | Stub for later implementation | Design schema with JSONB constraints column, implement validation incrementally |
| Default permissions | Admin gets all, others get none | Secure by default, explicit grants required |
| Discovery API | Yes, `GET /v1/workflow/actions` | Enables dynamic UI button generation |
| `update_field` manual support | **Excluded** | Use entity CRUD endpoints directly instead |

---

## Final Implementation Plan

### Phase 1: Schema Changes

Since this is not in production, **modify table definitions directly** rather than using ALTER TABLE.

#### 1.1 Modify automation_executions (v1.67)

**File**: `business/sdk/migrate/sql/migrate.sql`

**Replace existing v1.67 definition with:**
```sql
-- Version: 1.67
-- Description: Create workflow automation_executions table
CREATE TABLE workflow.automation_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_rules_id UUID REFERENCES workflow.automation_rules(id),  -- NOW NULLABLE for manual
    entity_type VARCHAR(50) NOT NULL,
    trigger_data JSONB,
    actions_executed JSONB,
    status VARCHAR(20),
    error_message TEXT,
    execution_time_ms INTEGER,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    trigger_source VARCHAR(20) DEFAULT 'automation',  -- NEW: 'automation' or 'manual'
    executed_by UUID REFERENCES core.users(id),       -- NEW: user who triggered
    action_type VARCHAR(100)                          -- NEW: for manual executions
);

CREATE INDEX idx_automation_executions_trigger_source ON workflow.automation_executions(trigger_source);
CREATE INDEX idx_automation_executions_executed_by ON workflow.automation_executions(executed_by);
```

#### 1.2 Add action_permissions Table (v1.990)

**Note**: Version numbers use three decimal places (1.990, 1.991) to avoid conflicts with Darwin migration's float-based version comparison where 1.99 would conflict with 1.9 and 1.100 would conflict with 1.10.

```sql
-- Version: 1.990
-- Description: Create workflow action_permissions table for manual action authorization
CREATE TABLE workflow.action_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES core.roles(id) ON DELETE CASCADE,
    action_type VARCHAR(100) NOT NULL,
    is_allowed BOOLEAN DEFAULT TRUE,
    constraints JSONB DEFAULT '{}',  -- Stubbed for future implementation
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, action_type)
);

CREATE INDEX idx_action_permissions_role ON workflow.action_permissions(role_id);
CREATE INDEX idx_action_permissions_type ON workflow.action_permissions(action_type);
```

#### 1.3 Seed Admin Permissions (v1.991)

```sql
-- Version: 1.991
-- Description: Seed default action permissions for admin role
INSERT INTO workflow.action_permissions (role_id, action_type, is_allowed)
SELECT r.id, action_type, true
FROM core.roles r
CROSS JOIN (VALUES
    ('allocate_inventory'),
    ('create_alert'),
    ('send_email'),
    ('send_notification'),
    ('seek_approval')
) AS actions(action_type)
WHERE r.name = 'admin';
```

**Note**: `update_field` is intentionally excluded from manual execution permissions.

---

### Phase 2: Core Infrastructure

#### 2.1 Update ActionExecutionContext

**File**: `business/sdk/workflow/models.go`

```go
type ActionExecutionContext struct {
    EntityID      uuid.UUID              `json:"entity_id,omitempty"`
    EntityName    string                 `json:"entity_name"`
    EventType     string                 `json:"event_type"`  // "on_create", "on_update", "on_delete", "manual_trigger"
    FieldChanges  map[string]FieldChange `json:"field_changes,omitempty"`
    RawData       map[string]interface{} `json:"raw_data,omitempty"`
    Timestamp     time.Time              `json:"timestamp"`
    UserID        uuid.UUID              `json:"user_id,omitempty"`
    RuleID        *uuid.UUID             `json:"rule_id,omitempty"`     // CHANGED: Now pointer (nil for manual)
    RuleName      string                 `json:"rule_name"`
    ExecutionID   uuid.UUID              `json:"execution_id,omitempty"`
    TriggerSource string                 `json:"trigger_source"`        // NEW: "automation" or "manual"
}
```

#### 2.2 Extend ActionHandler Interface

**File**: `business/sdk/workflow/interfaces.go`

```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string

    // NEW: Capability declarations
    SupportsManualExecution() bool  // false for update_field
    IsAsync() bool                  // true for allocate_inventory, send_email
    GetDescription() string         // Human-readable for discovery API
}
```

#### 2.3 Update All Handlers

| Handler File | SupportsManual | IsAsync | Description |
|--------------|----------------|---------|-------------|
| `workflowactions/inventory/allocate.go` | `true` | `true` | "Allocate inventory items from warehouse locations" |
| `workflowactions/communication/alert.go` | `true` | `false` | "Create an alert notification for users or roles" |
| `workflowactions/communication/email.go` | `true` | `true` | "Send an email to specified recipients" |
| `workflowactions/communication/notification.go` | `true` | `false` | "Send an in-app notification" |
| `workflowactions/approval/seek.go` | `true` | `false` | "Request approval from specified approvers" |
| `workflowactions/data/updatefield.go` | **`false`** | `false` | "Update a field value on an entity" |

#### 2.4 Create ActionService

**File**: `business/sdk/workflow/actionservice.go`

```go
type ActionService struct {
    log      *logger.Logger
    registry *ActionRegistry
    db       *sqlx.DB
}

func NewActionService(log *logger.Logger, db *sqlx.DB, registry *ActionRegistry) *ActionService

func (s *ActionService) Execute(ctx context.Context, req ExecuteRequest) (*ExecuteResult, error)
func (s *ActionService) GetExecutionStatus(ctx context.Context, executionID uuid.UUID) (*ExecutionStatus, error)
func (s *ActionService) ListAvailableActions() []ActionInfo
```

#### 2.5 Refactor ActionExecutor

**File**: `business/sdk/workflow/executor.go`

- Delegate to `ActionService.Execute()` internally
- Maintain backward compatibility for automation flow
- Set `TriggerSource: "automation"` for rule-triggered executions

---

### Phase 3: actionpermissionsbus Domain

**Directory**: `business/domain/workflow/actionpermissionsbus/`

#### 3.1 model.go

```go
package actionpermissionsbus

type ActionPermission struct {
    ID          uuid.UUID
    RoleID      uuid.UUID
    ActionType  string
    IsAllowed   bool
    Constraints json.RawMessage  // Stubbed for future use
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type NewActionPermission struct {
    RoleID      uuid.UUID
    ActionType  string
    IsAllowed   bool
    Constraints json.RawMessage
}

type UpdateActionPermission struct {
    IsAllowed   *bool
    Constraints *json.RawMessage
}
```

#### 3.2 filter.go

```go
package actionpermissionsbus

type QueryFilter struct {
    ID         *uuid.UUID
    RoleID     *uuid.UUID
    ActionType *string
    IsAllowed  *bool
}
```

#### 3.3 order.go

```go
package actionpermissionsbus

var DefaultOrderBy = order.NewBy(OrderByActionType, order.ASC)

const (
    OrderByID         = "id"
    OrderByRoleID     = "role_id"
    OrderByActionType = "action_type"
    OrderByCreatedAt  = "created_at"
)
```

#### 3.4 actionpermissionsbus.go

```go
package actionpermissionsbus

var (
    ErrNotFound = errors.New("action permission not found")
    ErrUnique   = errors.New("action permission already exists for role")
)

type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Create(ctx context.Context, ap ActionPermission) error
    Update(ctx context.Context, ap ActionPermission) error
    Delete(ctx context.Context, ap ActionPermission) error
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ActionPermission, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
    QueryByID(ctx context.Context, id uuid.UUID) (ActionPermission, error)
    QueryByRoleAndAction(ctx context.Context, roleID uuid.UUID, actionType string) (ActionPermission, error)
}

type Business struct {
    log    *logger.Logger
    storer Storer
}

func NewBusiness(log *logger.Logger, storer Storer) *Business

// Key method for permission checking
func (b *Business) CanUserExecuteAction(ctx context.Context, userID uuid.UUID, actionType string, userRoles []uuid.UUID) (bool, error)
```

#### 3.5 Database Store

**Directory**: `business/domain/workflow/actionpermissionsbus/stores/actionpermissionsdb/`

| File | Purpose |
|------|---------|
| `model.go` | DB model mapping (use `dbActionPermission` to avoid conflicts) |
| `filter.go` | Query filter implementation |
| `order.go` | Order by implementation |
| `actionpermissionsdb.go` | Database operations |

---

### Phase 4: App Layer (actionapp)

**Directory**: `app/domain/workflow/actionapp/`

#### 4.1 model.go

```go
package actionapp

// QueryParams for action listing
type QueryParams struct {
    Page    string
    Rows    string
    OrderBy string
}

// ExecuteRequest for manual action execution
type ExecuteRequest struct {
    Config     json.RawMessage        `json:"config" validate:"required"`
    EntityID   *string                `json:"entityId,omitempty"`
    EntityName string                 `json:"entityName,omitempty"`
    RawData    map[string]interface{} `json:"rawData,omitempty"`
}

func (app *ExecuteRequest) Decode(data []byte) error
func (app ExecuteRequest) Validate() error

// ExecuteResponse for action execution result
type ExecuteResponse struct {
    ExecutionID string      `json:"executionId"`
    Status      string      `json:"status"` // "completed", "queued", "failed"
    Result      interface{} `json:"result,omitempty"`
    TrackingURL string      `json:"trackingUrl,omitempty"`
    Error       string      `json:"error,omitempty"`
}

func (app ExecuteResponse) Encode() ([]byte, string, error)

// AvailableAction for discovery endpoint
type AvailableAction struct {
    Type         string `json:"type"`
    Description  string `json:"description"`
    IsAsync      bool   `json:"isAsync"`
}

// AvailableActions collection wrapper (implements Encoder)
type AvailableActions []AvailableAction

func (app AvailableActions) Encode() ([]byte, string, error)

// ExecutionStatus for tracking endpoint
type ExecutionStatus struct {
    ExecutionID string      `json:"executionId"`
    ActionType  string      `json:"actionType"`
    Status      string      `json:"status"`
    Result      interface{} `json:"result,omitempty"`
    Error       string      `json:"error,omitempty"`
    StartedAt   string      `json:"startedAt"`
    CompletedAt string      `json:"completedAt,omitempty"`
}

func (app ExecutionStatus) Encode() ([]byte, string, error)
```

#### 4.2 actionapp.go

```go
package actionapp

type App struct {
    actionService *workflow.ActionService
    actionPermBus *actionpermissionsbus.Business
    userBus       *userbus.Business
}

func NewApp(actionService *workflow.ActionService, actionPermBus *actionpermissionsbus.Business, userBus *userbus.Business) *App

func (a *App) Execute(ctx context.Context, actionType string, req ExecuteRequest) (ExecuteResponse, error)
func (a *App) ListAvailable(ctx context.Context, userID uuid.UUID) (AvailableActions, error)
func (a *App) GetExecutionStatus(ctx context.Context, executionID uuid.UUID) (ExecutionStatus, error)
```

---

### Phase 5: API Layer (actionapi)

**Directory**: `api/domain/http/workflow/actionapi/`

#### 5.1 route.go

```go
package actionapi

const RouteTable = "workflow_actions"

type Config struct {
    Log            *logger.Logger
    ActionService  *workflow.ActionService
    ActionPermBus  *actionpermissionsbus.Business
    UserBus        *userbus.Business
    AuthClient     *authclient.Client
    PermissionsBus *permissionsbus.Business
}

func Routes(app *web.App, cfg Config) {
    const version = "v1"

    api := newAPI(actionapp.NewApp(cfg.ActionService, cfg.ActionPermBus, cfg.UserBus))
    authen := mid.Authenticate(cfg.AuthClient)

    // List available actions (user sees only permitted actions)
    app.HandlerFunc(http.MethodGet, version, "/workflow/actions", api.list, authen)

    // Execute an action manually
    app.HandlerFunc(http.MethodPost, version, "/workflow/actions/{actionType}/execute", api.execute, authen)

    // Get execution status (for async tracking)
    app.HandlerFunc(http.MethodGet, version, "/workflow/executions/{executionId}", api.getExecutionStatus, authen)
}
```

#### 5.2 actionapi.go

```go
package actionapi

type api struct {
    actionApp *actionapp.App
}

func newAPI(actionApp *actionapp.App) *api

func (api *api) list(ctx context.Context, r *http.Request) web.Encoder
func (api *api) execute(ctx context.Context, r *http.Request) web.Encoder
func (api *api) getExecutionStatus(ctx context.Context, r *http.Request) web.Encoder
```

---

### Phase 6: Wiring in all.go

**File**: `api/cmd/services/ichor/build/all/all.go`

```go
// Imports to add
import (
    "github.com/timmaaaz/ichor/api/domain/http/workflow/actionapi"
    "github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
    "github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus/stores/actionpermissionsdb"
)

// In Routes() function, add:

// Action permissions business layer
actionPermBus := actionpermissionsbus.NewBusiness(cfg.Log,
    actionpermissionsdb.NewStore(cfg.Log, cfg.DB))

// Action service (wraps ActionExecutor)
actionService := workflow.NewActionService(cfg.Log, cfg.DB, actionExecutor.GetRegistry())

// Wire action API routes
actionapi.Routes(app, actionapi.Config{
    Log:            cfg.Log,
    ActionService:  actionService,
    ActionPermBus:  actionPermBus,
    UserBus:        userBus,
    AuthClient:     cfg.AuthClient,
    PermissionsBus: permissionsBus,
})
```

---

## Testing Plan

### Unit Tests

#### ActionService Tests
**File**: `business/sdk/workflow/actionservice_test.go`

| Test Case | Description |
|-----------|-------------|
| `TestActionService_Execute_Manual` | Execute action with `TriggerSource: "manual"` |
| `TestActionService_Execute_Automation` | Execute action with `TriggerSource: "automation"` |
| `TestActionService_Execute_UnknownActionType` | Returns error for unregistered action type |
| `TestActionService_Execute_ValidationFailure` | Handler.Validate() failure returns proper error |
| `TestActionService_Execute_HandlerError` | Handler.Execute() failure is recorded and returned |
| `TestActionService_Execute_NilRuleID` | Manual execution with nil RuleID succeeds |
| `TestActionService_RecordExecution` | Execution is recorded to database |

#### ActionHandler Interface Tests
**Location**: `business/sdk/workflow/workflowactions/{category}/{handler}_test.go`

Each handler needs tests for the **new interface methods**:

| Handler | Tests Required |
|---------|---------------|
| `allocate_inventory` | `TestAllocateInventoryHandler_SupportsManualExecution`, `TestAllocateInventoryHandler_IsAsync`, `TestAllocateInventoryHandler_GetDescription` |
| `create_alert` | `TestCreateAlertHandler_SupportsManualExecution`, `TestCreateAlertHandler_IsAsync`, `TestCreateAlertHandler_GetDescription` |
| `send_email` | `TestSendEmailHandler_SupportsManualExecution`, `TestSendEmailHandler_IsAsync`, `TestSendEmailHandler_GetDescription` |
| `send_notification` | `TestSendNotificationHandler_SupportsManualExecution`, `TestSendNotificationHandler_IsAsync`, `TestSendNotificationHandler_GetDescription` |
| `seek_approval` | `TestSeekApprovalHandler_SupportsManualExecution`, `TestSeekApprovalHandler_IsAsync`, `TestSeekApprovalHandler_GetDescription` |
| `update_field` | `TestUpdateFieldHandler_SupportsManualExecution` (must return `false`) |

#### ActionExecutionContext Tests
**File**: `business/sdk/workflow/models_test.go`

| Test Case | Description |
|-----------|-------------|
| `TestActionExecutionContext_ManualTrigger` | Context with `EventType: "manual_trigger"` is valid |
| `TestActionExecutionContext_NilRuleID` | Context with nil `RuleID` is valid for manual triggers |
| `TestActionExecutionContext_TriggerSource` | `TriggerSource` field correctly distinguishes automation vs manual |

#### ActionPermissions Business Layer Tests
**File**: `business/domain/workflow/actionpermissionsbus/actionpermissionsbus_test.go`

| Test Case | Description |
|-----------|-------------|
| `TestCanUserExecuteAction_Allowed` | User with role permission can execute |
| `TestCanUserExecuteAction_Denied` | User without permission gets denied |
| `TestCanUserExecuteAction_MultiRole` | User with multiple roles, any role grants permission |
| `TestCanUserExecuteAction_Constraints` | Constraint validation (stub test for future) |
| `TestActionPermissions_CRUD` | Create, Read, Update, Delete operations |

---

### Integration Tests

**Directory**: `api/cmd/services/ichor/tests/workflow/actionapi/`

#### Test Files

| File | Purpose |
|------|---------|
| `action_test.go` | Main test orchestrator |
| `seed_test.go` | Seed data for action execution tests |
| `execute_test.go` | Manual action execution tests |
| `list_test.go` | Action discovery endpoint tests |
| `tracking_test.go` | Async execution tracking tests |

#### Execute Endpoint Tests (`execute_test.go`)

| Test Case | HTTP Status | Description |
|-----------|-------------|-------------|
| `execute200_sync` | 200 | Successful sync action execution (create_alert) |
| `execute202_async` | 202 | Successful async action execution (allocate_inventory) |
| `execute400_invalid_config` | 400 | Invalid action configuration |
| `execute401_unauthenticated` | 401 | No auth token |
| `execute403_no_permission` | 403 | User lacks action permission |
| `execute404_unknown_action` | 404 | Action type not registered |
| `execute404_update_field` | 404 | `update_field` excluded from manual (returns not found or forbidden) |
| `execute422_validation_error` | 422 | Handler validation fails |

#### List Endpoint Tests (`list_test.go`)

| Test Case | HTTP Status | Description |
|-----------|-------------|-------------|
| `list200_admin` | 200 | Admin sees all manual-executable actions |
| `list200_user` | 200 | User sees only permitted actions |
| `list200_empty` | 200 | User with no permissions sees empty list |
| `list200_excludes_update_field` | 200 | `update_field` never appears in list |
| `list401_unauthenticated` | 401 | No auth token |

#### Tracking Endpoint Tests (`tracking_test.go`)

| Test Case | HTTP Status | Description |
|-----------|-------------|-------------|
| `tracking200_completed` | 200 | Get completed execution result |
| `tracking200_processing` | 200 | Get in-progress execution status |
| `tracking404_not_found` | 404 | Unknown execution ID |
| `tracking401_unauthenticated` | 401 | No auth token |

#### Seed Data Requirements (`seed_test.go`)

```go
type SeedData struct {
    // Users
    AdminUser    userbus.User      // Has all action permissions
    WarehouseMgr userbus.User      // Has allocate_inventory permission
    SalesRep     userbus.User      // Has create_alert permission
    BasicUser    userbus.User      // Has no action permissions

    // Roles
    AdminRole        rolebus.Role
    WarehouseMgrRole rolebus.Role
    SalesRepRole     rolebus.Role
    BasicRole        rolebus.Role

    // Action Permissions
    AdminPermissions        []actionpermissionsbus.ActionPermission
    WarehouseMgrPermissions []actionpermissionsbus.ActionPermission
    SalesRepPermissions     []actionpermissionsbus.ActionPermission

    // Test Entities (for action execution context)
    TestOrder       orderbus.Order
    TestInventory   inventorybus.InventoryItem
}
```

---

### Test Infrastructure Updates

#### Table Access for Tests
**File**: `business/domain/core/tableaccessbus/testutil.go`

Add entry for new table:
```go
{RoleID: uuid.Nil, TableName: "action_permissions", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
```

#### Action Permissions Test Utility
**New File**: `business/domain/workflow/actionpermissionsbus/testutil.go`

```go
package actionpermissionsbus

// TestSeedActionPermissions seeds action permissions for testing
func TestSeedActionPermissions(ctx context.Context, n int, api *Business, roleID uuid.UUID, actionTypes []string) ([]ActionPermission, error)
```

---

## Files Summary

### New Files to Create

| File | Purpose |
|------|---------|
| `business/sdk/workflow/actionservice.go` | Core execution service |
| `business/sdk/workflow/actionservice_test.go` | ActionService unit tests |
| `business/sdk/workflow/models_test.go` | ActionExecutionContext tests |
| `business/domain/workflow/actionpermissionsbus/model.go` | Permission model |
| `business/domain/workflow/actionpermissionsbus/actionpermissionsbus.go` | Permission business logic |
| `business/domain/workflow/actionpermissionsbus/actionpermissionsbus_test.go` | Permission unit tests |
| `business/domain/workflow/actionpermissionsbus/filter.go` | Query filters |
| `business/domain/workflow/actionpermissionsbus/order.go` | Order by definitions |
| `business/domain/workflow/actionpermissionsbus/testutil.go` | Test seed utility |
| `business/domain/workflow/actionpermissionsbus/stores/actionpermissionsdb/model.go` | DB model |
| `business/domain/workflow/actionpermissionsbus/stores/actionpermissionsdb/filter.go` | DB filter |
| `business/domain/workflow/actionpermissionsbus/stores/actionpermissionsdb/order.go` | DB order |
| `business/domain/workflow/actionpermissionsbus/stores/actionpermissionsdb/actionpermissionsdb.go` | DB store |
| `app/domain/workflow/actionapp/actionapp.go` | App layer |
| `app/domain/workflow/actionapp/model.go` | Request/response models |
| `api/domain/http/workflow/actionapi/actionapi.go` | HTTP handlers |
| `api/domain/http/workflow/actionapi/route.go` | Route registration |
| `api/cmd/services/ichor/tests/workflow/actionapi/action_test.go` | Integration test orchestrator |
| `api/cmd/services/ichor/tests/workflow/actionapi/seed_test.go` | Integration test seeding |
| `api/cmd/services/ichor/tests/workflow/actionapi/execute_test.go` | Execution tests |
| `api/cmd/services/ichor/tests/workflow/actionapi/list_test.go` | List endpoint tests |
| `api/cmd/services/ichor/tests/workflow/actionapi/tracking_test.go` | Tracking tests |

### Files to Modify

| File | Changes |
|------|---------|
| `business/sdk/workflow/interfaces.go` | Extend ActionHandler interface with 3 new methods |
| `business/sdk/workflow/models.go` | Update ActionExecutionContext (TriggerSource, optional RuleID) |
| `business/sdk/workflow/executor.go` | Delegate to ActionService, set TriggerSource |
| `business/sdk/workflow/workflowactions/communication/alert.go` | Add new interface methods |
| `business/sdk/workflow/workflowactions/communication/alert_test.go` | Add tests for new methods |
| `business/sdk/workflow/workflowactions/communication/email.go` | Add new interface methods |
| `business/sdk/workflow/workflowactions/communication/notification.go` | Add new interface methods |
| `business/sdk/workflow/workflowactions/approval/seek.go` | Add new interface methods |
| `business/sdk/workflow/workflowactions/data/updatefield.go` | Add new methods (SupportsManual returns false) |
| `business/sdk/workflow/workflowactions/data/updatefield_test.go` | Add test for SupportsManualExecution=false |
| `business/sdk/workflow/workflowactions/inventory/allocate.go` | Add new interface methods |
| `business/sdk/migrate/sql/migrate.sql` | Modify v1.67, add v1.99, v1.100 |
| `api/cmd/services/ichor/build/all/all.go` | Wire actionpermissionsbus, actionservice, actionapi |
| `business/domain/core/tableaccessbus/testutil.go` | Add action_permissions table access |

---

## Verification Plan

### Manual Testing
1. Start local cluster: `make dev-up`
2. Create test user with Warehouse Manager role
3. Grant `allocate_inventory` permission to role
4. Call `POST /v1/workflow/actions/allocate_inventory/execute` with valid config
5. Verify allocation queued and trackable
6. Verify same handler executes as automation
7. Verify `update_field` returns 404/403 when attempted manually

### Automated Testing
1. Run unit tests: `go test ./business/sdk/workflow/...`
2. Run domain tests: `go test ./business/domain/workflow/actionpermissionsbus/...`
3. Run integration tests: `go test ./api/cmd/services/ichor/tests/workflow/actionapi/...`
4. Run full test suite: `make test`

### Integration Verification
1. Create automation rule that triggers allocation on order create
2. Create order → verify automated allocation works (TriggerSource: "automation")
3. Call manual allocation API → verify same result (TriggerSource: "manual")
4. Compare audit trails for both paths in `automation_executions` table
5. Verify `update_field` cannot be executed via manual API

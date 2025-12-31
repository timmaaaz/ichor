# Phase 3: Alert System Enhancement

**Status**: Pending
**Category**: fullstack
**Dependencies**:
- Phase 1 (Form Configuration FK Default Resolution) - Completed
- Phase 2 (Workflow Integration for Status Transitions) - Completed

---

## Overview

Extend the `create_alert` workflow action to support role-based recipients, persistent storage, and user acknowledgment. Currently, the action in `workflowactions/communication/alert.go` is a stub that returns mock data.

**Architectural Approach**:
- **alertstore** - Pure data operations using `sqldb.*` helpers (no validation)
- **CreateAlertHandler.Validate()** - Validates action config (app layer equivalent for workflow actions)
- **alertapi** - Minimal API endpoints with inline validation (simple operations)
- **No alertapp layer** - Validation inline in alertapi handlers

**Key Business Value**: When inventory allocation fails (Phase 2), operations staff are notified via persistent alerts they can view and acknowledge.

---

## Goals

1. Create database tables for persistent alert storage
2. Create alertstore layer (pure CRUD, no validation)
3. Implement alert persistence in CreateAlertHandler using the store
4. Add minimal API endpoints for querying and acknowledging alerts
5. Add comprehensive tests following established patterns

---

## Tasks

### Task 1: Create Alert Database Tables

**Files to Modify:**
- `business/sdk/migrate/sql/migrate.sql`

**SQL Migration:**

```sql
-- Version: 1.76
-- Description: Create workflow alert tables

CREATE TABLE workflow.alerts (
   id UUID NOT NULL,
   alert_type VARCHAR(100) NOT NULL,
   severity VARCHAR(20) NOT NULL,
   title TEXT NOT NULL,
   message TEXT NOT NULL,
   context JSONB DEFAULT '{}',
   source_entity_name VARCHAR(100) NULL,
   source_entity_id UUID NULL,
   source_rule_id UUID NULL,
   status VARCHAR(20) NOT NULL DEFAULT 'active',
   expires_date TIMESTAMP NULL,
   created_date TIMESTAMP NOT NULL,
   updated_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (source_rule_id) REFERENCES workflow.automation_rules(id) ON DELETE SET NULL
);

CREATE INDEX idx_alerts_status ON workflow.alerts(status);
CREATE INDEX idx_alerts_severity ON workflow.alerts(severity);
CREATE INDEX idx_alerts_created_date ON workflow.alerts(created_date DESC);
CREATE INDEX idx_alerts_source_rule ON workflow.alerts(source_rule_id);

-- Version: 1.77
-- Description: Create alert recipients table

CREATE TABLE workflow.alert_recipients (
   id UUID NOT NULL,
   alert_id UUID NOT NULL,
   recipient_type VARCHAR(20) NOT NULL,
   recipient_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (alert_id) REFERENCES workflow.alerts(id) ON DELETE CASCADE,
   UNIQUE (alert_id, recipient_type, recipient_id)
);

CREATE INDEX idx_alert_recipients_alert ON workflow.alert_recipients(alert_id);
CREATE INDEX idx_alert_recipients_recipient ON workflow.alert_recipients(recipient_type, recipient_id);

-- Version: 1.78
-- Description: Create alert acknowledgments table

CREATE TABLE workflow.alert_acknowledgments (
   id UUID NOT NULL,
   alert_id UUID NOT NULL,
   acknowledged_by UUID NOT NULL,
   acknowledged_date TIMESTAMP NOT NULL,
   notes TEXT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (alert_id) REFERENCES workflow.alerts(id) ON DELETE CASCADE,
   FOREIGN KEY (acknowledged_by) REFERENCES core.users(id) ON DELETE CASCADE,
   UNIQUE (alert_id, acknowledged_by)
);

CREATE INDEX idx_alert_ack_alert ON workflow.alert_acknowledgments(alert_id);
CREATE INDEX idx_alert_ack_user ON workflow.alert_acknowledgments(acknowledged_by);
```

---

### Task 2: Create Alert Store Layer

**New Directory:** `business/sdk/workflow/alertstore/`

**New Files:**
- `business/sdk/workflow/alertstore/model.go` - DB models and conversions
- `business/sdk/workflow/alertstore/alertstore.go` - CRUD operations
- `business/sdk/workflow/alertstore/filter.go` - Query filter building
- `business/sdk/workflow/alertstore/order.go` - Order by clause building

**Key Pattern**: Store is pure CRUD with no validation logic. Uses `sqldb.NamedExecContext()`, `sqldb.NamedQuerySlice()`, etc.

**alertstore.go (key methods):**

```go
type Store struct {
    log *logger.Logger
    db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) *Store
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (*Store, error)

// CRUD operations - no validation, just data operations
func (s *Store) Create(ctx context.Context, a Alert) error
func (s *Store) CreateRecipient(ctx context.Context, r AlertRecipient) error
func (s *Store) QueryByID(ctx context.Context, alertID uuid.UUID) (Alert, error)
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error)
func (s *Store) Count(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter) (int, error)
func (s *Store) UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, now time.Time) error
func (s *Store) CreateAcknowledgment(ctx context.Context, ack AlertAcknowledgment) error
```

**model.go:**

```go
// Database models with db tags
type alert struct {
    ID               uuid.UUID       `db:"id"`
    AlertType        string          `db:"alert_type"`
    Severity         string          `db:"severity"`
    Title            string          `db:"title"`
    Message          string          `db:"message"`
    Context          json.RawMessage `db:"context"`
    SourceEntityName sql.NullString  `db:"source_entity_name"`
    SourceEntityID   sql.NullString  `db:"source_entity_id"`
    SourceRuleID     sql.NullString  `db:"source_rule_id"`
    Status           string          `db:"status"`
    ExpiresDate      sql.NullTime    `db:"expires_date"`
    CreatedDate      time.Time       `db:"created_date"`
    UpdatedDate      time.Time       `db:"updated_date"`
}

// Business models (exported)
type Alert struct {
    ID               uuid.UUID
    AlertType        string
    Severity         string
    // ... etc
}

// Conversion functions
func toDBAlert(a Alert) alert { ... }
func toBusAlert(db alert) Alert { ... }
func toBusAlerts(dbs []alert) []Alert { ... }
```

---

### Task 3: Update CreateAlertHandler

**Files to Modify:**
- `business/sdk/workflow/workflowactions/communication/alert.go`

**Key Points:**
- `Validate()` stays as-is (validates config before execution)
- `Execute()` uses alertstore for persistence
- Handler constructs Alert/AlertRecipient structs and passes to store

```go
type CreateAlertHandler struct {
    log   *logger.Logger
    store *alertstore.Store
}

func NewCreateAlertHandler(log *logger.Logger, db *sqlx.DB) *CreateAlertHandler {
    return &CreateAlertHandler{
        log:   log,
        store: alertstore.NewStore(log, db),
    }
}

// Validate - validates action config (app layer equivalent for workflow)
func (h *CreateAlertHandler) Validate(config json.RawMessage) error {
    var cfg AlertConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("invalid configuration format: %w", err)
    }

    if cfg.Message == "" {
        return fmt.Errorf("alert message is required")
    }

    if len(cfg.Recipients.Users) == 0 && len(cfg.Recipients.Roles) == 0 {
        return fmt.Errorf("at least one recipient is required")
    }

    validSeverities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
    if !validSeverities[cfg.Severity] {
        return fmt.Errorf("invalid severity level: %s", cfg.Severity)
    }

    return nil
}

// Execute - creates alert via store
func (h *CreateAlertHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
    var cfg AlertConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    now := time.Now()

    // Build Alert struct
    alert := alertstore.Alert{
        ID:               uuid.New(),
        AlertType:        cfg.AlertType,
        Severity:         cfg.Severity,
        Title:            workflow.ResolveTemplateVars(cfg.Title, execCtx.RawData),
        Message:          workflow.ResolveTemplateVars(cfg.Message, execCtx.RawData),
        Context:          cfg.Context,
        SourceEntityName: execCtx.EntityName,
        SourceEntityID:   execCtx.EntityID,
        SourceRuleID:     execCtx.RuleID,
        Status:           "active",
        CreatedDate:      now,
        UpdatedDate:      now,
    }

    // Create alert via store
    if err := h.store.Create(ctx, alert); err != nil {
        return nil, fmt.Errorf("create alert: %w", err)
    }

    // Create recipients via store
    for _, u := range cfg.Recipients.Users {
        uid, err := uuid.Parse(u)
        if err != nil {
            continue
        }
        r := alertstore.AlertRecipient{
            ID:            uuid.New(),
            AlertID:       alert.ID,
            RecipientType: "user",
            RecipientID:   uid,
            CreatedDate:   now,
        }
        if err := h.store.CreateRecipient(ctx, r); err != nil {
            h.log.Warn(ctx, "failed to create recipient", "error", err)
        }
    }
    // Similar for roles...

    return map[string]interface{}{
        "alert_id": alert.ID.String(),
        "status":   "created",
    }, nil
}
```

---

### Task 4: Create Alert API Layer

**New Files:**
- `api/domain/http/workflow/alertapi/alertapi.go`
- `api/domain/http/workflow/alertapi/route.go`
- `api/domain/http/workflow/alertapi/model.go`
- `api/domain/http/workflow/alertapi/filter.go`

**Key Points:**
- No alertapp layer - validation inline in handlers
- Simple operations (query, acknowledge, dismiss)
- Uses alertstore directly

**route.go:**

```go
const RouteTable = "workflow.alerts"

func Routes(app *web.App, cfg Config) {
    const version = "v1"
    store := alertstore.NewStore(cfg.Log, cfg.DB)
    api := newAPI(cfg.Log, store, cfg.UserRoleBus)
    authen := mid.Authenticate(cfg.AuthClient)

    app.HandlerFunc(http.MethodGet, version, "/workflow/alerts/mine", api.queryMine, authen)
    app.HandlerFunc(http.MethodGet, version, "/workflow/alerts/{id}", api.queryByID, authen)
    app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/{id}/acknowledge", api.acknowledge, authen)
    app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/{id}/dismiss", api.dismiss, authen)
}
```

**alertapi.go (acknowledge with inline validation):**

```go
func (a *api) acknowledge(ctx context.Context, r *http.Request) web.Encoder {
    // Parse path param
    id, err := uuid.Parse(web.Param(r, "id"))
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    // Get current user
    userID, err := mid.GetUserID(ctx)
    if err != nil {
        return errs.New(errs.Unauthenticated, err)
    }

    // Decode and validate request (inline - simple validation)
    var req AcknowledgeRequest
    if err := web.Decode(r, &req); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }
    // Notes field is optional, no validation needed

    // Create acknowledgment
    ack := alertstore.AlertAcknowledgment{
        ID:               uuid.New(),
        AlertID:          id,
        AcknowledgedBy:   userID,
        AcknowledgedDate: time.Now(),
        Notes:            req.Notes,
    }

    if err := a.store.CreateAcknowledgment(ctx, ack); err != nil {
        if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
            return errs.New(errs.AlreadyExists, fmt.Errorf("already acknowledged"))
        }
        return errs.Newf(errs.Internal, "create acknowledgment: %s", err)
    }

    // Update status
    if err := a.store.UpdateStatus(ctx, id, "acknowledged", time.Now()); err != nil {
        return errs.Newf(errs.Internal, "update status: %s", err)
    }

    // Return updated alert
    alert, err := a.store.QueryByID(ctx, id)
    if err != nil {
        return errs.Newf(errs.Internal, "query alert: %s", err)
    }

    return toAppAlert(alert)
}
```

---

### Task 5: Wire Alert API in all.go

**Files to Modify:**
- `api/cmd/services/ichor/build/all/all.go`

```go
import "github.com/timmaaaz/ichor/api/domain/http/workflow/alertapi"

// Route registration
alertapi.Routes(app, alertapi.Config{
    Log:         cfg.Log,
    DB:          cfg.DB,
    AuthClient:  cfg.AuthClient,
    UserRoleBus: userRoleBus,
})
```

---

### Task 6: Add Table Access Permissions

**Files to Modify:**
- `business/domain/core/tableaccessbus/testutil.go`

```go
{RoleID: uuid.Nil, TableName: "workflow.alerts", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
{RoleID: uuid.Nil, TableName: "workflow.alert_recipients", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
{RoleID: uuid.Nil, TableName: "workflow.alert_acknowledgments", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
```

---

### Task 7: Unit Tests for Alert Store

**New File:** `business/sdk/workflow/alertstore/alertstore_test.go`

**Test Pattern (unitest.Table):**

```go
func Test_AlertStore(t *testing.T) {
    t.Parallel()

    db := dbtest.NewDatabase(t, "Test_AlertStore")
    store := NewStore(logger.NewTest(os.Stdout, logger.LevelInfo, "TEST"), db.DB)

    unitest.Run(t, createTests(store), "create")
    unitest.Run(t, queryByIDTests(store), "queryByID")
    unitest.Run(t, updateStatusTests(store), "updateStatus")
    unitest.Run(t, acknowledgmentTests(store), "acknowledgment")
}

func createTests(store *Store) []unitest.Table {
    return []unitest.Table{
        {
            Name: "success",
            ExpResp: nil,
            ExcFunc: func(ctx context.Context) any {
                alert := Alert{
                    ID:          uuid.New(),
                    AlertType:   "test",
                    Severity:    "high",
                    Title:       "Test",
                    Message:     "Test message",
                    Status:      "active",
                    CreatedDate: time.Now(),
                    UpdatedDate: time.Now(),
                }
                return store.Create(ctx, alert)
            },
            CmpFunc: func(got, exp any) string {
                if got != nil {
                    return fmt.Sprintf("expected nil, got: %v", got)
                }
                return ""
            },
        },
    }
}
```

---

### Task 8: Unit Tests for CreateAlertHandler

**New File:** `business/sdk/workflow/workflowactions/communication/alert_test.go`

```go
func Test_CreateAlertHandler(t *testing.T) {
    t.Parallel()

    db := dbtest.NewDatabase(t, "Test_CreateAlertHandler")
    handler := NewCreateAlertHandler(
        logger.NewTest(os.Stdout, logger.LevelInfo, "TEST"),
        db.DB,
    )

    unitest.Run(t, validateTests(handler), "validate")
    unitest.Run(t, executeTests(handler), "execute")
}

func validateTests(handler *CreateAlertHandler) []unitest.Table {
    return []unitest.Table{
        {
            Name:    "valid-config",
            ExpResp: nil,
            ExcFunc: func(ctx context.Context) any {
                config := json.RawMessage(`{
                    "alert_type": "test",
                    "severity": "high",
                    "message": "Test",
                    "recipients": {"users": ["00000000-0000-0000-0000-000000000001"]}
                }`)
                return handler.Validate(config)
            },
            CmpFunc: func(got, exp any) string {
                if got != nil {
                    return fmt.Sprintf("expected nil, got: %v", got)
                }
                return ""
            },
        },
        {
            Name:    "missing-message",
            ExpResp: "alert message is required",
            ExcFunc: func(ctx context.Context) any {
                config := json.RawMessage(`{"severity": "high", "recipients": {"users": ["uuid"]}}`)
                err := handler.Validate(config)
                if err != nil {
                    return err.Error()
                }
                return ""
            },
            CmpFunc: func(got, exp any) string {
                if !strings.Contains(got.(string), exp.(string)) {
                    return fmt.Sprintf("expected %q in %q", exp, got)
                }
                return ""
            },
        },
        {
            Name:    "missing-recipients",
            ExpResp: "at least one recipient",
            ExcFunc: func(ctx context.Context) any {
                config := json.RawMessage(`{"severity": "high", "message": "Test", "recipients": {}}`)
                err := handler.Validate(config)
                if err != nil {
                    return err.Error()
                }
                return ""
            },
            CmpFunc: func(got, exp any) string {
                if !strings.Contains(got.(string), exp.(string)) {
                    return fmt.Sprintf("expected %q in %q", exp, got)
                }
                return ""
            },
        },
        {
            Name:    "invalid-severity",
            ExpResp: "invalid severity",
            ExcFunc: func(ctx context.Context) any {
                config := json.RawMessage(`{"severity": "extreme", "message": "Test", "recipients": {"users": ["uuid"]}}`)
                err := handler.Validate(config)
                if err != nil {
                    return err.Error()
                }
                return ""
            },
            CmpFunc: func(got, exp any) string {
                if !strings.Contains(got.(string), exp.(string)) {
                    return fmt.Sprintf("expected %q in %q", exp, got)
                }
                return ""
            },
        },
    }
}
```

---

### Task 9: Integration Tests for Alert API

**New Files:**
- `api/cmd/services/ichor/tests/workflow/alertapi/alert_test.go`
- `api/cmd/services/ichor/tests/workflow/alertapi/seed_test.go`
- `api/cmd/services/ichor/tests/workflow/alertapi/query_test.go`
- `api/cmd/services/ichor/tests/workflow/alertapi/acknowledge_test.go`

**alert_test.go:**

```go
func Test_Alert(t *testing.T) {
    t.Parallel()

    test := apitest.StartTest(t, "Test_Alert")
    sd, err := insertSeedData(test.DB, test.Auth)
    if err != nil {
        t.Fatalf("Seeding error: %s", err)
    }

    test.Run(t, queryMine200(sd), "queryMine-200")
    test.Run(t, queryMine401(sd), "queryMine-401")
    test.Run(t, queryByID200(sd), "queryByID-200")
    test.Run(t, queryByID404(sd), "queryByID-404")
    test.Run(t, acknowledge200(sd), "acknowledge-200")
    test.Run(t, acknowledge401(sd), "acknowledge-401")
    test.Run(t, dismiss200(sd), "dismiss-200")
}
```

**seed_test.go:**

```go
type SeedData struct {
    Users   []apitest.User
    Admins  []apitest.User
    AlertID uuid.UUID
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (SeedData, error) {
    ctx := context.Background()

    // Create test user and admin (standard pattern)
    usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, db.BusDomain.User)
    // ...

    // Seed alert via store
    store := alertstore.NewStore(db.Log, db.DB)
    alert := alertstore.Alert{
        ID:          uuid.New(),
        AlertType:   "test",
        Severity:    "high",
        Title:       "Test Alert",
        Message:     "Test message",
        Status:      "active",
        CreatedDate: time.Now(),
        UpdatedDate: time.Now(),
    }
    if err := store.Create(ctx, alert); err != nil {
        return SeedData{}, fmt.Errorf("seeding alert: %w", err)
    }

    // Link alert to user
    r := alertstore.AlertRecipient{
        ID:            uuid.New(),
        AlertID:       alert.ID,
        RecipientType: "user",
        RecipientID:   usrs[0].ID,
        CreatedDate:   time.Now(),
    }
    if err := store.CreateRecipient(ctx, r); err != nil {
        return SeedData{}, fmt.Errorf("seeding recipient: %w", err)
    }

    return SeedData{
        Users:   []apitest.User{user},
        Admins:  []apitest.User{admin},
        AlertID: alert.ID,
    }, nil
}
```

---

### Task 10: Update Seed Data for Allocation Failure Rule

**Files to Modify:**
- `business/sdk/dbtest/seedFrontend.go`

Update the "Allocation Failed - Alert Operations" rule with proper recipients config.

---

## Validation Criteria

- [ ] Go compilation passes (`go build ./...`)
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] Migration creates all 3 tables
- [ ] alertstore.Create() uses sqldb.NamedExecContext
- [ ] CreateAlertHandler.Validate() validates config correctly
- [ ] CreateAlertHandler.Execute() persists via store
- [ ] `/v1/workflow/alerts/mine` returns alerts for user
- [ ] `/v1/workflow/alerts/{id}/acknowledge` works
- [ ] Unit tests pass for alertstore
- [ ] Unit tests pass for CreateAlertHandler
- [ ] Integration tests pass for alertapi

---

## Deliverables

- [ ] Database migration (1.76, 1.77, 1.78)
- [ ] alertstore package (pure CRUD)
- [ ] Updated CreateAlertHandler using store
- [ ] alertapi package with inline validation
- [ ] alertapi wired in all.go
- [ ] Table access permissions
- [ ] Unit tests for alertstore
- [ ] Unit tests for CreateAlertHandler
- [ ] Integration tests for alertapi
- [ ] Updated seedFrontend.go

---

## Architecture Summary

```
Workflow Action Flow:
CreateAlertHandler.Validate() → Validates config (app layer equivalent)
CreateAlertHandler.Execute() → Uses alertstore.Create()

API Flow:
alertapi handlers → inline validation → alertstore methods

Store Layer:
alertstore.Create/Query/Update → sqldb.NamedExecContext/NamedQuerySlice
```

**No alertapp layer** - validation is:
1. In CreateAlertHandler.Validate() for workflow actions
2. Inline in alertapi handlers for API requests

---

**Last Updated**: 2025-12-31
**Phase Author**: Claude Code

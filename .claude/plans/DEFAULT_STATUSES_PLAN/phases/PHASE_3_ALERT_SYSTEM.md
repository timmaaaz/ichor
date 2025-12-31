# Phase 3: Alert System Enhancement

**Status**: Pending
**Category**: fullstack
**Dependencies**:
- Phase 1 (Form Configuration FK Default Resolution) - Completed
- Phase 2 (Workflow Integration for Status Transitions) - Completed

---

## Overview

Extend the `create_alert` workflow action to support role-based recipients, persistent storage, and user acknowledgment. Currently, the action in `workflowactions/communication/alert.go` is a stub that returns mock data.

**Key Business Value**: When inventory allocation fails (Phase 2), operations staff are notified via persistent alerts they can view and acknowledge.

---

## Architectural Decision Record

### ADR: Simplified Architecture for Alert System

**Decision**: Use a minimal `alertbus` business layer (thin wrapper) with no `alertapp` application layer. The `alertbus` delegates to `alertdb` store with minimal business logic.

**Context**: The standard Ardan Labs pattern is API → App → Business → Store. However, alerts have a constrained scope:

| Factor | Assessment |
|--------|------------|
| **Creation** | Only via workflow action - no API create endpoint |
| **Business rules** | None beyond basic validation (severity values, required fields) |
| **State transitions** | Simple: active → acknowledged/dismissed |
| **Domain events** | Nothing downstream consumes alert events |
| **Caching needs** | None - alerts are read infrequently |
| **Transaction coordination** | Only alert + recipients (handled in CreateAlertHandler) |

**Consequences**:
- **Positive**: Consistent package naming with other domains (`alertbus/stores/alertdb`)
- **Positive**: Business layer provides clean interface for future expansion
- **Positive**: No alertapp layer - appropriate simplification for read-only API operations
- **Negative**: Thin business layer adds slight indirection
- **Mitigation**: Business layer is minimal - mostly delegates to store

**Architectural Approach**:
- **alertbus** (`business/domain/workflow/alertbus/`) - Minimal business layer with Storer interface
- **alertdb** (`business/domain/workflow/alertbus/stores/alertdb/`) - Pure data operations using `sqldb.*` helpers
- **CreateAlertHandler** - Uses alertbus for persistence
- **alertapi** - Minimal API endpoints with inline validation + authorization, uses alertbus directly
- **No alertapp layer** - Validation inline in alertapi handlers (simple read/update operations)

**Review Date**: Revisit if alerts require complex business logic, multiple creation entry points, or domain events.

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
   severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
   title TEXT NOT NULL,
   message TEXT NOT NULL,
   context JSONB DEFAULT '{}',
   source_entity_name VARCHAR(100) NULL,
   source_entity_id UUID NULL,
   source_rule_id UUID NULL,
   status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'dismissed')),
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
CREATE INDEX idx_alerts_expires_date ON workflow.alerts(expires_date) WHERE expires_date IS NOT NULL;

-- Version: 1.77
-- Description: Create alert recipients table

CREATE TABLE workflow.alert_recipients (
   id UUID NOT NULL,
   alert_id UUID NOT NULL,
   recipient_type VARCHAR(20) NOT NULL CHECK (recipient_type IN ('user', 'role')),
   recipient_id UUID NOT NULL,
   created_date TIMESTAMP NOT NULL,
   PRIMARY KEY (id),
   FOREIGN KEY (alert_id) REFERENCES workflow.alerts(id) ON DELETE CASCADE,
   UNIQUE (alert_id, recipient_type, recipient_id)
);

CREATE INDEX idx_alert_recipients_alert ON workflow.alert_recipients(alert_id);
CREATE INDEX idx_alert_recipients_recipient ON workflow.alert_recipients(recipient_type, recipient_id);
CREATE INDEX idx_alert_recipients_lookup ON workflow.alert_recipients(recipient_id, recipient_type, alert_id);

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

### Task 2: Create Alert Business and Store Layers

**New Directory Structure:**
```
business/domain/workflow/alertbus/
├── alertbus.go           # Minimal business layer
├── model.go              # Alert, AlertRecipient, AlertAcknowledgment models
├── filter.go             # QueryFilter
├── order.go              # Order by constants
└── stores/alertdb/
    ├── alertdb.go        # CRUD operations
    ├── model.go          # DB models and conversions
    ├── filter.go         # SQL filter building
    └── order.go          # SQL order building
```

#### alertbus Package (Minimal Business Layer)

**alertbus.go:**

```go
package alertbus

import (
    "context"
    "errors"
    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/sdk/order"
    "github.com/timmaaaz/ichor/business/sdk/page"
    "github.com/timmaaaz/ichor/foundation/logger"
)

var (
    ErrNotFound      = errors.New("alert not found")
    ErrNotRecipient  = errors.New("user is not a recipient of this alert")
    ErrAlreadyAcked  = errors.New("alert already acknowledged by this user")
)

// Storer interface defines data operations
type Storer interface {
    Create(ctx context.Context, alert Alert) error
    CreateRecipients(ctx context.Context, recipients []AlertRecipient) error  // Batch insert
    CreateAcknowledgment(ctx context.Context, ack AlertAcknowledgment) error
    QueryByID(ctx context.Context, alertID uuid.UUID) (Alert, error)
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error)  // Admin query
    QueryByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)  // Admin count
    CountByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter) (int, error)
    UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, now time.Time) error
    IsRecipient(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error)
}

// Business manages alert operations
type Business struct {
    log    *logger.Logger
    storer Storer
}

func NewBusiness(log *logger.Logger, storer Storer) *Business {
    return &Business{log: log, storer: storer}
}

// Create creates a new alert (called by CreateAlertHandler)
func (b *Business) Create(ctx context.Context, alert Alert) error {
    return b.storer.Create(ctx, alert)
}

// CreateRecipients adds multiple recipients to an alert (batch insert)
func (b *Business) CreateRecipients(ctx context.Context, recipients []AlertRecipient) error {
    if len(recipients) == 0 {
        return nil
    }
    return b.storer.CreateRecipients(ctx, recipients)
}

// QueryByID returns a single alert
func (b *Business) QueryByID(ctx context.Context, alertID uuid.UUID) (Alert, error) {
    return b.storer.QueryByID(ctx, alertID)
}

// Query returns all alerts (admin only - no recipient filtering)
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error) {
    return b.storer.Query(ctx, filter, orderBy, pg)
}

// Count returns count of all alerts (admin only)
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
    return b.storer.Count(ctx, filter)
}

// QueryMine returns alerts for a user (directly or via roles)
func (b *Business) QueryMine(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter, orderBy order.By, pg page.Page) ([]Alert, error) {
    return b.storer.QueryByUserID(ctx, userID, roleIDs, filter, orderBy, pg)
}

// CountMine returns count of alerts for a user
func (b *Business) CountMine(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter QueryFilter) (int, error) {
    return b.storer.CountByUserID(ctx, userID, roleIDs, filter)
}

// Acknowledge marks an alert as acknowledged by a user
// Validates that user is a recipient before allowing acknowledgment
func (b *Business) Acknowledge(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID, notes string, now time.Time) (Alert, error) {
    // Security check: verify user is a recipient
    isRecipient, err := b.storer.IsRecipient(ctx, alertID, userID, roleIDs)
    if err != nil {
        return Alert{}, err
    }
    if !isRecipient {
        return Alert{}, ErrNotRecipient
    }

    ack := AlertAcknowledgment{
        ID:               uuid.New(),
        AlertID:          alertID,
        AcknowledgedBy:   userID,
        AcknowledgedDate: now,
        Notes:            notes,
    }

    if err := b.storer.CreateAcknowledgment(ctx, ack); err != nil {
        return Alert{}, err
    }

    if err := b.storer.UpdateStatus(ctx, alertID, StatusAcknowledged, now); err != nil {
        return Alert{}, err
    }

    return b.storer.QueryByID(ctx, alertID)
}

// Dismiss marks an alert as dismissed by a user
// Validates that user is a recipient before allowing dismissal
func (b *Business) Dismiss(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID, now time.Time) (Alert, error) {
    // Security check: verify user is a recipient
    isRecipient, err := b.storer.IsRecipient(ctx, alertID, userID, roleIDs)
    if err != nil {
        return Alert{}, err
    }
    if !isRecipient {
        return Alert{}, ErrNotRecipient
    }

    if err := b.storer.UpdateStatus(ctx, alertID, StatusDismissed, now); err != nil {
        return Alert{}, err
    }

    return b.storer.QueryByID(ctx, alertID)
}
```

**model.go:**

```go
package alertbus

import (
    "encoding/json"
    "github.com/google/uuid"
    "time"
)

// Status constants
const (
    StatusActive       = "active"
    StatusAcknowledged = "acknowledged"
    StatusDismissed    = "dismissed"
)

// Severity constants
const (
    SeverityLow      = "low"
    SeverityMedium   = "medium"
    SeverityHigh     = "high"
    SeverityCritical = "critical"
)

type Alert struct {
    ID               uuid.UUID
    AlertType        string
    Severity         string
    Title            string
    Message          string
    Context          json.RawMessage
    SourceEntityName string
    SourceEntityID   uuid.UUID
    SourceRuleID     uuid.UUID
    Status           string
    ExpiresDate      *time.Time  // Pointer to represent NULL
    CreatedDate      time.Time
    UpdatedDate      time.Time
}

type AlertRecipient struct {
    ID            uuid.UUID
    AlertID       uuid.UUID
    RecipientType string // "user" or "role"
    RecipientID   uuid.UUID
    CreatedDate   time.Time
}

type AlertAcknowledgment struct {
    ID               uuid.UUID
    AlertID          uuid.UUID
    AcknowledgedBy   uuid.UUID
    AcknowledgedDate time.Time
    Notes            string
}
```

#### alertdb Package (Store Layer)

**stores/alertdb/alertdb.go:**

```go
package alertdb

type Store struct {
    log *logger.Logger
    db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) *Store
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (*Store, error)

// Implements alertbus.Storer interface
func (s *Store) Create(ctx context.Context, a alertbus.Alert) error
func (s *Store) CreateRecipients(ctx context.Context, recipients []alertbus.AlertRecipient) error  // Batch insert
func (s *Store) CreateAcknowledgment(ctx context.Context, ack alertbus.AlertAcknowledgment) error
func (s *Store) QueryByID(ctx context.Context, alertID uuid.UUID) (alertbus.Alert, error)
func (s *Store) Query(ctx context.Context, filter alertbus.QueryFilter, orderBy order.By, pg page.Page) ([]alertbus.Alert, error)  // Admin
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter alertbus.QueryFilter, orderBy order.By, pg page.Page) ([]alertbus.Alert, error)
func (s *Store) Count(ctx context.Context, filter alertbus.QueryFilter) (int, error)  // Admin
func (s *Store) CountByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter alertbus.QueryFilter) (int, error)
func (s *Store) UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, now time.Time) error
func (s *Store) IsRecipient(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error)
```

**stores/alertdb/model.go:**

```go
package alertdb

// Database models with db tags (unexported)
type dbAlert struct {
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

// Conversion functions
func toDBAlert(a alertbus.Alert) dbAlert { ... }
func toBusAlert(db dbAlert) alertbus.Alert { ... }
func toBusAlerts(dbs []dbAlert) []alertbus.Alert { ... }
```

---

### Task 3: Update CreateAlertHandler

**Files to Modify:**
- `business/sdk/workflow/workflowactions/communication/alert.go`

**Key Points:**
- `Validate()` stays as-is (validates config before execution)
- `Execute()` uses alertbus for persistence
- Handler constructs Alert/AlertRecipient structs and passes to business layer

```go
type CreateAlertHandler struct {
    log      *logger.Logger
    alertBus *alertbus.Business
}

func NewCreateAlertHandler(log *logger.Logger, alertBus *alertbus.Business) *CreateAlertHandler {
    return &CreateAlertHandler{
        log:      log,
        alertBus: alertBus,
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

// Execute - creates alert via business layer
func (h *CreateAlertHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
    var cfg AlertConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    now := time.Now()

    // Build Alert struct
    alert := alertbus.Alert{
        ID:               uuid.New(),
        AlertType:        cfg.AlertType,
        Severity:         cfg.Severity,
        Title:            workflow.ResolveTemplateVars(cfg.Title, execCtx.RawData),
        Message:          workflow.ResolveTemplateVars(cfg.Message, execCtx.RawData),
        Context:          cfg.Context,
        SourceEntityName: execCtx.EntityName,
        SourceEntityID:   execCtx.EntityID,
        SourceRuleID:     execCtx.RuleID,
        Status:           alertbus.StatusActive,
        CreatedDate:      now,
        UpdatedDate:      now,
    }

    // Build recipients slice - validate all UUIDs first (fail fast on invalid config)
    var recipients []alertbus.AlertRecipient

    for _, u := range cfg.Recipients.Users {
        uid, err := uuid.Parse(u)
        if err != nil {
            return nil, fmt.Errorf("invalid user UUID %q: %w", u, err)
        }
        recipients = append(recipients, alertbus.AlertRecipient{
            ID:            uuid.New(),
            AlertID:       alert.ID,
            RecipientType: "user",
            RecipientID:   uid,
            CreatedDate:   now,
        })
    }

    for _, r := range cfg.Recipients.Roles {
        rid, err := uuid.Parse(r)
        if err != nil {
            return nil, fmt.Errorf("invalid role UUID %q: %w", r, err)
        }
        recipients = append(recipients, alertbus.AlertRecipient{
            ID:            uuid.New(),
            AlertID:       alert.ID,
            RecipientType: "role",
            RecipientID:   rid,
            CreatedDate:   now,
        })
    }

    // Create alert via business layer
    if err := h.alertBus.Create(ctx, alert); err != nil {
        return nil, fmt.Errorf("create alert: %w", err)
    }

    // Create all recipients via batch insert
    if err := h.alertBus.CreateRecipients(ctx, recipients); err != nil {
        return nil, fmt.Errorf("create recipients: %w", err)
    }

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
- Uses alertbus directly (business layer handles recipient authorization)
- All routes require authentication + authorization middleware

**route.go:**

```go
type Config struct {
    Log         *logger.Logger
    AlertBus    *alertbus.Business
    UserRoleBus *userrolebus.Business
    AuthClient  *authclient.Client
}

func Routes(app *web.App, cfg Config) {
    const version = "v1"
    api := newAPI(cfg.Log, cfg.AlertBus, cfg.UserRoleBus)
    authen := mid.Authenticate(cfg.AuthClient)

    // User endpoints - authentication only, business layer handles recipient filtering
    // Everyone can access alerts table, but they only see alerts they're recipients of
    app.HandlerFunc(http.MethodGet, version, "/workflow/alerts/mine", api.queryMine, authen)
    app.HandlerFunc(http.MethodGet, version, "/workflow/alerts/{id}", api.queryByID, authen)
    app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/{id}/acknowledge", api.acknowledge, authen)
    app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/{id}/dismiss", api.dismiss, authen)

    // Admin endpoint - requires admin role, sees all alerts
    app.HandlerFunc(http.MethodGet, version, "/workflow/alerts", api.query, authen,
        mid.Authorize(cfg.AuthClient, nil, "", "", auth.RuleAdminOnly))
}
```

**alertapi.go:**

```go
type api struct {
    log         *logger.Logger
    alertBus    *alertbus.Business
    userRoleBus *userrolebus.Business
}

func newAPI(log *logger.Logger, alertBus *alertbus.Business, userRoleBus *userrolebus.Business) *api {
    return &api{log: log, alertBus: alertBus, userRoleBus: userRoleBus}
}

// query returns all alerts (admin only)
func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
    qp := parseQueryParams(r)

    filter, err := parseFilter(qp)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    orderBy, err := parseOrder(qp)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    pg, err := page.Parse(qp.Page, qp.Rows)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    alerts, err := a.alertBus.Query(ctx, filter, orderBy, pg)
    if err != nil {
        return errs.Newf(errs.Internal, "query: %s", err)
    }

    total, err := a.alertBus.Count(ctx, filter)
    if err != nil {
        return errs.Newf(errs.Internal, "count: %s", err)
    }

    return query.NewResult(toAppAlerts(alerts), total, pg)
}

// acknowledge delegates to business layer which handles recipient authorization
func (a *api) acknowledge(ctx context.Context, r *http.Request) web.Encoder {
    // Parse path param
    id, err := uuid.Parse(web.Param(r, "id"))
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    // Get current user from context
    userID, err := mid.GetUserID(ctx)
    if err != nil {
        return errs.New(errs.Unauthenticated, err)
    }

    // Get user's roles for recipient check
    roleIDs, err := a.getUserRoleIDs(ctx, userID)
    if err != nil {
        return errs.Newf(errs.Internal, "get user roles: %s", err)
    }

    // Decode request (inline - simple validation)
    var req AcknowledgeRequest
    if err := web.Decode(r, &req); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    // Business layer handles:
    // 1. Recipient authorization check
    // 2. Creating acknowledgment record
    // 3. Updating alert status
    alert, err := a.alertBus.Acknowledge(ctx, id, userID, roleIDs, req.Notes, time.Now())
    if err != nil {
        if errors.Is(err, alertbus.ErrNotRecipient) {
            return errs.New(errs.PermissionDenied, err)
        }
        if errors.Is(err, alertbus.ErrNotFound) {
            return errs.New(errs.NotFound, err)
        }
        if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
            return errs.New(errs.AlreadyExists, fmt.Errorf("already acknowledged"))
        }
        return errs.Newf(errs.Internal, "acknowledge: %s", err)
    }

    return toAppAlert(alert)
}

// dismiss delegates to business layer which handles recipient authorization
func (a *api) dismiss(ctx context.Context, r *http.Request) web.Encoder {
    id, err := uuid.Parse(web.Param(r, "id"))
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    userID, err := mid.GetUserID(ctx)
    if err != nil {
        return errs.New(errs.Unauthenticated, err)
    }

    roleIDs, err := a.getUserRoleIDs(ctx, userID)
    if err != nil {
        return errs.Newf(errs.Internal, "get user roles: %s", err)
    }

    alert, err := a.alertBus.Dismiss(ctx, id, userID, roleIDs, time.Now())
    if err != nil {
        if errors.Is(err, alertbus.ErrNotRecipient) {
            return errs.New(errs.PermissionDenied, err)
        }
        if errors.Is(err, alertbus.ErrNotFound) {
            return errs.New(errs.NotFound, err)
        }
        return errs.Newf(errs.Internal, "dismiss: %s", err)
    }

    return toAppAlert(alert)
}

// getUserRoleIDs fetches role IDs for the current user
func (a *api) getUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
    userRoles, err := a.userRoleBus.QueryByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }
    roleIDs := make([]uuid.UUID, len(userRoles))
    for i, ur := range userRoles {
        roleIDs[i] = ur.RoleID
    }
    return roleIDs, nil
}
```

---

### Task 5: Wire Alert API in all.go

**Files to Modify:**
- `api/cmd/services/ichor/build/all/all.go`

```go
import (
    "github.com/timmaaaz/ichor/api/domain/http/workflow/alertapi"
    "github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
    "github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
)

// Instantiate alertbus (around line 320 with other business layer instantiations)
alertBus := alertbus.NewBusiness(cfg.Log, alertdb.NewStore(cfg.Log, cfg.DB))

// Route registration (around line 520 with other route registrations)
alertapi.Routes(app, alertapi.Config{
    Log:         cfg.Log,
    AlertBus:    alertBus,
    UserRoleBus: userRoleBus,
    AuthClient:  cfg.AuthClient,
})
```

**Also update workflow action handler registration** in `business/sdk/workflow/workflowactions/register.go`:

```go
// Update BusDependencies struct to include AlertBus
type BusDependencies struct {
    // ... existing fields
    Alert *alertbus.Business
}

// Update handler registration
registry.Register(communication.NewCreateAlertHandler(config.Log, config.Buses.Alert))
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

**New File:** `business/domain/workflow/alertbus/stores/alertdb/alertdb_test.go`

**Test Pattern (unitest.Table):**

```go
func Test_AlertDB(t *testing.T) {
    t.Parallel()

    db := dbtest.NewDatabase(t, "Test_AlertDB")
    store := NewStore(logger.NewTest(os.Stdout, logger.LevelInfo, "TEST"), db.DB)

    unitest.Run(t, createTests(store), "create")
    unitest.Run(t, queryByIDTests(store), "queryByID")
    unitest.Run(t, updateStatusTests(store), "updateStatus")
    unitest.Run(t, acknowledgmentTests(store), "acknowledgment")
    unitest.Run(t, isRecipientTests(store), "isRecipient")
}

func createTests(store *Store) []unitest.Table {
    return []unitest.Table{
        {
            Name: "success",
            ExpResp: nil,
            ExcFunc: func(ctx context.Context) any {
                alert := alertbus.Alert{
                    ID:          uuid.New(),
                    AlertType:   "test",
                    Severity:    alertbus.SeverityHigh,
                    Title:       "Test",
                    Message:     "Test message",
                    Status:      alertbus.StatusActive,
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
    alertBus := alertbus.NewBusiness(
        logger.NewTest(os.Stdout, logger.LevelInfo, "TEST"),
        alertdb.NewStore(logger.NewTest(os.Stdout, logger.LevelInfo, "TEST"), db.DB),
    )
    handler := NewCreateAlertHandler(
        logger.NewTest(os.Stdout, logger.LevelInfo, "TEST"),
        alertBus,
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

    // Seed alert via business layer
    alertBus := alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))
    alert := alertbus.Alert{
        ID:          uuid.New(),
        AlertType:   "test",
        Severity:    alertbus.SeverityHigh,
        Title:       "Test Alert",
        Message:     "Test message",
        Status:      alertbus.StatusActive,
        CreatedDate: time.Now(),
        UpdatedDate: time.Now(),
    }
    if err := alertBus.Create(ctx, alert); err != nil {
        return SeedData{}, fmt.Errorf("seeding alert: %w", err)
    }

    // Link alert to user
    r := alertbus.AlertRecipient{
        ID:            uuid.New(),
        AlertID:       alert.ID,
        RecipientType: "user",
        RecipientID:   usrs[0].ID,
        CreatedDate:   time.Now(),
    }
    if err := alertBus.CreateRecipient(ctx, r); err != nil {
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
- [ ] Migration creates all 3 tables with CHECK constraints
- [ ] alertdb.Create() uses sqldb.NamedExecContext
- [ ] alertbus.Acknowledge() validates recipient authorization
- [ ] alertbus.Dismiss() validates recipient authorization
- [ ] CreateAlertHandler.Validate() validates config correctly
- [ ] CreateAlertHandler.Execute() persists via alertbus
- [ ] `/v1/workflow/alerts/mine` returns alerts for user with authorization
- [ ] `/v1/workflow/alerts/{id}/acknowledge` checks recipient authorization
- [ ] Unit tests pass for alertdb
- [ ] Unit tests pass for CreateAlertHandler
- [ ] Integration tests pass for alertapi (including 403 tests for non-recipients)

---

## Deliverables

- [ ] Database migration (1.76, 1.77, 1.78) with CHECK constraints
- [ ] alertbus package (minimal business layer with recipient authorization)
- [ ] alertdb package (pure CRUD store)
- [ ] Updated CreateAlertHandler using alertbus
- [ ] alertapi package with inline validation + authorization middleware
- [ ] alertapi wired in all.go with PermissionsBus
- [ ] Workflow action handler registration updated with alertbus dependency
- [ ] Table access permissions
- [ ] Unit tests for alertdb
- [ ] Unit tests for CreateAlertHandler
- [ ] Integration tests for alertapi (including authorization tests)
- [ ] Updated seedFrontend.go

---

## Architecture Summary

```
Package Structure:
business/domain/workflow/alertbus/           ← Minimal business layer
├── alertbus.go                              ← Storer interface + recipient auth
├── model.go                                 ← Alert, AlertRecipient, AlertAcknowledgment
├── filter.go, order.go
└── stores/alertdb/                          ← Pure CRUD store
    ├── alertdb.go
    ├── model.go
    └── filter.go, order.go

Workflow Action Flow:
CreateAlertHandler.Validate() → Validates config (fail fast on invalid UUIDs)
CreateAlertHandler.Execute() → alertbus.Create() + alertbus.CreateRecipients() (batch)

API Flow:
alertapi handlers → mid.Authenticate() → alertbus methods (with recipient check)

Authorization Model:
- User endpoints: Authentication only (no table-level permissions)
- Admin endpoint (/workflow/alerts): Requires auth.RuleAdminOnly
- Business layer filters alerts by recipient (user can only see their alerts)
- Acknowledge/Dismiss verify IsRecipient() before allowing operation

Security:
- Invalid recipient UUIDs cause action failure (not silently skipped)
- QueryMine returns only alerts where user is recipient (directly or via role)
- queryByID also validates recipient access in business layer
```

**No alertapp layer** - validation is:
1. In CreateAlertHandler.Validate() for workflow actions
2. Inline in alertapi handlers for API requests
3. Recipient authorization in alertbus.Acknowledge() and alertbus.Dismiss()

---

**Last Updated**: 2025-12-31
**Phase Author**: Claude Code

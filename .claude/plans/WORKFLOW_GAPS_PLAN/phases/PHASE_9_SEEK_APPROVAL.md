# Phase 9: Implement seek_approval

**Category**: Backend + Database
**Status**: Pending
**Dependencies**: None (Phase 1 whitelist already done)
**Effort**: High (6 tasks, ~15 files created/modified)

---

## Overview

`seek_approval` is the most complex remaining stub. Currently:
- `Execute()` returns a fake `approval_id` with `"output": "approved"` hardcoded — the workflow always takes the approved branch
- No approval record is persisted
- No approvers are notified
- No mechanism exists for approvers to submit decisions
- `IsAsync()` returns `false` — not wired into async completion
- `selectActivityFunc()` hardcodes `"ExecuteActionActivity"` for all actions — the `ExecuteAsyncActionActivity` path is dead code in production

The full implementation requires:
1. Enable the async activity routing path (`selectActivityFunc`)
2. A new `workflow.approval_requests` database table
3. An `approvalrequestbus` business package (following Ardan Labs patterns)
4. `SeekApprovalHandler` implements `AsyncActivityHandler.StartAsync()` (pauses Temporal workflow)
5. An API layer (`approvalapi`) for approvers to query pending requests and submit decisions
6. Wiring in `register.go`, `all.go`, and the worker binary

### Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Activity routing | Enable async via `selectActivityFunc` | No wasted worker goroutines; clean Temporal pattern for human-in-the-loop |
| Approval type | `"any"` only (first approver wins) | `"all"` and `"majority"` accepted by schema but treated as `"any"` with log warning. Deferred to future phase. |
| Authorization | Approver + admin + role | Verify caller is in approvers array OR has admin role. Follows `alertbus.IsRecipient` pattern. |
| Alert notification | Single alert + per-approver recipients | Follows `CreateAlertHandler` pattern: one `alertbus.Alert`, multiple `alertbus.AlertRecipient` rows |
| Query endpoints | `GET /mine` + `GET /` (admin) | Matches `alertapi` pattern |
| Timeout | Schema stores `timeout_hours`, no auto-expiry | Auto-expiry via Temporal timer deferred to future phase |

---

## Codebase Facts (Verified by Discovery)

These facts were verified by reading the actual source code:

1. **`selectActivityFunc`** (workflow.go:614-621) hardcodes `"ExecuteActionActivity"` for ALL action types. The `_ = actionType` line explicitly discards the argument. `ExecuteAsyncActionActivity` exists (activities.go:108-154) but is never called.

2. **`humanActionTypes`** (workflow.go:35-41) contains: `manager_approval`, `manual_review`, `human_verification`, `approval_request`. Does NOT contain `seek_approval`. These maps only affect timeout config in `activityOptions()`, not routing.

3. **`AsyncRegistry`** (activities_async.go:41-61) exists and is functional but always created empty. Worker binary (main.go:216-218) does not set `AsyncRegistry` on `Activities` struct.

4. **`ActionExecutionContext`** (models.go:57-69) has NO `ActionName` field. Action name is only on `ActionActivityInput` (temporal models.go:336-345). `buildExecContext()` (activities.go:167-216) does not include `ActionName`.

5. **Alert creation pattern** (alert.go:126-178): One `alertbus.Alert` created, then `alertbus.CreateRecipients()` called with `[]AlertRecipient` (one per user/role recipient). Two separate bus calls required.

6. **`AsyncCompleter.Complete`** (async_completer.go:34) takes `ActionActivityOutput` — NOT a raw map. The `Result` field's `"output"` key drives edge resolution in `GetNextActions()` (graph_executor.go:104-107).

7. **Current migration version**: `1.996` → next is `1.997`.

8. **`RegisterCoreActions`** (register.go:146-167) currently calls `approval.NewSeekApprovalHandler(log, db)` with 2 args. Must be updated when constructor signature changes.

9. **`BusDependencies`** (register.go:44-59) has no `ApprovalRequest` field. Must be added.

10. **`alertbus.Business`** has no delegate (alertbus.go:55-60) — constructed as `NewBusiness(log, storer)`. Approval requests similarly should NOT trigger workflow events (avoids circular trigger loops).

---

## Task Breakdown

### Task 1: Enable Async Activity Routing

**Goal**: Make `selectActivityFunc` route `seek_approval` to `ExecuteAsyncActionActivity`.

**Files to modify:**

| File | Change |
|------|--------|
| `business/sdk/workflow/temporal/workflow.go` | Modify `selectActivityFunc` to check `humanActionTypes`; add `"seek_approval"` to `humanActionTypes` |
| `business/sdk/workflow/temporal/allocate_inventory_async_test.go` | Update `TestAllActionsRouteThroughSyncActivity` — `seek_approval` now routes to async |

**Implementation details:**

`selectActivityFunc` currently:
```go
func selectActivityFunc(actionType string) string {
    _ = actionType
    return "ExecuteActionActivity"
}
```

Change to:
```go
func selectActivityFunc(actionType string) string {
    if humanActionTypes[actionType] {
        return "ExecuteAsyncActionActivity"
    }
    return "ExecuteActionActivity"
}
```

Add `"seek_approval": true` to `humanActionTypes` map (line 35-41):
```go
var humanActionTypes = map[string]bool{
    "manager_approval":   true,
    "manual_review":      true,
    "human_verification": true,
    "approval_request":   true,
    "seek_approval":      true,
}
```

**Exact test changes in `allocate_inventory_async_test.go`:**

1. **`TestAllActionsRouteThroughSyncActivity`**: Remove `"seek_approval"` from this test's action type list (it no longer routes to sync).

2. **`TestHumanActionsGetMultiDayTimeouts`**: Add `"seek_approval"` to the `humanTypes` slice (since it's now in `humanActionTypes` map).

3. **Add new test `TestSeekApprovalRoutesToAsyncActivity`**:
```go
func TestSeekApprovalRoutesToAsyncActivity(t *testing.T) {
    result := selectActivityFunc("seek_approval")
    require.Equal(t, "ExecuteAsyncActionActivity", result)
}
```

**Validation:**
- `go build ./business/sdk/workflow/temporal/...`
- `go test ./business/sdk/workflow/temporal/... -run TestAllActions`
- `go test ./business/sdk/workflow/temporal/... -run TestSeekApproval`
- `go test ./business/sdk/workflow/temporal/... -run TestHumanActions`

---

### Task 2: Create `workflow.approval_requests` Migration

**File**: `business/sdk/migrate/sql/migrate.sql`

Append version `1.997`:

```sql
-- Version: 1.997
-- Description: Add workflow approval requests table for seek_approval async activity

CREATE TABLE workflow.approval_requests (
    approval_request_id UUID        NOT NULL,
    execution_id        UUID        NOT NULL REFERENCES workflow.automation_executions(automation_execution_id),
    rule_id             UUID        NOT NULL REFERENCES workflow.automation_rules(automation_rule_id),
    action_name         VARCHAR(100) NOT NULL,
    approvers           UUID[]      NOT NULL,
    approval_type       VARCHAR(20) NOT NULL DEFAULT 'any' CHECK (approval_type IN ('any', 'all', 'majority')),
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'timed_out', 'expired')),
    timeout_hours       INT         NOT NULL DEFAULT 72,
    task_token          TEXT        NOT NULL,
    approval_message    TEXT,
    resolved_by         UUID        REFERENCES core.users(user_id),
    resolution_reason   TEXT,
    created_date        TIMESTAMP   NOT NULL DEFAULT NOW(),
    resolved_date       TIMESTAMP,

    PRIMARY KEY (approval_request_id)
);

CREATE INDEX idx_approval_requests_execution ON workflow.approval_requests(execution_id);
CREATE INDEX idx_approval_requests_status ON workflow.approval_requests(status) WHERE status = 'pending';
CREATE INDEX idx_approval_requests_approvers ON workflow.approval_requests USING GIN (approvers);
```

**Schema notes:**
- `task_token TEXT NOT NULL` — base64-encoded Temporal task token. NOT NULL because every request must have one (it's the correlation ID for async completion).
- `resolved_by UUID` — nullable (NULL when pending). References `core.users(user_id)`.
- `approvers UUID[]` — PostgreSQL array of user UUIDs. GIN index enables `@>` (contains) queries for "find approvals where I'm an approver".
- `approval_type` defaults to `'any'` — all types accepted by schema, but only `'any'` implemented initially.

**Validation:**
- `make migrate` succeeds
- `\d workflow.approval_requests` shows correct schema

---

### Task 3: Create `approvalrequestbus` Business Package

**New directory**: `business/domain/workflow/approvalrequestbus/`

Follow the Ardan Labs bus package pattern (reference: `inventorytransactionbus`, `alertbus`).

**File structure:**
```
business/domain/workflow/approvalrequestbus/
├── model.go
├── approvalrequestbus.go
├── filter.go
├── order.go
└── stores/approvalrequestdb/
    ├── approvalrequestdb.go
    ├── model.go
    ├── filter.go
    └── order.go
```

#### model.go

```go
package approvalrequestbus

type ApprovalRequest struct {
    ID               uuid.UUID
    ExecutionID      uuid.UUID
    RuleID           uuid.UUID
    ActionName       string
    Approvers        []uuid.UUID
    ApprovalType     string     // "any" | "all" | "majority"
    Status           string     // "pending" | "approved" | "rejected" | "timed_out" | "expired"
    TimeoutHours     int
    TaskToken        string     // base64-encoded Temporal task token
    ApprovalMessage  string
    ResolvedBy       *uuid.UUID // nil when pending
    ResolutionReason string
    CreatedDate      time.Time
    ResolvedDate     *time.Time // nil when pending
}

type NewApprovalRequest struct {
    ExecutionID     uuid.UUID
    RuleID          uuid.UUID
    ActionName      string
    Approvers       []uuid.UUID
    ApprovalType    string
    TimeoutHours    int
    TaskToken       string
    ApprovalMessage string
}

// No UpdateApprovalRequest needed — Resolve is an atomic store-level operation
// (see Task 3 Storer.Resolve method)
```

**Key differences from old plan:**
- `ResolvedBy` is `*uuid.UUID` (nullable — nil when unresolved)
- `ResolvedDate` is `*time.Time` (nullable — nil when unresolved)
- `ID` generated in `Create()` (not caller-provided) — matches bus convention
- No `event.go` — approval requests should NOT trigger workflow events (avoids circular triggers, same as `alertbus`)

#### approvalrequestbus.go

```go
type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Create(ctx context.Context, req ApprovalRequest) error
    QueryByID(ctx context.Context, id uuid.UUID) (ApprovalRequest, error)
    Resolve(ctx context.Context, id, resolvedBy uuid.UUID, status, reason string) (ApprovalRequest, error)
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]ApprovalRequest, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
    IsApprover(ctx context.Context, approvalID, userID uuid.UUID) (bool, error)
}

type Business struct {
    log    *logger.Logger
    storer Storer
}

func NewBusiness(log *logger.Logger, storer Storer) *Business
```

Methods:
- `Create(ctx, NewApprovalRequest) (ApprovalRequest, error)` — generates UUID + timestamps, calls `storer.Create`
- `QueryByID(ctx, uuid.UUID) (ApprovalRequest, error)` — thin wrapper with otel span
- `Resolve(ctx, id uuid.UUID, resolvedBy uuid.UUID, status, reason string) (ApprovalRequest, error)` — validates status is `"pending"`, does conditional update, returns updated request
- `Query(ctx, QueryFilter, order.By, page.Page) ([]ApprovalRequest, error)` — standard list
- `Count(ctx, QueryFilter) (int, error)` — standard count
- `IsApprover(ctx, approvalID, userID uuid.UUID) (bool, error)` — checks `approvers @> ARRAY[userID]`

The `Resolve` method is critical — it must be the single point for status transitions. **It must be atomic to prevent TOCTOU races** (two approvers both reading status="pending" then both succeeding). Use a conditional UPDATE that returns zero rows if already resolved:

**Storer interface** — add a dedicated `Resolve` method (not generic `Update`):
```go
type Storer interface {
    NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
    Create(ctx context.Context, req ApprovalRequest) error
    QueryByID(ctx context.Context, id uuid.UUID) (ApprovalRequest, error)
    Resolve(ctx context.Context, id, resolvedBy uuid.UUID, status, reason string) (ApprovalRequest, error)
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]ApprovalRequest, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
    IsApprover(ctx context.Context, approvalID, userID uuid.UUID) (bool, error)
}
```

**Store-level `Resolve`** — single atomic SQL statement:
```go
const resolveQuery = `
    UPDATE workflow.approval_requests
    SET status = :status, resolved_by = :resolved_by, resolution_reason = :resolution_reason, resolved_date = NOW()
    WHERE approval_request_id = :id AND status = 'pending'
    RETURNING *`

func (s *Store) Resolve(ctx context.Context, id, resolvedBy uuid.UUID, status, reason string) (approvalrequestbus.ApprovalRequest, error) {
    data := struct {
        ID               uuid.UUID `db:"id"`
        Status           string    `db:"status"`
        ResolvedBy       uuid.UUID `db:"resolved_by"`
        ResolutionReason string    `db:"resolution_reason"`
    }{
        ID: id, Status: status, ResolvedBy: resolvedBy, ResolutionReason: reason,
    }

    var dbReq dbApprovalRequest
    // Use NamedQueryStruct — returns ErrDBNotFound if zero rows returned (i.e., already resolved)
    if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, resolveQuery, data, &dbReq); err != nil {
        if errors.Is(err, sqldb.ErrDBNotFound) {
            return approvalrequestbus.ApprovalRequest{}, approvalrequestbus.ErrAlreadyResolved
        }
        return approvalrequestbus.ApprovalRequest{}, fmt.Errorf("resolve: %w", err)
    }
    return toBusApprovalRequest(dbReq)
}
```

**Business-level `Resolve`** — thin wrapper (no read-then-write):
```go
func (b *Business) Resolve(ctx context.Context, id, resolvedBy uuid.UUID, status, reason string) (ApprovalRequest, error) {
    ctx, span := otel.AddSpan(ctx, "business.approvalrequestbus.resolve")
    defer span.End()

    req, err := b.storer.Resolve(ctx, id, resolvedBy, status, reason)
    if err != nil {
        return ApprovalRequest{}, fmt.Errorf("resolve approval request: %w", err)
    }

    return req, nil
}
```

**Why this is correct**: The `WHERE status = 'pending'` clause makes the UPDATE atomic — if two concurrent resolves race, only the first gets a row back via `RETURNING *`. The second gets zero rows, which `NamedQueryStruct` surfaces as `ErrDBNotFound`, mapped to `ErrAlreadyResolved`. No transaction needed.

**Error variables:**
```go
var (
    ErrNotFound        = errors.New("approval request not found")
    ErrAlreadyResolved = errors.New("approval request already resolved")
    ErrNotApprover     = errors.New("user is not an approver for this request")
)
```

#### stores/approvalrequestdb/model.go

DB struct uses `db:` tags for sqlx and handles PostgreSQL arrays:

```go
import (
    "github.com/timmaaaz/ichor/business/sdk/sqldb/dbarray"
)

type dbApprovalRequest struct {
    ID               uuid.UUID      `db:"approval_request_id"`
    ExecutionID      uuid.UUID      `db:"execution_id"`
    RuleID           uuid.UUID      `db:"rule_id"`
    ActionName       string         `db:"action_name"`
    Approvers        dbarray.String `db:"approvers"`       // UUID[] stored as text array via dbarray
    ApprovalType     string         `db:"approval_type"`
    Status           string         `db:"status"`
    TimeoutHours     int            `db:"timeout_hours"`
    TaskToken        string         `db:"task_token"`
    ApprovalMessage  sql.NullString `db:"approval_message"`
    ResolvedBy       sql.NullString `db:"resolved_by"`     // nullable UUID
    ResolutionReason sql.NullString `db:"resolution_reason"`
    CreatedDate      time.Time      `db:"created_date"`
    ResolvedDate     sql.NullTime   `db:"resolved_date"`
}
```

**Array handling**: Use `dbarray.String` (from `github.com/timmaaaz/ichor/business/sdk/sqldb/dbarray`) — this codebase does NOT vendor `github.com/lib/pq`. `dbarray.String` implements `driver.Valuer` and `sql.Scanner` for PostgreSQL text arrays.

**UUID↔string conversion in model converters:**
```go
func toDBApprovalRequest(req approvalrequestbus.ApprovalRequest) dbApprovalRequest {
    // Convert []uuid.UUID → dbarray.String
    approvers := make(dbarray.String, len(req.Approvers))
    for i, id := range req.Approvers {
        approvers[i] = id.String()
    }
    return dbApprovalRequest{
        // ...
        Approvers: approvers,
    }
}

func toBusApprovalRequest(db dbApprovalRequest) (approvalrequestbus.ApprovalRequest, error) {
    // Convert dbarray.String → []uuid.UUID
    approvers := make([]uuid.UUID, len(db.Approvers))
    for i, s := range db.Approvers {
        id, err := uuid.Parse(s)
        if err != nil {
            return approvalrequestbus.ApprovalRequest{}, fmt.Errorf("parse approver UUID at index %d: %w", i, err)
        }
        approvers[i] = id
    }
    return approvalrequestbus.ApprovalRequest{
        // ...
        Approvers: approvers,
    }, nil
}
```

**IsApprover query**: Named parameters with `ARRAY[:user_id]::uuid[]` won't work with sqlx named queries. Use a raw positional parameter instead:
```go
// In approvalrequestdb.go — IsApprover uses a raw query (not named params)
const isApproverQuery = `SELECT EXISTS(SELECT 1 FROM workflow.approval_requests WHERE approval_request_id = $1 AND $2::uuid = ANY(approvers))`

func (s *Store) IsApprover(ctx context.Context, approvalID, userID uuid.UUID) (bool, error) {
    var exists bool
    if err := s.db.QueryRowContext(ctx, isApproverQuery, approvalID, userID).Scan(&exists); err != nil {
        return false, fmt.Errorf("is approver: %w", err)
    }
    return exists, nil
}
```

#### filter.go

```go
type QueryFilter struct {
    ID          *uuid.UUID
    ExecutionID *uuid.UUID
    RuleID      *uuid.UUID
    Status      *string
    ApproverID  *uuid.UUID  // Filter by approver (uses @> array contains)
}
```

The `ApproverID` filter enables the "mine" endpoint: `WHERE approvers @> ARRAY[:approver_id]::uuid[]`

#### order.go

Default order: `created_date DESC` (newest first).

Fields: `created_date`, `status`, `approval_type`.

**Validation:**
- `go build ./business/domain/workflow/approvalrequestbus/...`
- Unit tests for `toDBApprovalRequest`/`toBusApprovalRequest` conversions

---

### Task 4: Implement `SeekApprovalHandler.StartAsync`

**Files to modify:**

| File | Change |
|------|--------|
| `business/sdk/workflow/workflowactions/approval/seek.go` | Add dependencies, implement `StartAsync`, update constructor |
| `business/sdk/workflow/workflowactions/register.go` | Add `ApprovalRequest` to `BusDependencies`, update `RegisterAll` and `RegisterCoreActions` |
| `business/sdk/workflow/models.go` | Add `ActionName` field to `ActionExecutionContext` |
| `business/sdk/workflow/temporal/activities.go` | Populate `ActionName` in `buildExecContext()` |
| `api/sdk/http/apitest/workflow.go` | Update to match new constructor signature |
| `api/cmd/tooling/admin/commands/validateworkflows.go` | Update to match new constructor signature |

#### Updated Handler Struct

```go
type SeekApprovalHandler struct {
    log                *logger.Logger
    db                 *sqlx.DB
    approvalRequestBus *approvalrequestbus.Business
    alertBus           *alertbus.Business
}

func NewSeekApprovalHandler(
    log *logger.Logger,
    db *sqlx.DB,
    approvalRequestBus *approvalrequestbus.Business,
    alertBus *alertbus.Business,
) *SeekApprovalHandler
```

**Constructor nil guards**: `approvalRequestBus` and `alertBus` can be nil (graceful degradation for `RegisterCoreActions` path where buses aren't available).

#### IsAsync Change

```go
func (h *SeekApprovalHandler) IsAsync() bool {
    return true // Now routes through ExecuteAsyncActionActivity
}
```

#### StartAsync Implementation

```go
func (h *SeekApprovalHandler) StartAsync(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext, taskToken []byte) error {
    // 1. Parse and validate config
    // 2. Parse approver UUIDs
    // 3. Log warning if approval_type != "any" (only "any" implemented)
    // 4. Persist approval request via approvalRequestBus.Create
    // 5. Create ONE alert + per-approver AlertRecipients (following CreateAlertHandler pattern)
    // 6. Return nil (Temporal holds activity open via ErrResultPending)
}
```

**Critical: Alert notification must follow the CreateAlertHandler pattern:**
```go
// ONE alert per approval request
alert := alertbus.Alert{
    ID:               uuid.New(),
    AlertType:        "approval_request",
    Severity:         alertbus.SeverityHigh,
    Title:            resolveTemplateVars(fmt.Sprintf("Approval Required: %s", execCtx.RuleName), execCtx.RawData),
    Message:          resolveTemplateVars(cfg.ApprovalMessage, execCtx.RawData),
    SourceEntityName: execCtx.EntityName,
    SourceEntityID:   execCtx.EntityID,
    SourceRuleID:     sourceRuleID,
    Status:           alertbus.StatusActive,
    CreatedDate:      now,
    UpdatedDate:      now,
}
h.alertBus.Create(ctx, alert)

// Per-approver recipients
var recipients []alertbus.AlertRecipient
for _, approverID := range approvers {
    recipients = append(recipients, alertbus.AlertRecipient{
        ID:            uuid.New(),
        AlertID:       alert.ID,
        RecipientType: "user",
        RecipientID:   approverID,
        CreatedDate:   now,
    })
}
h.alertBus.CreateRecipients(ctx, recipients)
```

#### ActionName in ActionExecutionContext

Add field to `models.go`:
```go
type ActionExecutionContext struct {
    // ... existing fields ...
    ActionName    string `json:"action_name"`
}
```

Populate in `buildExecContext()` (activities.go):
```go
execCtx := workflow.ActionExecutionContext{
    // ... existing fields ...
    ActionName:    input.ActionName,
}
```

#### RegisterCoreActions Update

`RegisterCoreActions` explicitly passes `nil, nil` for the two new bus parameters:
```go
// RegisterCoreActions — buses not available in core path (test/visualization only)
registry.Register(approval.NewSeekApprovalHandler(log, db, nil, nil))
```

**`StartAsync` must nil-guard both bus parameters.** If called without them (which happens when only `RegisterCoreActions` was used), return a clear error:
```go
func (h *SeekApprovalHandler) StartAsync(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext, taskToken []byte) error {
    if h.approvalRequestBus == nil || h.alertBus == nil {
        return fmt.Errorf("seek_approval requires approval request bus and alert bus (not available in core registration)")
    }
    // ... rest of implementation
}
```

This is acceptable because `RegisterCoreActions` only populates the sync `ActionRegistry`, not `AsyncRegistry`. The async path (`ExecuteAsyncActionActivity`) requires the worker, which uses `RegisterAll` with real bus instances.

#### RegisterAll Update

```go
// Approval actions — need approval request bus + alert bus
registry.Register(approval.NewSeekApprovalHandler(
    config.Log,
    config.DB,
    config.Buses.ApprovalRequest,
    config.Buses.Alert,
))
```

Add to `BusDependencies`:
```go
type BusDependencies struct {
    // ... existing fields ...
    ApprovalRequest *approvalrequestbus.Business
}
```

#### Execute Fallback

Keep `Execute()` as a fallback for manual execution. Update it to create a real approval request (same as `StartAsync` minus the task token):
```go
func (h *SeekApprovalHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
    if h.approvalRequestBus == nil {
        // Graceful degradation: return stub result (same as current behavior)
        return map[string]any{"output": "approved", "status": "stub"}, nil
    }
    // Create approval request with empty task token (manual execution, no Temporal)
    // Return {"approval_id": "...", "output": "pending", "status": "pending"}
}
```

**Validation:**
- `go build ./business/sdk/workflow/...`
- `go build ./api/...`
- Unit tests for `StartAsync` with mock `approvalrequestbus` and `alertbus`

---

### Task 5: Add Approval API Layer

**New package**: `api/domain/http/workflow/approvalapi/`

Follow the `alertapi` pattern exactly.

**File structure:**
```
api/domain/http/workflow/approvalapi/
├── approvalapi.go   -- Handler implementations
├── route.go         -- Config struct, Routes function
└── model.go         -- Request/response types, toApp converters
```

#### route.go

```go
package approvalapi

type Config struct {
    Log               *logger.Logger
    ApprovalBus       *approvalrequestbus.Business
    UserRoleBus       *userrolebus.Business
    AuthClient        *authclient.Client
    PermissionsBus    *permissionsbus.Business
    AsyncCompleter    *temporal.AsyncCompleter
}

const RouteTable = "workflow.approval_requests"

func Routes(app *web.App, cfg Config) {
    const version = "v1"
    api := newAPI(cfg)
    authen := mid.Authenticate(cfg.AuthClient)

    // Approver endpoints — auth only, business layer filters by approver
    app.HandlerFunc(http.MethodGet, version, "/workflow/approvals/mine", api.queryMine, authen)

    // Resolve endpoint — auth only, business layer checks approver/admin
    app.HandlerFunc(http.MethodPost, version, "/workflow/approvals/{id}/resolve", api.resolve, authen)

    // Single lookup
    app.HandlerFunc(http.MethodGet, version, "/workflow/approvals/{id}", api.queryByID, authen)

    // Admin endpoint — requires admin permission
    app.HandlerFunc(http.MethodGet, version, "/workflow/approvals", api.query, authen,
        mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))
}
```

#### approvalapi.go — Resolve Handler

This is the most critical endpoint. It must:
1. Parse approval request ID from path
2. Parse resolution body (`{"resolution": "approved"|"rejected", "reason": "..."}`)
3. Get authenticated user + roles
4. **Authorization check**: verify user is approver OR admin
5. Call `approvalBus.Resolve()` (which checks status == "pending")
6. Decode base64 task token
7. Construct `ActionActivityOutput` with correct `Result["output"]` key
8. Call `asyncCompleter.Complete(ctx, taskToken, output)`
9. Return resolved approval request

```go
func (a *api) resolve(ctx context.Context, r *http.Request) web.Encoder {
    id, err := uuid.Parse(web.Param(r, "id"))
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    userID, err := mid.GetUserID(ctx)
    if err != nil {
        return errs.New(errs.Unauthenticated, err)
    }

    var req ResolveRequest
    if err := web.Decode(r, &req); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    // Validate resolution value
    if req.Resolution != "approved" && req.Resolution != "rejected" {
        return errs.Newf(errs.InvalidArgument, "resolution must be 'approved' or 'rejected'")
    }

    // Authorization: check if user is approver or admin
    isApprover, err := a.approvalBus.IsApprover(ctx, id, userID)
    if err != nil {
        if errors.Is(err, approvalrequestbus.ErrNotFound) {
            return errs.New(errs.NotFound, err)
        }
        return errs.Newf(errs.Internal, "check approver: %s", err)
    }

    if !isApprover {
        // Check admin role
        roleIDs, err := a.getUserRoleIDs(ctx, userID)
        if err != nil {
            return errs.Newf(errs.Internal, "get user roles: %s", err)
        }
        if !a.isAdmin(roleIDs) {
            return errs.New(errs.PermissionDenied, approvalrequestbus.ErrNotApprover)
        }
    }

    // Resolve in DB (checks pending status, returns ErrAlreadyResolved if not pending)
    approval, err := a.approvalBus.Resolve(ctx, id, userID, req.Resolution, req.Reason)
    if err != nil {
        if errors.Is(err, approvalrequestbus.ErrAlreadyResolved) {
            return errs.Newf(errs.FailedPrecondition, "approval already resolved")
        }
        return errs.Newf(errs.Internal, "resolve: %s", err)
    }

    // Complete the Temporal activity
    taskToken, err := base64.StdEncoding.DecodeString(approval.TaskToken)
    if err != nil {
        return errs.Newf(errs.Internal, "decode task token: %s", err)
    }

    output := temporal.ActionActivityOutput{
        ActionID:   uuid.Nil, // Not available at this layer
        ActionName: approval.ActionName,
        Result: map[string]any{
            "output":      req.Resolution, // "approved" or "rejected" — drives edge selection
            "approval_id": approval.ID.String(),
            "resolved_by": userID.String(),
            "reason":      req.Reason,
        },
        Success: true,
    }

    if err := a.asyncCompleter.Complete(ctx, taskToken, output); err != nil {
        // Log but don't fail — the DB is already updated
        a.log.Error(ctx, "failed to complete Temporal activity",
            "approval_id", id,
            "error", err)
    }

    return toAppApproval(approval)
}
```

**Critical: `Result["output"]` drives branching.** The graph executor's `GetNextActions()` (graph_executor.go:104) extracts `result["output"]` to match against edge `SourceOutput` values. If `output` is `"approved"`, edges with `SourceOutput = "approved"` are followed. If `"rejected"`, edges with `SourceOutput = "rejected"` are followed.

#### model.go

```go
type ResolveRequest struct {
    Resolution string `json:"resolution" validate:"required"` // "approved" | "rejected"
    Reason     string `json:"reason"`                         // optional
}

type AppApproval struct {
    ID               string   `json:"id"`
    ExecutionID      string   `json:"execution_id"`
    RuleID           string   `json:"rule_id"`
    ActionName       string   `json:"action_name"`
    Approvers        []string `json:"approvers"`
    ApprovalType     string   `json:"approval_type"`
    Status           string   `json:"status"`
    TimeoutHours     int      `json:"timeout_hours"`
    ApprovalMessage  string   `json:"approval_message"`
    ResolvedBy       string   `json:"resolved_by,omitempty"`
    ResolutionReason string   `json:"resolution_reason,omitempty"`
    CreatedDate      string   `json:"created_date"`
    ResolvedDate     string   `json:"resolved_date,omitempty"`
}
```

#### queryMine Handler

```go
func (a *api) queryMine(ctx context.Context, r *http.Request) web.Encoder {
    userID, _ := mid.GetUserID(ctx)

    filter := approvalrequestbus.QueryFilter{
        ApproverID: &userID,
    }
    // Parse optional ?status=pending query param
    if status := r.URL.Query().Get("status"); status != "" {
        filter.Status = &status
    }

    pg, err := page.Parse(r)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    orderBy, err := order.Parse(orderByFields, r, defaultOrderBy)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    items, err := a.approvalBus.Query(ctx, filter, orderBy, pg)
    if err != nil {
        return errs.Newf(errs.Internal, "query: %s", err)
    }

    // Count with same filter for pagination total
    total, err := a.approvalBus.Count(ctx, filter)
    if err != nil {
        return errs.Newf(errs.Internal, "count: %s", err)
    }

    return query.NewResult(toAppApprovals(items), total, pg)
}
```

**Validation:**
- `go build ./api/domain/http/workflow/approvalapi/...`
- All endpoints return correct HTTP status codes
- Resolve endpoint produces correct `ActionActivityOutput` for edge resolution

---

### Task 6: Wire Everything Together

**Files to modify:**

| File | Change |
|------|--------|
| `api/cmd/services/ichor/build/all/all.go` | Instantiate `approvalrequestbus`, wire `approvalapi.Routes`, update `BusDependencies` |
| `api/cmd/services/workflow-worker/main.go` | Instantiate `approvalrequestbus`, populate `AsyncRegistry`, create `AsyncCompleter` |
| `api/sdk/http/apitest/workflow.go` | Update `InitWorkflowInfra` to register `seek_approval` in `AsyncRegistry` |
| `api/cmd/tooling/admin/commands/validateworkflows.go` | Update constructor call site |

#### all.go Changes

```go
// Near existing alertBus instantiation:
approvalRequestBus := approvalrequestbus.NewBusiness(cfg.Log, approvalrequestdb.NewStore(cfg.Log, cfg.DB))

// Update BusDependencies:
Buses: workflowactions.BusDependencies{
    // ... existing ...
    ApprovalRequest: approvalRequestBus,
}

// Add approvalapi routes:
approvalapi.Routes(app, approvalapi.Config{
    Log:            cfg.Log,
    ApprovalBus:    approvalRequestBus,
    UserRoleBus:    a.UserRoleBus,
    AuthClient:     a.AuthClient,
    PermissionsBus: a.PermissionsBus,
    AsyncCompleter: nil, // API service doesn't have Temporal client for completion
})
```

**Important**: The API service (`all.go`) does NOT have an `AsyncCompleter` — it has a Temporal client for dispatching workflows, not for completing activities. The `AsyncCompleter` lives in the worker process.

**Resolution**: The resolve endpoint needs `AsyncCompleter`. Two options:
1. **Create `AsyncCompleter` in `all.go`** from the existing Temporal client (the client can both dispatch and complete activities)
2. **Move resolve endpoint to worker HTTP server** (adds complexity)

Option 1 is correct: `client.Client` satisfies `ActivityCompleter` interface. Create `AsyncCompleter` in `all.go`:
```go
var asyncCompleter *temporalpkg.AsyncCompleter
if cfg.TemporalClient != nil {
    asyncCompleter = temporalpkg.NewAsyncCompleter(cfg.TemporalClient)
}
```

#### Worker Changes

```go
// Instantiate approval request bus
approvalRequestBus := approvalrequestbus.NewBusiness(log, approvalrequestdb.NewStore(log, db))

// Create async registry with seek_approval
asyncRegistry := temporal.NewAsyncRegistry()
asyncRegistry.Register("seek_approval", approval.NewSeekApprovalHandler(log, db, approvalRequestBus, alertBus))

// Register activities with BOTH registries
w.RegisterActivity(&temporal.Activities{
    Registry:      actionRegistry,
    AsyncRegistry: asyncRegistry,
})
```

#### apitest/workflow.go Changes

**Scope clarification**: Integration tests will NOT exercise the full async `seek_approval` path end-to-end (that requires a real approval DB + Temporal async completion). The `apitest/workflow.go` change is **constructor-signature-only** — pass `nil` for the new bus parameters. The `StartAsync` unit tests (Task 4) cover the handler logic in isolation.

```go
// In InitWorkflowInfra — constructor signature update only, nil buses
asyncRegistry := temporal.NewAsyncRegistry()
asyncRegistry.Register("seek_approval", approval.NewSeekApprovalHandler(log, db, nil, nil))

w.RegisterActivity(&temporal.Activities{
    Registry:      actionRegistry,
    AsyncRegistry: asyncRegistry,
})
```

If `seek_approval` is triggered in an integration test, `StartAsync` will return the nil-guard error ("seek_approval requires approval request bus...") which is expected and acceptable for constructor-only tests.

**Validation:**
- `go build ./...`
- `go vet ./...`
- Worker binary starts with populated `AsyncRegistry`
- API service starts with `approvalapi` routes registered

---

## Testing Strategy

### Unit Tests

| Test File | What It Tests |
|-----------|---------------|
| `business/domain/workflow/approvalrequestbus/stores/approvalrequestdb/model_test.go` | DB ↔ bus model conversion, UUID array handling |
| `business/sdk/workflow/workflowactions/approval/seek_test.go` | `StartAsync` with mock bus, `Validate`, `Execute` fallback |
| `business/sdk/workflow/temporal/workflow_test.go` | `selectActivityFunc` now routes human types to async |

### Integration Tests

Using `apitest.WorkflowInfra` pattern:

1. **Seed a workflow rule** with `seek_approval` action + conditional edges (approved → action_A, rejected → action_B)
2. **Trigger the workflow** via `wf.WorkflowTrigger.OnEntityEvent`
3. **Assert approval_request created in DB** (query by execution_id)
4. **Call resolve endpoint** with `{"resolution": "approved"}`
5. **Assert Temporal workflow resumed** and followed the `approved` branch
6. **Assert approval_request status updated** to `"approved"` in DB

### Test File Location

Integration tests: `api/cmd/services/ichor/tests/workflow/approvalapi/`

---

## Gotchas

1. **`ActionName` not on `ActionExecutionContext`** — CONFIRMED missing. Must add field to `models.go` and populate in `buildExecContext()` from `input.ActionName`. Without this, the approval request has no way to record which action within a rule is seeking approval.

2. **Task token is `[]byte` (binary)** — Base64-encode before storing in DB TEXT column. Decode before calling `AsyncCompleter.Complete()`.

3. **Double-submit guard** — The `Resolve` method in `approvalrequestbus` must check `status == "pending"` before updating. If two approvers race, only the first succeeds. Temporal also guards: only the first `CompleteActivity` call for a task token succeeds; subsequent calls return an error that should be logged but not returned to the user.

4. **`majority`/`all` approval types** — Accepted by schema CHECK constraint but treated as `"any"`. Log a warning in `StartAsync` when `approval_type != "any"`. Future phase adds `approval_votes` table for per-approver tracking.

5. **Timeout handling** — `timeout_hours` is stored but NOT enforced. No automatic expiry mechanism. Future phase: add a Temporal timer that fires `AsyncCompleter.Fail()` or `Complete(timed_out)` after the timeout period.

6. **`AsyncCompleter` location** — Created from the Temporal client in BOTH `all.go` (for the resolve API endpoint) AND `worker/main.go` (for potential future use). The Temporal `client.Client` satisfies the `ActivityCompleter` interface for both dispatching and completing.

7. **`selectActivityFunc` change is global** — After modification, ALL entries in `humanActionTypes` will route to `ExecuteAsyncActionActivity`. Currently `manager_approval`, `manual_review`, `human_verification`, and `approval_request` are in this map but have no `AsyncActivityHandler` implementations registered. If a workflow uses these action types, the async registry lookup will fail with "no handler registered." This is acceptable because these action types have no production handlers — they're placeholder names. But document the risk.

8. **`RegisterCoreActions` passes nil** — The test/visualization path creates `SeekApprovalHandler` with nil buses. `StartAsync` will fail if called without `approvalRequestBus`. This is acceptable: `RegisterCoreActions` only populates the sync `ActionRegistry`, not `AsyncRegistry`. The async path requires the worker, which uses `RegisterAll`.

---

## File Change Summary

### New Files (6)
| File | Description |
|------|-------------|
| `business/domain/workflow/approvalrequestbus/model.go` | Business models |
| `business/domain/workflow/approvalrequestbus/approvalrequestbus.go` | Business + Storer interface |
| `business/domain/workflow/approvalrequestbus/filter.go` | Query filters |
| `business/domain/workflow/approvalrequestbus/order.go` | Order by constants |
| `business/domain/workflow/approvalrequestbus/stores/approvalrequestdb/approvalrequestdb.go` | DB store |
| `business/domain/workflow/approvalrequestbus/stores/approvalrequestdb/model.go` | DB models + converters |
| `business/domain/workflow/approvalrequestbus/stores/approvalrequestdb/filter.go` | DB filter |
| `business/domain/workflow/approvalrequestbus/stores/approvalrequestdb/order.go` | DB order |
| `api/domain/http/workflow/approvalapi/approvalapi.go` | API handlers |
| `api/domain/http/workflow/approvalapi/route.go` | Routes + Config |
| `api/domain/http/workflow/approvalapi/model.go` | API models |

### Modified Files (~10)
| File | Change |
|------|--------|
| `business/sdk/migrate/sql/migrate.sql` | Add version 1.997 |
| `business/sdk/workflow/models.go` | Add `ActionName` to `ActionExecutionContext` |
| `business/sdk/workflow/temporal/activities.go` | Populate `ActionName` in `buildExecContext` |
| `business/sdk/workflow/temporal/workflow.go` | Modify `selectActivityFunc`, add `seek_approval` to `humanActionTypes` |
| `business/sdk/workflow/temporal/allocate_inventory_async_test.go` | Update routing test |
| `business/sdk/workflow/workflowactions/approval/seek.go` | Full rewrite: new deps, `StartAsync`, updated `Execute` |
| `business/sdk/workflow/workflowactions/register.go` | Add `ApprovalRequest` to `BusDependencies`, update registrations |
| `api/cmd/services/ichor/build/all/all.go` | Wire `approvalrequestbus`, `approvalapi`, `AsyncCompleter` |
| `api/cmd/services/workflow-worker/main.go` | Wire `AsyncRegistry`, `approvalrequestbus` |
| `api/sdk/http/apitest/workflow.go` | Update for new constructor signature |
| `api/cmd/tooling/admin/commands/validateworkflows.go` | Update constructor call |

---

## Build Order

Execute tasks in this order (dependencies flow downward):

```
Task 1: Enable async routing (workflow.go)
    ↓
Task 2: Create migration (migrate.sql)
    ↓
Task 3: Create approvalrequestbus (business package)
    ↓
Task 4: Implement SeekApprovalHandler.StartAsync (handler + register.go)
    ↓
Task 5: Add approvalapi (API layer)
    ↓
Task 6: Wire everything (all.go, worker, apitest)
    ↓
go build ./... && go vet ./... && go test ./...
```

Tasks 1 and 2 are independent and can be done in parallel. Task 3 depends on Task 2 (needs migration for integration tests). Tasks 4-6 are sequential.

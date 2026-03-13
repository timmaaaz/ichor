# Approval Resolve Idempotency Tests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add tests for the idempotent `resolve` endpoint and `retryTemporalCompletion` logic added in the Temporal Async Completion Durability fix.

**Architecture:** Two tiers: (1) unit tests in `package approvalapi` that call `retryTemporalCompletion` directly via mock injection into the unexported `api` struct, and (2) integration tests that exercise the full HTTP resolve endpoint twice to prove HTTP 200 is returned instead of 412 on duplicate submissions.

**Tech Stack:** Go `testing`, `net/http/httptest`, `dbtest.NewDatabase`, `apitest.StartTest`, `workflow.TestSeedFullWorkflow`, `approvalrequestdb.NewStore`

---

## Background

The resolve endpoint (`POST /v1/workflow/approvals/{id}/resolve`) was made idempotent in a recent fix. The key behaviors to test:

1. `retryTemporalCompletion` — called when `ErrAlreadyResolved` fires, handles 4 cases:
   - Empty task token → no Temporal call, returns the resolved approval
   - Non-empty token + Complete succeeds → ClearTaskToken called, returns approval
   - Non-empty token + Complete fails → ClearTaskToken NOT called, returns approval
   - `asyncCompleter == nil` → skips Temporal, returns approval
2. Happy path token clearing — after a first successful resolve with a non-nil asyncCompleter, `ClearTaskToken` is called
3. Integration regression — double-submit returns HTTP 200, not 412

### Relevant files

- `api/domain/http/workflow/approvalapi/approvalapi.go` — the handler under test
- `api/domain/http/workflow/approvalapi/model.go` — `Approval` response type
- `business/domain/workflow/approvalrequestbus/approvalrequestbus.go` — `Storer` interface
- `business/domain/workflow/approvalrequestbus/stores/approvalrequestdb/approvalrequestdb.go` — real store
- `business/sdk/workflow/temporal/async_completer.go` — `ActivityCompleter` interface, `AsyncCompleter`
- `business/sdk/dbtest/dbtest.go` — test DB setup
- `api/sdk/http/apitest/apitest.go` — HTTP integration test helpers
- `business/sdk/workflow/seed_test_helpers.go` — `workflow.TestSeedFullWorkflow`

### Key constraint: context injection

`mid.GetUserID` uses an unexported context key (`ctxKey`) inside `app/sdk/mid`. There is no exported setter. This means `resolve` cannot be called from unit tests without going through the auth middleware. Therefore:

- **Unit tests** only call `a.retryTemporalCompletion()` directly (it does not use `mid.GetUserID`)
- **Integration tests** test the full resolve flow through the HTTP mux with a real JWT

---

## Task 1: Unit Tests for `retryTemporalCompletion`

**Files:**
- Create: `api/domain/http/workflow/approvalapi/resolve_test.go`

> Note: This file uses `package approvalapi` (NOT `approvalapi_test`) to access the unexported `api` struct.

### Step 1: Write the failing test file

Create `api/domain/http/workflow/approvalapi/resolve_test.go`:

```go
package approvalapi

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// Mock infrastructure

// mockStorer is a controllable Storer for unit tests.
type mockStorer struct {
	queryByIDResult approvalrequestbus.ApprovalRequest
	queryByIDErr    error
	clearCalled     bool
	clearErr        error
}

func (s *mockStorer) NewWithTx(_ sqldb.CommitRollbacker) (approvalrequestbus.Storer, error) {
	return s, nil
}
func (s *mockStorer) Create(_ context.Context, _ approvalrequestbus.ApprovalRequest) error {
	return nil
}
func (s *mockStorer) QueryByID(_ context.Context, _ uuid.UUID) (approvalrequestbus.ApprovalRequest, error) {
	return s.queryByIDResult, s.queryByIDErr
}
func (s *mockStorer) Resolve(_ context.Context, _, _ uuid.UUID, _, _ string) (approvalrequestbus.ApprovalRequest, error) {
	return approvalrequestbus.ApprovalRequest{}, nil
}
func (s *mockStorer) Query(_ context.Context, _ approvalrequestbus.QueryFilter, _ order.By, _ page.Page) ([]approvalrequestbus.ApprovalRequest, error) {
	return nil, nil
}
func (s *mockStorer) Count(_ context.Context, _ approvalrequestbus.QueryFilter) (int, error) {
	return 0, nil
}
func (s *mockStorer) IsApprover(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (s *mockStorer) ClearTaskToken(_ context.Context, _ uuid.UUID) error {
	s.clearCalled = true
	return s.clearErr
}

// mockActivityCompleter controls the Temporal CompleteActivity call.
type mockActivityCompleter struct {
	err    error
	called bool
}

func (m *mockActivityCompleter) CompleteActivity(_ context.Context, _ []byte, _ any, _ error) error {
	m.called = true
	return m.err
}

// validToken returns a base64-encoded fake task token.
func validToken() string {
	return base64.StdEncoding.EncodeToString([]byte("fake-temporal-task-token"))
}

// resolvedApproval builds a minimal ApprovalRequest in approved state.
func resolvedApproval(taskToken string) approvalrequestbus.ApprovalRequest {
	resolvedBy := uuid.New()
	resolvedDate := time.Now()
	return approvalrequestbus.ApprovalRequest{
		ID:               uuid.New(),
		ExecutionID:      uuid.New(),
		RuleID:           uuid.New(),
		ActionName:       "seek_approval_0",
		Approvers:        []uuid.UUID{uuid.New()},
		ApprovalType:     "any",
		Status:           approvalrequestbus.StatusApproved,
		TimeoutHours:     72,
		TaskToken:        taskToken,
		ResolvedBy:       &resolvedBy,
		ResolutionReason: "looks good",
		CreatedDate:      time.Now(),
		ResolvedDate:     &resolvedDate,
	}
}

// newTestAPI constructs an api with the given storer and completer.
func newTestAPI(storer approvalrequestbus.Storer, completer temporal.ActivityCompleter) *api {
	log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
	bus := approvalrequestbus.NewBusiness(log, storer)

	var asyncCompleter *temporal.AsyncCompleter
	if completer != nil {
		asyncCompleter = temporal.NewAsyncCompleter(completer)
	}

	return &api{
		log:            log,
		approvalBus:    bus,
		asyncCompleter: asyncCompleter,
	}
}

// =============================================================================
// Tests

func TestRetryTemporalCompletion_EmptyToken(t *testing.T) {
	// When task_token is empty, Temporal was already notified (true duplicate).
	// Expect: no Temporal call, approval returned as-is.
	approval := resolvedApproval("") // empty token

	storer := &mockStorer{queryByIDResult: approval}
	completer := &mockActivityCompleter{}
	a := newTestAPI(storer, completer)

	result := a.retryTemporalCompletion(context.Background(), approval.ID)

	// Must not return an error encoder.
	if _, isErr := result.(error); isErr {
		t.Fatalf("expected success encoder, got error: %v", result)
	}

	// Temporal must NOT have been called (token was empty).
	if completer.called {
		t.Fatal("Temporal CompleteActivity should not be called when token is empty")
	}

	// ClearTaskToken must NOT be called.
	if storer.clearCalled {
		t.Fatal("ClearTaskToken should not be called when token is already empty")
	}
}

func TestRetryTemporalCompletion_TokenPresentCompleteSucceeds(t *testing.T) {
	// When task_token is present and Complete() succeeds:
	// Expect: Temporal called, ClearTaskToken called, approval returned.
	approval := resolvedApproval(validToken())

	storer := &mockStorer{queryByIDResult: approval}
	completer := &mockActivityCompleter{err: nil}
	a := newTestAPI(storer, completer)

	result := a.retryTemporalCompletion(context.Background(), approval.ID)

	if _, isErr := result.(error); isErr {
		t.Fatalf("expected success encoder, got error: %v", result)
	}

	if !completer.called {
		t.Fatal("Temporal CompleteActivity should have been called")
	}

	if !storer.clearCalled {
		t.Fatal("ClearTaskToken should have been called after successful Complete")
	}
}

func TestRetryTemporalCompletion_TokenPresentCompleteFails(t *testing.T) {
	// When task_token is present but Complete() fails:
	// Expect: Temporal called, ClearTaskToken NOT called, approval still returned (fail-open).
	approval := resolvedApproval(validToken())

	storer := &mockStorer{queryByIDResult: approval}
	completer := &mockActivityCompleter{err: errors.New("temporal rpc timeout")}
	a := newTestAPI(storer, completer)

	result := a.retryTemporalCompletion(context.Background(), approval.ID)

	// Should still return the approval (not an error) — fail-open behavior.
	if _, isErr := result.(error); isErr {
		t.Fatalf("expected success encoder even on Temporal failure, got error: %v", result)
	}

	if !completer.called {
		t.Fatal("Temporal CompleteActivity should have been attempted")
	}

	// ClearTaskToken must NOT be called when Complete failed.
	if storer.clearCalled {
		t.Fatal("ClearTaskToken must not be called when Complete() failed")
	}
}

func TestRetryTemporalCompletion_NilCompleter(t *testing.T) {
	// When asyncCompleter is nil (Temporal not configured):
	// Expect: no panic, approval returned immediately.
	approval := resolvedApproval(validToken()) // token present but no completer

	storer := &mockStorer{queryByIDResult: approval}
	a := newTestAPI(storer, nil) // nil completer

	result := a.retryTemporalCompletion(context.Background(), approval.ID)

	if _, isErr := result.(error); isErr {
		t.Fatalf("expected success encoder, got error: %v", result)
	}

	if storer.clearCalled {
		t.Fatal("ClearTaskToken should not be called when asyncCompleter is nil")
	}
}

func TestRetryTemporalCompletion_QueryByIDFails(t *testing.T) {
	// When the DB lookup fails during retry:
	// Expect: internal error returned.
	storer := &mockStorer{
		queryByIDErr: errors.New("db connection lost"),
	}
	a := newTestAPI(storer, nil)

	result := a.retryTemporalCompletion(context.Background(), uuid.New())

	appErr, ok := result.(*errs.Error)
	if !ok {
		t.Fatalf("expected *errs.Error, got %T", result)
	}
	if appErr.Code != errs.Internal {
		t.Fatalf("expected Internal error code, got %v", appErr.Code)
	}
}
```

### Step 2: Run to verify it fails (missing types/methods will surface here)

```bash
go test ./api/domain/http/workflow/approvalapi/... -run TestRetry -v
```

Expected: compile error or FAIL (the test file is new, nothing exists yet)

### Step 3: Verify it compiles and passes

The implementation already exists from the durability fix. Run again:

```bash
go test ./api/domain/http/workflow/approvalapi/... -run TestRetry -v
```

Expected output:
```
--- PASS: TestRetryTemporalCompletion_EmptyToken (0.00s)
--- PASS: TestRetryTemporalCompletion_TokenPresentCompleteSucceeds (0.00s)
--- PASS: TestRetryTemporalCompletion_TokenPresentCompleteFails (0.00s)
--- PASS: TestRetryTemporalCompletion_NilCompleter (0.00s)
--- PASS: TestRetryTemporalCompletion_QueryByIDFails (0.00s)
ok  	github.com/timmaaaz/ichor/api/domain/http/workflow/approvalapi
```

### Step 4: Check `errs.Error` type assertion

The `*errs.Error` type check in the last test needs verification. Look at `app/sdk/errs` to confirm the concrete type returned by `errs.Newf`:

```bash
grep -n "type Error\|func Newf" /Users/jaketimmer/src/work/superior/ichor/ichor/app/sdk/errs/*.go | head -10
```

If the type is different (e.g., it's an interface), adjust the type assertion in `TestRetryTemporalCompletion_QueryByIDFails` accordingly.

### Step 5: Commit

```bash
git add api/domain/http/workflow/approvalapi/resolve_test.go
git commit -m "test(approvalapi): unit tests for retryTemporalCompletion idempotent retry logic"
```

---

## Task 2: Integration Tests for Double-Submit Regression

**Files:**
- Create: `api/cmd/services/ichor/tests/workflow/approvalapi/seed_test.go`
- Create: `api/cmd/services/ichor/tests/workflow/approvalapi/resolve_test.go`

These go through the full HTTP stack with a live DB container. They verify that double-submitting a resolve returns HTTP 200 instead of 412.

### Step 1: Write the seed file

Create `api/cmd/services/ichor/tests/workflow/approvalapi/seed_test.go`:

```go
package approvalapi_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ApproveSeedData holds test-specific state for resolve tests.
type ApproveSeedData struct {
	apitest.SeedData

	// ApprovalID is the pending approval request seeded for testing.
	ApprovalID uuid.UUID

	// ApprovalWithTokenID is a pending approval request with a non-empty task_token.
	// Used to test retry path behavior.
	ApprovalWithTokenID uuid.UUID
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (ApproveSeedData, error) {
	ctx := context.Background()

	// -------------------------------------------------------------------------
	// Seed users

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		return ApproveSeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	admin := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// -------------------------------------------------------------------------
	// Seed workflow data (provides valid execution_id and rule_id FKs)

	wfData, err := workflow.TestSeedFullWorkflow(ctx, admins[0].ID, db.BusDomain.Workflow)
	if err != nil {
		return ApproveSeedData{}, fmt.Errorf("seeding workflow: %w", err)
	}

	executionID1 := wfData.AutomationExecutions[0].ID
	executionID2 := wfData.AutomationExecutions[1].ID
	ruleID := wfData.AutomationRules[0].ID

	// -------------------------------------------------------------------------
	// Create approval requests directly via the store

	approvalStore := approvalrequestdb.NewStore(db.Log, db.DB)
	approvalBus := approvalrequestbus.NewBusiness(db.Log, approvalStore)

	// Approval 1: no task_token (simulates manual/non-Temporal approval)
	req1 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     executionID1,
		RuleID:          ruleID,
		ActionName:      "seek_approval_0",
		Approvers:       []uuid.UUID{admins[0].ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusPending,
		TimeoutHours:    72,
		TaskToken:       "",
		ApprovalMessage: "Please approve",
		CreatedDate:     time.Now(),
	}
	if err := approvalBus.(*approvalrequestbus.Business); err != nil {
		// approvalBus is already *approvalrequestbus.Business from NewBusiness
	}

	// NOTE: Create method is on the Business directly; call through the bus.
	// We build the approval manually since there's no NewApprovalRequest helper
	// that skips UUID generation — use the store directly.
	if err := approvalStore.Create(ctx, req1); err != nil {
		return ApproveSeedData{}, fmt.Errorf("creating approval 1: %w", err)
	}

	// Approval 2: with task_token (simulates Temporal-backed approval)
	fakeToken := base64.StdEncoding.EncodeToString([]byte("temporal-task-token-abc"))
	req2 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     executionID2,
		RuleID:          ruleID,
		ActionName:      "seek_approval_1",
		Approvers:       []uuid.UUID{admins[0].ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusPending,
		TimeoutHours:    72,
		TaskToken:       fakeToken,
		ApprovalMessage: "Please approve",
		CreatedDate:     time.Now(),
	}
	if err := approvalStore.Create(ctx, req2); err != nil {
		return ApproveSeedData{}, fmt.Errorf("creating approval 2: %w", err)
	}

	return ApproveSeedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.User{admin},
		},
		ApprovalID:          req1.ID,
		ApprovalWithTokenID: req2.ID,
	}, nil
}
```

> **Note:** The seed file has a compile error placeholder (`if err := approvalBus.(*approvalrequestbus.Business); err != nil`). Remove that block — it was a mistake. See corrected version in Step 2.

### Step 2: Correct the seed file (remove spurious block)

The seed file should use `approvalStore.Create` directly (already done). Remove the block that type-asserts `approvalBus`. The correct seeding pattern is simply:

```go
approvalStore := approvalrequestdb.NewStore(db.Log, db.DB)

if err := approvalStore.Create(ctx, req1); err != nil { ... }
if err := approvalStore.Create(ctx, req2); err != nil { ... }
```

Final `insertSeedData` function (corrected):

```go
func insertSeedData(db *dbtest.Database, ath *auth.Auth) (ApproveSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		return ApproveSeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	admin := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	wfData, err := workflow.TestSeedFullWorkflow(ctx, admins[0].ID, db.BusDomain.Workflow)
	if err != nil {
		return ApproveSeedData{}, fmt.Errorf("seeding workflow: %w", err)
	}

	executionID1 := wfData.AutomationExecutions[0].ID
	executionID2 := wfData.AutomationExecutions[1].ID
	ruleID := wfData.AutomationRules[0].ID

	approvalStore := approvalrequestdb.NewStore(db.Log, db.DB)

	req1 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     executionID1,
		RuleID:          ruleID,
		ActionName:      "seek_approval_0",
		Approvers:       []uuid.UUID{admins[0].ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusPending,
		TimeoutHours:    72,
		TaskToken:       "",
		ApprovalMessage: "Please approve",
		CreatedDate:     time.Now(),
	}
	if err := approvalStore.Create(ctx, req1); err != nil {
		return ApproveSeedData{}, fmt.Errorf("creating approval: %w", err)
	}

	fakeToken := base64.StdEncoding.EncodeToString([]byte("temporal-task-token-abc"))
	req2 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     executionID2,
		RuleID:          ruleID,
		ActionName:      "seek_approval_1",
		Approvers:       []uuid.UUID{admins[0].ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusPending,
		TimeoutHours:    72,
		TaskToken:       fakeToken,
		ApprovalMessage: "Please approve with token",
		CreatedDate:     time.Now(),
	}
	if err := approvalStore.Create(ctx, req2); err != nil {
		return ApproveSeedData{}, fmt.Errorf("creating approval with token: %w", err)
	}

	return ApproveSeedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.User{admin},
		},
		ApprovalID:          req1.ID,
		ApprovalWithTokenID: req2.ID,
	}, nil
}
```

### Step 3: Write the resolve test file

Create `api/cmd/services/ichor/tests/workflow/approvalapi/resolve_test.go`:

```go
package approvalapi_test

import (
	"net/http"
	"testing"

	"github.com/timmaaaz/ichor/api/domain/http/workflow/approvalapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_ApprovalResolve(t *testing.T) {
	at := apitest.StartTest(t, "Test_ApprovalResolve")

	sd, err := insertSeedData(at.DB, at.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	at.Run(t, resolveTests(at, sd), "resolve")
}

func resolveTests(at *apitest.Test, sd ApproveSeedData) []apitest.Table {
	adminToken := sd.Admins[0].Token
	approvalURL := "/v1/workflow/approvals/" + sd.ApprovalID.String() + "/resolve"
	approvalWithTokenURL := "/v1/workflow/approvals/" + sd.ApprovalWithTokenID.String() + "/resolve"

	resolveBody := approvalapi.ResolveRequest{
		Resolution: "approved",
		Reason:     "looks good",
	}

	return []apitest.Table{
		{
			Name:       "first-resolve-returns-200",
			URL:        approvalURL,
			Token:      adminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      resolveBody,
			GotResp:    &approvalapi.Approval{},
			ExpResp: &approvalapi.Approval{
				Status: "approved",
			},
			CmpFunc: func(got, exp any) string {
				gotApp := got.(*approvalapi.Approval)
				expApp := exp.(*approvalapi.Approval)
				if gotApp.Status != expApp.Status {
					return "expected status=approved, got " + gotApp.Status
				}
				return ""
			},
		},
		{
			Name:       "double-resolve-returns-200-not-412",
			URL:        approvalURL, // same approval as above — already resolved
			Token:      adminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      resolveBody,
			GotResp:    &approvalapi.Approval{},
			ExpResp: &approvalapi.Approval{
				Status: "approved",
			},
			CmpFunc: func(got, exp any) string {
				gotApp := got.(*approvalapi.Approval)
				expApp := exp.(*approvalapi.Approval)
				if gotApp.Status != expApp.Status {
					return "expected status=approved on retry, got " + gotApp.Status
				}
				return ""
			},
		},
		{
			Name:       "resolve-with-task-token-returns-200",
			URL:        approvalWithTokenURL,
			Token:      adminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      resolveBody,
			GotResp:    &approvalapi.Approval{},
			ExpResp: &approvalapi.Approval{
				Status: "approved",
			},
			CmpFunc: func(got, exp any) string {
				gotApp := got.(*approvalapi.Approval)
				expApp := exp.(*approvalapi.Approval)
				if gotApp.Status != expApp.Status {
					return "expected status=approved, got " + gotApp.Status
				}
				return ""
			},
		},
		{
			Name:       "double-resolve-with-token-returns-200",
			URL:        approvalWithTokenURL, // already resolved above
			Token:      adminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      resolveBody,
			GotResp:    &approvalapi.Approval{},
			ExpResp: &approvalapi.Approval{
				Status: "approved",
			},
			CmpFunc: func(got, exp any) string {
				gotApp := got.(*approvalapi.Approval)
				expApp := exp.(*approvalapi.Approval)
				if gotApp.Status != expApp.Status {
					return "expected status=approved on retry with token, got " + gotApp.Status
				}
				return ""
			},
		},
	}
}
```

> **Important**: The `apitest.Table.Run` method iterates in order. The table entries are intentionally ordered: first resolve, then double-resolve. This means the first entry's side-effect (resolving the approval) is a prerequisite for the second entry to exercise the retry path.

### Step 4: Run the integration tests to verify they fail

```bash
go test ./api/cmd/services/ichor/tests/workflow/approvalapi/... -v
```

Expected: compile errors (package doesn't exist yet, types need to be correct)

### Step 5: Fix any compile errors

Common issues to look out for:
- `approvalapi.ResolveRequest` may need to be exported — check `model.go`. It is already exported.
- `workflow.TestSeedFullWorkflow` signature — confirm it takes `(ctx, userID, wfBusDomain)`. Check: `business/sdk/workflow/` for `TestSeedFullWorkflow`.
- The `Workflow` field on `db.BusDomain` — verify the field name by checking `dbtest.go`'s `BusDomain` struct.

To check these:
```bash
grep -n "TestSeedFullWorkflow\|BusDomain" /Users/jaketimmer/src/work/superior/ichor/ichor/business/sdk/workflow/workflowactions/approval/seek_test.go | head -10
grep -n "Workflow\b" /Users/jaketimmer/src/work/superior/ichor/ichor/business/sdk/dbtest/dbtest.go | head -10
```

### Step 6: Run the integration tests to verify they pass

```bash
go test ./api/cmd/services/ichor/tests/workflow/approvalapi/... -v -timeout 120s
```

Expected:
```
--- PASS: Test_ApprovalResolve/resolve-first-resolve-returns-200 (0.XXs)
--- PASS: Test_ApprovalResolve/resolve-double-resolve-returns-200-not-412 (0.XXs)
--- PASS: Test_ApprovalResolve/resolve-resolve-with-task-token-returns-200 (0.XXs)
--- PASS: Test_ApprovalResolve/resolve-double-resolve-with-token-returns-200 (0.XXs)
ok  github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/approvalapi
```

### Step 7: Commit

```bash
git add api/cmd/services/ichor/tests/workflow/approvalapi/
git commit -m "test(approvalapi): integration tests for idempotent resolve endpoint"
```

---

## Task 3: Build verification

Run a full build of all affected packages to ensure nothing is broken:

```bash
go build ./api/domain/http/workflow/approvalapi/... && \
go build ./business/domain/workflow/approvalrequestbus/... && \
go build ./api/cmd/services/ichor/tests/workflow/approvalapi/... && \
echo "all packages build OK"
```

Expected: `all packages build OK`

---

## What is NOT tested (and why)

| Scenario | Reason not tested |
|---|---|
| Happy-path `ClearTaskToken` after `Complete` succeeds | Requires `asyncCompleter != nil` in integration test, but Temporal client is nil in tests. Covered by unit test for `retryTemporalCompletion` instead. |
| Expired task token (7-day Temporal timeout) | External Temporal state, not testable without a live Temporal instance. The code path logs and returns approval — covered by `TestRetryTemporalCompletion_TokenPresentCompleteFails`. |
| `ClearTaskToken` DB error | Low priority; logs the error and continues. Can add to unit tests as a low-risk extension. |

---

## Summary

| Task | File | Tests added |
|---|---|---|
| Unit | `api/domain/http/workflow/approvalapi/resolve_test.go` | 5 unit tests for `retryTemporalCompletion` |
| Integration | `api/cmd/services/ichor/tests/workflow/approvalapi/seed_test.go` | Seeding helpers |
| Integration | `api/cmd/services/ichor/tests/workflow/approvalapi/resolve_test.go` | 4 HTTP tests (first resolve + double-submit x2 approval types) |

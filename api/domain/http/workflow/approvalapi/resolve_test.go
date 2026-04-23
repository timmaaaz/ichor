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
	err         error
	called      bool
	capturedOut any
}

func (m *mockActivityCompleter) CompleteActivity(_ context.Context, _ []byte, out any, _ error) error {
	m.called = true
	m.capturedOut = out
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
	bus := approvalrequestbus.NewBusiness(log, nil, storer)

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
	// Payload must carry the same keys as the primary resolve path so
	// downstream workflow steps see consistent context regardless of path.
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

	out, ok := completer.capturedOut.(temporal.ActionActivityOutput)
	if !ok {
		t.Fatalf("expected temporal.ActionActivityOutput, got %T", completer.capturedOut)
	}
	for _, key := range []string{"output", "approval_id", "resolved_by", "reason"} {
		if _, present := out.Result[key]; !present {
			t.Errorf("retry payload missing key %q; primary resolve path sets it", key)
		}
	}
	if got := out.Result["resolved_by"]; got != approval.ResolvedBy.String() {
		t.Errorf("resolved_by = %v, want %s", got, approval.ResolvedBy.String())
	}
	if got := out.Result["reason"]; got != approval.ResolutionReason {
		t.Errorf("reason = %v, want %q", got, approval.ResolutionReason)
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

func TestRetryTemporalCompletion_QueryByIDGenericError(t *testing.T) {
	// Generic DB error → 500 Internal.
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

// TestCompleteAndClear_CallsClearTokenAfterSuccess verifies that when
// asyncCompleter.Complete succeeds, ClearTaskToken is called on the storer.
func TestCompleteAndClear_CallsClearTokenAfterSuccess(t *testing.T) {
	approval := resolvedApproval(validToken())

	storer := &mockStorer{}
	completer := &mockActivityCompleter{err: nil}
	a := newTestAPI(storer, completer)

	req := ResolveRequest{Resolution: "approved", Reason: "looks good"}
	a.completeAndClear(context.Background(), approval.ID, approval, req, uuid.New())

	if !completer.called {
		t.Fatal("CompleteActivity should have been called")
	}
	if !storer.clearCalled {
		t.Fatal("ClearTaskToken should have been called after successful Complete")
	}
}

// TestCompleteAndClear_NoClearOnCompleteFailure verifies that when
// asyncCompleter.Complete fails, ClearTaskToken is NOT called.
func TestCompleteAndClear_NoClearOnCompleteFailure(t *testing.T) {
	approval := resolvedApproval(validToken())

	storer := &mockStorer{}
	completer := &mockActivityCompleter{err: errors.New("temporal timeout")}
	a := newTestAPI(storer, completer)

	req := ResolveRequest{Resolution: "approved", Reason: "looks good"}
	a.completeAndClear(context.Background(), approval.ID, approval, req, uuid.New())

	if !completer.called {
		t.Fatal("CompleteActivity should have been attempted")
	}
	if storer.clearCalled {
		t.Fatal("ClearTaskToken must not be called when Complete() fails")
	}
}

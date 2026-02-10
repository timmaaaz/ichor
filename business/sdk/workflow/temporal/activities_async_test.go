package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"go.temporal.io/sdk/testsuite"
)

// =============================================================================
// Mock Async Handler
// =============================================================================

type mockAsyncHandler struct {
	actionType     string
	startAsyncErr  error
	capturedToken  []byte
	capturedConfig json.RawMessage
	called         int
}

func (h *mockAsyncHandler) Execute(_ context.Context, _ json.RawMessage, _ workflow.ActionExecutionContext) (any, error) {
	return nil, errors.New("should not be called for async")
}

func (h *mockAsyncHandler) StartAsync(_ context.Context, config json.RawMessage, _ workflow.ActionExecutionContext, taskToken []byte) error {
	h.called++
	h.capturedToken = taskToken
	h.capturedConfig = config
	return h.startAsyncErr
}

func (h *mockAsyncHandler) Validate(_ json.RawMessage) error { return nil }
func (h *mockAsyncHandler) GetType() string                  { return h.actionType }
func (h *mockAsyncHandler) SupportsManualExecution() bool    { return false }
func (h *mockAsyncHandler) IsAsync() bool                    { return true }
func (h *mockAsyncHandler) GetDescription() string           { return "mock async" }

// =============================================================================
// Async Activity Tests via SDK TestActivityEnvironment
// =============================================================================

func TestAsyncActivity_CallsStartAsync(t *testing.T) {
	handler := &mockAsyncHandler{actionType: "send_email"}
	asyncReg := NewAsyncRegistry()
	asyncReg.Register("send_email", handler)

	activities := &Activities{
		Registry:      workflow.NewActionRegistry(),
		AsyncRegistry: asyncReg,
	}

	input := ActionActivityInput{
		ActionID:    uuid.New(),
		ActionName:  "send_email_1",
		ActionType:  "send_email",
		Config:      json.RawMessage(`{"to": "user@example.com"}`),
		Context:     map[string]any{"entity_id": "123"},
		RuleID:      uuid.New(),
		ExecutionID: uuid.New(),
		RuleName:    "test-rule",
	}

	// Use SDK TestActivityEnvironment which provides proper activity context.
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities)

	// ExecuteAsyncActionActivity returns activity.ErrResultPending on success.
	// The SDK TestActivityEnvironment surfaces this as an error with the
	// message: "not error: do not autocomplete, using Client.CompleteActivity() to complete"
	_, err := env.ExecuteActivity(activities.ExecuteAsyncActionActivity, input)

	// ErrResultPending is expected behavior, not a failure.
	// SDK test env wraps it differently than a real error.
	require.Error(t, err, "should return ErrResultPending wrapped as error")
	require.Contains(t, err.Error(), "CompleteActivity",
		"error should be the ErrResultPending message, not a real failure")

	require.Equal(t, 1, handler.called, "StartAsync should be called")
	// JSON may be compacted during Temporal SDK round-trip (spaces removed).
	require.JSONEq(t, `{"to": "user@example.com"}`, string(handler.capturedConfig))
}

func TestAsyncActivity_StartAsyncError_ReturnsError(t *testing.T) {
	handler := &mockAsyncHandler{
		actionType:    "send_email",
		startAsyncErr: errors.New("queue unavailable"),
	}
	asyncReg := NewAsyncRegistry()
	asyncReg.Register("send_email", handler)

	activities := &Activities{
		Registry:      workflow.NewActionRegistry(),
		AsyncRegistry: asyncReg,
	}

	input := ActionActivityInput{
		ActionID:    uuid.New(),
		ActionName:  "send_email_1",
		ActionType:  "send_email",
		Config:      json.RawMessage(`{}`),
		Context:     map[string]any{},
		RuleID:      uuid.New(),
		ExecutionID: uuid.New(),
		RuleName:    "test-rule",
	}

	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities)

	_, err := env.ExecuteActivity(activities.ExecuteAsyncActionActivity, input)
	require.Error(t, err)
	// Should be a real error, NOT ErrResultPending.
	require.Contains(t, err.Error(), "queue unavailable")
	require.Contains(t, err.Error(), "send_email_1")
}

func TestAsyncActivity_UnknownAsyncType(t *testing.T) {
	asyncReg := NewAsyncRegistry()

	activities := &Activities{
		Registry:      workflow.NewActionRegistry(),
		AsyncRegistry: asyncReg,
	}

	input := ActionActivityInput{
		ActionID:    uuid.New(),
		ActionName:  "unknown_1",
		ActionType:  "nonexistent_type",
		Config:      json.RawMessage(`{}`),
		Context:     map[string]any{},
		RuleID:      uuid.New(),
		ExecutionID: uuid.New(),
		RuleName:    "test-rule",
	}

	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(activities)

	_, err := env.ExecuteActivity(activities.ExecuteAsyncActionActivity, input)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no handler registered")
}

// =============================================================================
// Mock ActivityCompleter
// =============================================================================

type mockActivityCompleter struct {
	completedToken  []byte
	completedResult any
	failedToken     []byte
	failedErr       error
	returnErr       error
}

func (m *mockActivityCompleter) CompleteActivity(_ context.Context, taskToken []byte, result any, err error) error {
	if err != nil {
		m.failedToken = taskToken
		m.failedErr = err
	} else {
		m.completedToken = taskToken
		m.completedResult = result
	}
	return m.returnErr
}

// =============================================================================
// AsyncCompleter Tests
// =============================================================================

func TestAsyncCompleter_Complete_Success(t *testing.T) {
	mock := &mockActivityCompleter{}
	completer := NewAsyncCompleter(mock)

	token := []byte("test-token-123")
	result := ActionActivityOutput{
		Success: true,
		Result:  map[string]any{"status": "processed"},
	}

	err := completer.Complete(context.Background(), token, result)
	require.NoError(t, err)
	require.Equal(t, token, mock.completedToken)
	require.NotNil(t, mock.completedResult)
}

func TestAsyncCompleter_Complete_EmptyResult(t *testing.T) {
	mock := &mockActivityCompleter{}
	completer := NewAsyncCompleter(mock)

	err := completer.Complete(context.Background(), []byte("token"), ActionActivityOutput{})
	require.NoError(t, err)
	require.NotNil(t, mock.completedResult)
}

func TestAsyncCompleter_Fail_Success(t *testing.T) {
	mock := &mockActivityCompleter{}
	completer := NewAsyncCompleter(mock)

	token := []byte("test-token-456")
	activityErr := errors.New("processing failed")

	err := completer.Fail(context.Background(), token, activityErr)
	require.NoError(t, err)
	require.Equal(t, token, mock.failedToken)
	require.Equal(t, activityErr, mock.failedErr)
}

func TestAsyncCompleter_Complete_ClientError(t *testing.T) {
	mock := &mockActivityCompleter{returnErr: errors.New("temporal unavailable")}
	completer := NewAsyncCompleter(mock)

	err := completer.Complete(context.Background(), []byte("token"), ActionActivityOutput{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "temporal unavailable")
}

func TestAsyncCompleter_Fail_ClientError(t *testing.T) {
	mock := &mockActivityCompleter{returnErr: errors.New("temporal unavailable")}
	completer := NewAsyncCompleter(mock)

	err := completer.Fail(context.Background(), []byte("token"), errors.New("activity failed"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "temporal unavailable")
}

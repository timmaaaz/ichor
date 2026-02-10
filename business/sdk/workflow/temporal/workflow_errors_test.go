package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// retryTestHandler: fails for the first failCount calls, then succeeds.
// =============================================================================

type retryTestHandler struct {
	actionType string
	failCount  int
	result     any
	callCount  int
}

func (h *retryTestHandler) Execute(_ context.Context, _ json.RawMessage, _ workflow.ActionExecutionContext) (any, error) {
	h.callCount++
	if h.callCount <= h.failCount {
		return nil, fmt.Errorf("attempt %d/%d failed", h.callCount, h.failCount)
	}
	return h.result, nil
}

func (h *retryTestHandler) Validate(_ json.RawMessage) error { return nil }
func (h *retryTestHandler) GetType() string                  { return h.actionType }
func (h *retryTestHandler) SupportsManualExecution() bool    { return false }
func (h *retryTestHandler) IsAsync() bool                    { return false }
func (h *retryTestHandler) GetDescription() string           { return "retry test" }

// =============================================================================
// Activity Failure → Workflow Failure
// =============================================================================

func TestError_ActivityFailure_FailsWorkflow(t *testing.T) {
	handler := &testActionHandler{
		actionType: "fail_action",
		err:        errors.New("action execution failed"),
	}
	env := setupTestEnv(t, newTestRegistry(handler))

	graph, _ := wfLinearGraph("fail_action")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "error-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowError()
	require.Error(t, err)
	// Error wrapping uses action name (action_0) from wfLinearGraph.
	require.Contains(t, err.Error(), "action_0")
	require.Contains(t, err.Error(), "action execution failed")
	// Regular action retries 3 times before failing (MaximumAttempts=3).
	require.Equal(t, 3, handler.called)
}

func TestError_ActivityFailure_MidChain_StopsExecution(t *testing.T) {
	// Graph: A (success) → B (fail) → C (success)
	// Verify: A runs, B fails (after retries), C does NOT run.
	handlerA := &testActionHandler{actionType: "action_a", result: map[string]any{"ok": true}}
	handlerB := &testActionHandler{actionType: "action_b", err: errors.New("mid-chain failure")}
	handlerC := &testActionHandler{actionType: "action_c", result: map[string]any{"ok": true}}

	reg := newTestRegistry(handlerA, handlerB, handlerC)
	env := setupTestEnv(t, reg)

	graph, _ := wfLinearGraph("action_a", "action_b", "action_c")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "mid-chain-error",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.Error(t, env.GetWorkflowError())
	require.Equal(t, 1, handlerA.called, "A should run once (success)")
	require.Equal(t, 3, handlerB.called, "B retries 3 times before failing")
	require.Equal(t, 0, handlerC.called, "C should NOT run after B fails")
}

func TestError_ActivityFailure_ErrorContainsActionName(t *testing.T) {
	handler := &testActionHandler{
		actionType: "specific_action",
		err:        errors.New("detailed error message"),
	}
	env := setupTestEnv(t, newTestRegistry(handler))

	graph, _ := wfLinearGraph("specific_action")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "error-detail-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	err := env.GetWorkflowError()
	require.Error(t, err)
	// Error wrapping: "execute action %s (%s): ..." with action name from graph.
	// wfLinearGraph names actions "action_0", "action_1", etc.
	require.Contains(t, err.Error(), "action_0")
	require.Contains(t, err.Error(), "detailed error message")
}

// =============================================================================
// Branch Failure in Parallel Execution
// =============================================================================

func TestError_ParallelBranchFailure_FailsWorkflow(t *testing.T) {
	// Parallel: fork → branchA (success) and branchB (fail) → convergence
	// wfParallelGraph uses fork_action and merge_action internally.
	forkHandler := &testActionHandler{actionType: "fork_action", result: map[string]any{}}
	handlerA := &testActionHandler{actionType: "branch_a", result: map[string]any{"ok": true}}
	handlerB := &testActionHandler{actionType: "branch_b", err: errors.New("branch B failed")}
	mergeHandler := &testActionHandler{actionType: "merge_action", result: map[string]any{"merged": true}}

	reg := newTestRegistry(forkHandler, handlerA, handlerB, mergeHandler)
	env := setupTestEnv(t, reg)

	graph, _ := wfParallelGraph("branch_a", "branch_b")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "parallel-error-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	err := env.GetWorkflowError()
	require.Error(t, err)
	require.Contains(t, err.Error(), "branch B failed")
	require.Equal(t, 0, mergeHandler.called, "merge should NOT run when a branch fails")
}

func TestError_ParallelBothBranchesFail(t *testing.T) {
	forkHandler := &testActionHandler{actionType: "fork_action", result: map[string]any{}}
	handlerA := &testActionHandler{actionType: "branch_a", err: errors.New("branch A failed")}
	handlerB := &testActionHandler{actionType: "branch_b", err: errors.New("branch B failed")}
	mergeHandler := &testActionHandler{actionType: "merge_action", result: map[string]any{}}

	reg := newTestRegistry(forkHandler, handlerA, handlerB, mergeHandler)
	env := setupTestEnv(t, reg)

	graph, _ := wfParallelGraph("branch_a", "branch_b")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "both-branches-fail",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.Error(t, env.GetWorkflowError())
	require.Equal(t, 0, mergeHandler.called)
}

// =============================================================================
// Fire-and-Forget Error Isolation
// =============================================================================

func TestError_FireAndForget_ErrorIsolation(t *testing.T) {
	// Fire-and-forget: fork → branchA (success) and branchB (fail)
	// Parent workflow should SUCCEED — fire-and-forget branches run independently.
	// executeFireAndForget returns nil immediately.
	forkHandler := &testActionHandler{actionType: "fork_action", result: map[string]any{}}
	handlerA := &testActionHandler{actionType: "branch_a", result: map[string]any{"ok": true}}
	handlerB := &testActionHandler{actionType: "branch_b", err: errors.New("fire-forget error")}

	reg := newTestRegistry(forkHandler, handlerA, handlerB)
	env := setupTestEnv(t, reg)

	graph, _ := wfFireForgetGraph("branch_a", "branch_b")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "fire-forget-error",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	err := env.GetWorkflowError()
	require.NoError(t, err, "fire-and-forget branch errors should NOT fail parent")
}

// =============================================================================
// Handler Not Found
// =============================================================================

func TestError_HandlerNotFound(t *testing.T) {
	env := setupTestEnv(t, newTestRegistry())

	graph, _ := wfLinearGraph("nonexistent_type")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "missing-handler",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	err := env.GetWorkflowError()
	require.Error(t, err)
	require.Contains(t, err.Error(), "no handler registered")
}

// =============================================================================
// Retry Policy Tests (Workflow-Level)
// =============================================================================

func TestError_RetryPolicy_RegularAction_SucceedsAfterRetries(t *testing.T) {
	// Regular action: MaximumAttempts=3.
	// Handler fails on first 2 calls, succeeds on 3rd.
	handler := &retryTestHandler{
		actionType: "retry_action",
		failCount:  2,
		result:     map[string]any{"status": "ok"},
	}

	reg := workflow.NewActionRegistry()
	reg.Register(handler)
	env := setupTestEnv(t, reg)

	graph, _ := wfLinearGraph("retry_action")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "retry-success",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, 3, handler.callCount, "should be called 3 times (2 failures + 1 success)")
}

func TestError_RetryPolicy_RegularAction_ExhaustsRetries(t *testing.T) {
	// Regular action: MaximumAttempts=3.
	// Handler always fails. After 3 attempts, workflow should fail.
	handler := &retryTestHandler{
		actionType: "always_fail",
		failCount:  100,
	}

	reg := workflow.NewActionRegistry()
	reg.Register(handler)
	env := setupTestEnv(t, reg)

	graph, _ := wfLinearGraph("always_fail")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "retry-exhausted",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.Error(t, env.GetWorkflowError())
	require.Equal(t, 3, handler.callCount, "should attempt exactly 3 times")
}

// =============================================================================
// Retry Policy Tests (Unit-Level via activityOptions)
// =============================================================================

func TestError_RetryPolicy_RegularAction_Config(t *testing.T) {
	opts := activityOptions("update_field")
	require.Equal(t, int32(3), opts.RetryPolicy.MaximumAttempts)
	require.Equal(t, 5*time.Minute, opts.StartToCloseTimeout)
	require.Equal(t, time.Second, opts.RetryPolicy.InitialInterval)
	require.Equal(t, 2.0, opts.RetryPolicy.BackoffCoefficient)
}

func TestError_RetryPolicy_AsyncAction_Config(t *testing.T) {
	// Long-running actions: 3 retries allowed (Temporal handles safely), longer timeouts.
	// Previously MaximumAttempts=1 to avoid RabbitMQ duplicates, but now Temporal-only.
	opts := activityOptions("send_email")
	require.Equal(t, int32(3), opts.RetryPolicy.MaximumAttempts,
		"long-running action should retry (Temporal handles safely)")
	require.Equal(t, 30*time.Minute, opts.StartToCloseTimeout)
	require.Equal(t, time.Minute, opts.HeartbeatTimeout)
}

func TestError_RetryPolicy_HumanAction_Config(t *testing.T) {
	// Human actions: MaximumAttempts=1, multi-day timeout.
	opts := activityOptions("manager_approval")
	require.Equal(t, int32(1), opts.RetryPolicy.MaximumAttempts,
		"human action should NOT retry (MaximumAttempts=1)")
	require.Equal(t, 7*24*time.Hour, opts.StartToCloseTimeout)
	require.Equal(t, time.Hour, opts.HeartbeatTimeout)
}

func TestError_RetryPolicy_AllLongRunningTypes(t *testing.T) {
	// Verify all long-running action types get retries (Temporal handles safely).
	// Previously MaximumAttempts=1 to avoid RabbitMQ duplicates, but now Temporal-only.
	longRunningTypes := []string{
		"allocate_inventory", "send_email", "credit_check",
		"fraud_detection", "third_party_api_call", "reserve_shipping",
	}
	for _, at := range longRunningTypes {
		opts := activityOptions(at)
		require.Equal(t, int32(3), opts.RetryPolicy.MaximumAttempts,
			"long-running type %s should have MaximumAttempts=3", at)
	}
}

func TestError_RetryPolicy_AllHumanTypes(t *testing.T) {
	// Verify all known human action types get MaximumAttempts=1.
	humanTypes := []string{
		"manager_approval", "manual_review",
		"human_verification", "approval_request",
	}
	for _, ht := range humanTypes {
		opts := activityOptions(ht)
		require.Equal(t, int32(1), opts.RetryPolicy.MaximumAttempts,
			"human type %s should have MaximumAttempts=1", ht)
	}
}

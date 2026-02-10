package temporal

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Parallel with Convergence
// =============================================================================

func TestWorkflow_ParallelBranches_Convergence(t *testing.T) {
	forkHandler := &testActionHandler{actionType: "fork_action", result: map[string]any{"forked": true}}
	branchAHandler := &testActionHandler{actionType: "branch_a_type", result: map[string]any{"price": 100}}
	branchBHandler := &testActionHandler{actionType: "branch_b_type", result: map[string]any{"quantity": 5}}
	mergeHandler := &testActionHandler{actionType: "merge_action", result: map[string]any{"merged": true}}

	env := setupTestEnv(t, newTestRegistry(forkHandler, branchAHandler, branchBHandler, mergeHandler))

	graph, _ := wfParallelGraph("branch_a_type", "branch_b_type")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "parallel-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	require.Equal(t, 1, forkHandler.called)
	require.Equal(t, 1, branchAHandler.called)
	require.Equal(t, 1, branchBHandler.called)
	require.Equal(t, 1, mergeHandler.called, "merge action at convergence should be called")
}

// =============================================================================
// Fire-and-Forget
// =============================================================================

func TestWorkflow_FireAndForget(t *testing.T) {
	forkHandler := &testActionHandler{actionType: "fork_action", result: map[string]any{"forked": true}}
	branchAHandler := &testActionHandler{actionType: "branch_a_type", result: map[string]any{"a": 1}}
	branchBHandler := &testActionHandler{actionType: "branch_b_type", result: map[string]any{"b": 2}}

	env := setupTestEnv(t, newTestRegistry(forkHandler, branchAHandler, branchBHandler))

	graph, _ := wfFireForgetGraph("branch_a_type", "branch_b_type")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "fire-forget-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Fork handler called. Branch handlers launched as child workflows
	// with PARENT_CLOSE_POLICY_ABANDON.
	// In SDK test suite, child workflows execute synchronously.
	require.Equal(t, 1, forkHandler.called)
	require.Equal(t, 1, branchAHandler.called)
	require.Equal(t, 1, branchBHandler.called)
}

// =============================================================================
// Direct Child Workflow Tests (ExecuteBranchUntilConvergence)
// =============================================================================

func TestWorkflow_BranchUntilConvergence_SingleBranch(t *testing.T) {
	handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"v": 0}}
	handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"v": 1}}
	handler2 := &testActionHandler{actionType: "step_2", result: map[string]any{"v": 2}}

	env := setupTestEnv(t, newTestRegistry(handler0, handler1, handler2))

	// Build a linear graph: step_0 -> step_1 -> step_2
	graph, ids := wfLinearGraph("step_0", "step_1", "step_2")

	// Branch starts at step_0, converges at step_2 (step_2 should NOT execute).
	branchInput := BranchInput{
		StartAction:      graph.Actions[0],
		ConvergencePoint: ids[2],
		Graph:            graph,
		InitialContext:   NewMergedContext(map[string]any{"trigger": "test"}),
		RuleID:           uuid.New(),
		ExecutionID:      uuid.New(),
		RuleName:         "branch-test",
	}

	env.ExecuteWorkflow(ExecuteBranchUntilConvergence, branchInput)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	require.Equal(t, 1, handler0.called)
	require.Equal(t, 1, handler1.called)
	require.Equal(t, 0, handler2.called, "convergence point should NOT be executed by branch")
}

func TestWorkflow_BranchUntilConvergence_FireAndForget(t *testing.T) {
	handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"v": 0}}
	handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"v": 1}}

	env := setupTestEnv(t, newTestRegistry(handler0, handler1))

	graph, _ := wfLinearGraph("step_0", "step_1")
	branchInput := BranchInput{
		StartAction:      graph.Actions[0],
		ConvergencePoint: uuid.Nil, // Fire-and-forget: run until end
		Graph:            graph,
		InitialContext:   NewMergedContext(map[string]any{}),
		RuleID:           uuid.New(),
		ExecutionID:      uuid.New(),
		RuleName:         "fire-forget-branch",
	}

	env.ExecuteWorkflow(ExecuteBranchUntilConvergence, branchInput)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Both actions executed (no convergence to stop at).
	require.Equal(t, 1, handler0.called)
	require.Equal(t, 1, handler1.called)
}

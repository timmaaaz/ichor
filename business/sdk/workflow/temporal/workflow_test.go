package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"go.temporal.io/sdk/testsuite"
)

// =============================================================================
// Test Action Handler (implements all 6 ActionHandler methods)
// =============================================================================

// testActionHandler is a configurable mock action handler for testing.
// Tracks call count so tests can verify which handlers were invoked.
type testActionHandler struct {
	actionType string
	result     any
	err        error
	called     int
}

func (h *testActionHandler) Execute(_ context.Context, _ json.RawMessage, _ workflow.ActionExecutionContext) (any, error) {
	h.called++
	if h.err != nil {
		return nil, h.err
	}
	return h.result, nil
}

func (h *testActionHandler) Validate(_ json.RawMessage) error   { return nil }
func (h *testActionHandler) GetType() string                    { return h.actionType }
func (h *testActionHandler) SupportsManualExecution() bool      { return false }
func (h *testActionHandler) IsAsync() bool                      { return false }
func (h *testActionHandler) GetDescription() string             { return "test handler" }

// newTestRegistry creates an ActionRegistry pre-populated with test handlers.
func newTestRegistry(handlers ...*testActionHandler) *workflow.ActionRegistry {
	reg := workflow.NewActionRegistry()
	for _, h := range handlers {
		reg.Register(h)
	}
	return reg
}

// =============================================================================
// SDK Test Suite Setup
// =============================================================================

// setupTestEnv creates a Temporal test environment with registered workflows/activities.
func setupTestEnv(t *testing.T, registry *workflow.ActionRegistry) *testsuite.TestWorkflowEnvironment {
	t.Helper()
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(ExecuteGraphWorkflow)
	env.RegisterWorkflow(ExecuteBranchUntilConvergence)

	activities := &Activities{
		Registry:      registry,
		AsyncRegistry: NewAsyncRegistry(),
	}
	env.RegisterActivity(activities)

	return env
}

// =============================================================================
// Workflow Graph Builder Helpers
// =============================================================================

// wfLinearGraph creates: start -> action[0] -> action[1] -> ... -> action[N-1]
// Each action gets a unique UUID. Returns graph and ordered action IDs.
func wfLinearGraph(actionTypes ...string) (GraphDefinition, []uuid.UUID) {
	if len(actionTypes) == 0 {
		return GraphDefinition{}, nil
	}

	actions := make([]ActionNode, len(actionTypes))
	ids := make([]uuid.UUID, len(actionTypes))

	for i, at := range actionTypes {
		ids[i] = uuid.New()
		actions[i] = ActionNode{
			ID:         ids[i],
			Name:       fmt.Sprintf("action_%d", i),
			ActionType: at,
			Config:     json.RawMessage(`{}`),
			IsActive:   true,
		}
	}

	edges := make([]ActionEdge, 0, len(actionTypes))
	edges = append(edges, ActionEdge{
		ID:             uuid.New(),
		SourceActionID: nil,
		TargetActionID: ids[0],
		EdgeType:       EdgeTypeStart,
		SortOrder:      1,
	})

	for i := 1; i < len(actionTypes); i++ {
		src := ids[i-1]
		edges = append(edges, ActionEdge{
			ID:             uuid.New(),
			SourceActionID: &src,
			TargetActionID: ids[i],
			EdgeType:       EdgeTypeSequence,
			SortOrder:      1,
		})
	}

	return GraphDefinition{Actions: actions, Edges: edges}, ids
}

// wfDiamondIDs holds action IDs for a diamond-shaped condition graph.
type wfDiamondIDs struct {
	Condition uuid.UUID
	TrueArm   uuid.UUID
	FalseArm  uuid.UUID
	Merge     uuid.UUID
}

// wfConditionDiamond creates:
//
//	start -> condition --(true_branch)--> trueArm --(sequence)--> merge
//	                   \--(false_branch)-> falseArm --(sequence)--> merge
func wfConditionDiamond(trueArmType, falseArmType string) (GraphDefinition, wfDiamondIDs) {
	condID := uuid.New()
	trueID := uuid.New()
	falseID := uuid.New()
	mergeID := uuid.New()

	actions := []ActionNode{
		{ID: condID, Name: "condition", ActionType: "evaluate_condition", Config: json.RawMessage(`{}`), IsActive: true},
		{ID: trueID, Name: "true_arm", ActionType: trueArmType, Config: json.RawMessage(`{}`), IsActive: true},
		{ID: falseID, Name: "false_arm", ActionType: falseArmType, Config: json.RawMessage(`{}`), IsActive: true},
		{ID: mergeID, Name: "merge", ActionType: "merge_action", Config: json.RawMessage(`{}`), IsActive: true},
	}

	edges := []ActionEdge{
		{ID: uuid.New(), SourceActionID: nil, TargetActionID: condID, EdgeType: EdgeTypeStart, SortOrder: 1},
		{ID: uuid.New(), SourceActionID: &condID, TargetActionID: trueID, EdgeType: EdgeTypeTrueBranch, SortOrder: 1},
		{ID: uuid.New(), SourceActionID: &condID, TargetActionID: falseID, EdgeType: EdgeTypeFalseBranch, SortOrder: 2},
		{ID: uuid.New(), SourceActionID: &trueID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
		{ID: uuid.New(), SourceActionID: &falseID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
	}

	return GraphDefinition{Actions: actions, Edges: edges}, wfDiamondIDs{
		Condition: condID,
		TrueArm:   trueID,
		FalseArm:  falseID,
		Merge:     mergeID,
	}
}

// wfParallelIDs holds action IDs for a parallel convergence graph.
type wfParallelIDs struct {
	Fork    uuid.UUID
	BranchA uuid.UUID
	BranchB uuid.UUID
	Merge   uuid.UUID
}

// wfParallelGraph creates:
//
//	start -> fork --(sequence)--> branchA --(sequence)--> merge
//	              \--(sequence)--> branchB --(sequence)--> merge
func wfParallelGraph(branchAType, branchBType string) (GraphDefinition, wfParallelIDs) {
	forkID := uuid.New()
	branchAID := uuid.New()
	branchBID := uuid.New()
	mergeID := uuid.New()

	actions := []ActionNode{
		{ID: forkID, Name: "fork", ActionType: "fork_action", Config: json.RawMessage(`{}`), IsActive: true},
		{ID: branchAID, Name: "branch_a", ActionType: branchAType, Config: json.RawMessage(`{}`), IsActive: true},
		{ID: branchBID, Name: "branch_b", ActionType: branchBType, Config: json.RawMessage(`{}`), IsActive: true},
		{ID: mergeID, Name: "merge", ActionType: "merge_action", Config: json.RawMessage(`{}`), IsActive: true},
	}

	edges := []ActionEdge{
		{ID: uuid.New(), SourceActionID: nil, TargetActionID: forkID, EdgeType: EdgeTypeStart, SortOrder: 1},
		{ID: uuid.New(), SourceActionID: &forkID, TargetActionID: branchAID, EdgeType: EdgeTypeSequence, SortOrder: 1},
		{ID: uuid.New(), SourceActionID: &forkID, TargetActionID: branchBID, EdgeType: EdgeTypeSequence, SortOrder: 2},
		{ID: uuid.New(), SourceActionID: &branchAID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
		{ID: uuid.New(), SourceActionID: &branchBID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
	}

	return GraphDefinition{Actions: actions, Edges: edges}, wfParallelIDs{
		Fork:    forkID,
		BranchA: branchAID,
		BranchB: branchBID,
		Merge:   mergeID,
	}
}

// wfFireForgetIDs holds action IDs for a fire-and-forget graph.
type wfFireForgetIDs struct {
	Fork    uuid.UUID
	BranchA uuid.UUID
	BranchB uuid.UUID
}

// wfFireForgetGraph creates parallel branches with NO convergence:
//
//	start -> fork --(sequence)--> branchA
//	              \--(sequence)--> branchB
func wfFireForgetGraph(branchAType, branchBType string) (GraphDefinition, wfFireForgetIDs) {
	forkID := uuid.New()
	branchAID := uuid.New()
	branchBID := uuid.New()

	actions := []ActionNode{
		{ID: forkID, Name: "fork", ActionType: "fork_action", Config: json.RawMessage(`{}`), IsActive: true},
		{ID: branchAID, Name: "branch_a", ActionType: branchAType, Config: json.RawMessage(`{}`), IsActive: true},
		{ID: branchBID, Name: "branch_b", ActionType: branchBType, Config: json.RawMessage(`{}`), IsActive: true},
	}

	edges := []ActionEdge{
		{ID: uuid.New(), SourceActionID: nil, TargetActionID: forkID, EdgeType: EdgeTypeStart, SortOrder: 1},
		{ID: uuid.New(), SourceActionID: &forkID, TargetActionID: branchAID, EdgeType: EdgeTypeSequence, SortOrder: 1},
		{ID: uuid.New(), SourceActionID: &forkID, TargetActionID: branchBID, EdgeType: EdgeTypeSequence, SortOrder: 2},
	}

	return GraphDefinition{Actions: actions, Edges: edges}, wfFireForgetIDs{
		Fork:    forkID,
		BranchA: branchAID,
		BranchB: branchBID,
	}
}

// =============================================================================
// Single Action Tests
// =============================================================================

func TestWorkflow_SingleAction(t *testing.T) {
	handler := &testActionHandler{
		actionType: "test_action",
		result:     map[string]any{"status": "done"},
	}
	env := setupTestEnv(t, newTestRegistry(handler))

	graph, _ := wfLinearGraph("test_action")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "test-rule",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{"entity_id": "123"},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, 1, handler.called)
}

// =============================================================================
// Sequential Chain Tests
// =============================================================================

func TestWorkflow_SequentialChain(t *testing.T) {
	handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"val": "a"}}
	handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"val": "b"}}
	handler2 := &testActionHandler{actionType: "step_2", result: map[string]any{"val": "c"}}

	env := setupTestEnv(t, newTestRegistry(handler0, handler1, handler2))

	graph, _ := wfLinearGraph("step_0", "step_1", "step_2")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "chain-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{"entity_id": "456"},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	require.Equal(t, 1, handler0.called)
	require.Equal(t, 1, handler1.called)
	require.Equal(t, 1, handler2.called)
}

// =============================================================================
// Condition Branching Tests
// =============================================================================

func TestWorkflow_ConditionBranch_True(t *testing.T) {
	condHandler := &testActionHandler{
		actionType: "evaluate_condition",
		result:     map[string]any{"branch_taken": "true_branch"},
	}
	trueHandler := &testActionHandler{actionType: "true_action", result: map[string]any{"path": "true"}}
	falseHandler := &testActionHandler{actionType: "false_action", result: map[string]any{"path": "false"}}
	mergeHandler := &testActionHandler{actionType: "merge_action", result: map[string]any{"merged": true}}

	env := setupTestEnv(t, newTestRegistry(condHandler, trueHandler, falseHandler, mergeHandler))

	graph, _ := wfConditionDiamond("true_action", "false_action")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "cond-true-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	require.Equal(t, 1, condHandler.called)
	require.Equal(t, 1, trueHandler.called)
	require.Equal(t, 0, falseHandler.called, "false_branch handler should NOT be called")
	require.Equal(t, 1, mergeHandler.called)
}

func TestWorkflow_ConditionBranch_False(t *testing.T) {
	condHandler := &testActionHandler{
		actionType: "evaluate_condition",
		result:     map[string]any{"branch_taken": "false_branch"},
	}
	trueHandler := &testActionHandler{actionType: "true_action", result: map[string]any{"path": "true"}}
	falseHandler := &testActionHandler{actionType: "false_action", result: map[string]any{"path": "false"}}
	mergeHandler := &testActionHandler{actionType: "merge_action", result: map[string]any{"merged": true}}

	env := setupTestEnv(t, newTestRegistry(condHandler, trueHandler, falseHandler, mergeHandler))

	graph, _ := wfConditionDiamond("true_action", "false_action")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "cond-false-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	require.Equal(t, 1, condHandler.called)
	require.Equal(t, 0, trueHandler.called, "true_branch handler should NOT be called")
	require.Equal(t, 1, falseHandler.called)
	require.Equal(t, 1, mergeHandler.called)
}

// =============================================================================
// Validation Error Tests
// =============================================================================

func TestWorkflow_InvalidInput_NoRuleID(t *testing.T) {
	env := setupTestEnv(t, newTestRegistry())

	input := WorkflowInput{
		ExecutionID: uuid.New(),
		Graph: GraphDefinition{
			Actions: []ActionNode{{ID: uuid.New(), Name: "a", ActionType: "t", Config: json.RawMessage(`{}`), IsActive: true}},
			Edges:   []ActionEdge{{ID: uuid.New(), EdgeType: EdgeTypeStart, TargetActionID: uuid.New()}},
		},
	}
	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	require.Contains(t, env.GetWorkflowError().Error(), "rule_id is required")
}

func TestWorkflow_InvalidInput_NoStartEdges(t *testing.T) {
	env := setupTestEnv(t, newTestRegistry())

	actionID := uuid.New()
	targetID := uuid.New()
	input := WorkflowInput{
		RuleID:      uuid.New(),
		ExecutionID: uuid.New(),
		Graph: GraphDefinition{
			Actions: []ActionNode{{ID: actionID, Name: "orphan", ActionType: "test", Config: json.RawMessage(`{}`), IsActive: true}},
			Edges:   []ActionEdge{{ID: uuid.New(), SourceActionID: &actionID, TargetActionID: targetID, EdgeType: EdgeTypeSequence}},
		},
	}
	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	require.Contains(t, env.GetWorkflowError().Error(), "start edge")
}

func TestWorkflow_InvalidInput_EmptyGraph(t *testing.T) {
	env := setupTestEnv(t, newTestRegistry())

	input := WorkflowInput{
		RuleID:      uuid.New(),
		ExecutionID: uuid.New(),
		Graph:       GraphDefinition{},
	}
	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	require.Contains(t, env.GetWorkflowError().Error(), "at least one action")
}

// =============================================================================
// Error Propagation Tests
// =============================================================================

func TestWorkflow_ActionError_FailsWorkflow(t *testing.T) {
	failHandler := &testActionHandler{
		actionType: "fail_action",
		err:        fmt.Errorf("simulated failure"),
	}
	env := setupTestEnv(t, newTestRegistry(failHandler))

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
	require.Error(t, env.GetWorkflowError())
	require.Contains(t, env.GetWorkflowError().Error(), "action_0")
}

func TestWorkflow_UnknownActionType(t *testing.T) {
	env := setupTestEnv(t, newTestRegistry())

	graph, _ := wfLinearGraph("unknown_type")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "unknown-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	require.Contains(t, env.GetWorkflowError().Error(), "no handler registered")
}

// =============================================================================
// Deactivated Action Tests
// =============================================================================

// TestWorkflow_DeactivatedAction_StillExecutes documents that the current
// implementation does NOT skip deactivated actions. Update when skip logic is added.
func TestWorkflow_DeactivatedAction_StillExecutes(t *testing.T) {
	handler := &testActionHandler{
		actionType: "test_action",
		result:     map[string]any{"status": "done"},
	}
	env := setupTestEnv(t, newTestRegistry(handler))

	actionID := uuid.New()
	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: actionID, Name: "deactivated", ActionType: "test_action", Config: json.RawMessage(`{}`), IsActive: false, DeactivatedBy: uuid.New()},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: actionID, EdgeType: EdgeTypeStart, SortOrder: 1},
		},
	}
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "deactivated-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, 1, handler.called)
}

// =============================================================================
// Delay Action Tests
// =============================================================================

func TestWorkflow_DelayAction(t *testing.T) {
	// Graph: start -> delay(5s) -> test_action
	// The Temporal test env auto-skips timers, so the test completes instantly.
	handler := &testActionHandler{
		actionType: "test_action",
		result:     map[string]any{"status": "after_delay"},
	}
	env := setupTestEnv(t, newTestRegistry(handler))

	delayID := uuid.New()
	actionID := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: delayID, Name: "wait", ActionType: "delay", Config: json.RawMessage(`{"duration": "5s"}`), IsActive: true},
			{ID: actionID, Name: "after_wait", ActionType: "test_action", Config: json.RawMessage(`{}`), IsActive: true},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: delayID, EdgeType: EdgeTypeStart, SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &delayID, TargetActionID: actionID, EdgeType: EdgeTypeSequence, SortOrder: 1},
		},
	}

	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "delay-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{"entity_id": "123"},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	// The test_action after the delay should have been called
	require.Equal(t, 1, handler.called)
}

func TestWorkflow_DelayInBranch(t *testing.T) {
	// Graph: start -> condition --(true_branch)--> delay(5s) -> true_action --(sequence)--> merge
	//                            \--(false_branch)--> false_action --(sequence)--> merge
	condHandler := &testActionHandler{
		actionType: "evaluate_condition",
		result: workflow.ConditionResult{
			Evaluated:   true,
			Result:      true,
			BranchTaken: EdgeTypeTrueBranch,
		},
	}
	trueHandler := &testActionHandler{
		actionType: "true_action",
		result:     map[string]any{"branch": "true"},
	}
	falseHandler := &testActionHandler{
		actionType: "false_action",
		result:     map[string]any{"branch": "false"},
	}
	mergeHandler := &testActionHandler{
		actionType: "merge_action",
		result:     map[string]any{"merged": true},
	}

	env := setupTestEnv(t, newTestRegistry(condHandler, trueHandler, falseHandler, mergeHandler))

	condID := uuid.New()
	delayID := uuid.New()
	trueID := uuid.New()
	falseID := uuid.New()
	mergeID := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: condID, Name: "condition", ActionType: "evaluate_condition", Config: json.RawMessage(`{}`), IsActive: true},
			{ID: delayID, Name: "wait", ActionType: "delay", Config: json.RawMessage(`{"duration": "5s"}`), IsActive: true},
			{ID: trueID, Name: "true_arm", ActionType: "true_action", Config: json.RawMessage(`{}`), IsActive: true},
			{ID: falseID, Name: "false_arm", ActionType: "false_action", Config: json.RawMessage(`{}`), IsActive: true},
			{ID: mergeID, Name: "merge", ActionType: "merge_action", Config: json.RawMessage(`{}`), IsActive: true},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: condID, EdgeType: EdgeTypeStart, SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &condID, TargetActionID: delayID, EdgeType: EdgeTypeTrueBranch, SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &delayID, TargetActionID: trueID, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &condID, TargetActionID: falseID, EdgeType: EdgeTypeFalseBranch, SortOrder: 2},
			{ID: uuid.New(), SourceActionID: &trueID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &falseID, TargetActionID: mergeID, EdgeType: EdgeTypeSequence, SortOrder: 1},
		},
	}

	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "delay-branch-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{"entity_id": "456"},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Condition was evaluated
	require.Equal(t, 1, condHandler.called)
	// True branch was taken (after delay)
	require.Equal(t, 1, trueHandler.called)
	// False branch was not taken
	require.Equal(t, 0, falseHandler.called)
	// Merge was called
	require.Equal(t, 1, mergeHandler.called)
}

func TestParseDelayConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  json.RawMessage
		wantErr bool
	}{
		{"valid_5s", json.RawMessage(`{"duration":"5s"}`), false},
		{"valid_24h", json.RawMessage(`{"duration":"24h"}`), false},
		{"empty_duration", json.RawMessage(`{"duration":""}`), true},
		{"missing_duration", json.RawMessage(`{}`), true},
		{"negative", json.RawMessage(`{"duration":"-1h"}`), true},
		{"invalid_json", json.RawMessage(`{bad}`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDelayConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

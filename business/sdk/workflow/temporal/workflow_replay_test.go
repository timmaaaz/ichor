package temporal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
)

// =============================================================================
// Integration Test Helpers
// =============================================================================

// startTestWorker creates a Temporal worker for integration tests.
// Uses a unique task queue per test to prevent cross-test interference.
func startTestWorker(t *testing.T, tc temporalclient.Client, taskQueue string, handlers ...*testActionHandler) worker.Worker {
	t.Helper()
	reg := newTestRegistry(handlers...)

	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(ExecuteGraphWorkflow)
	w.RegisterWorkflow(ExecuteBranchUntilConvergence)
	w.RegisterActivity(&Activities{
		Registry:      reg,
		AsyncRegistry: NewAsyncRegistry(),
	})
	require.NoError(t, w.Start())
	return w
}

// fetchHistory retrieves the completed workflow history as a *historypb.History
// for replay testing.
func fetchHistory(t *testing.T, ctx context.Context, tc temporalclient.Client, workflowID, runID string) *historypb.History {
	t.Helper()

	iter := tc.GetWorkflowHistory(ctx, workflowID, runID, false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

	var events []*historypb.HistoryEvent
	for iter.HasNext() {
		event, err := iter.Next()
		require.NoError(t, err)
		events = append(events, event)
	}

	return &historypb.History{Events: events}
}

// =============================================================================
// Integration Tests (Real Temporal Container)
// =============================================================================

func TestIntegration_SimpleWorkflow_RealServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c := foundationtemporal.GetTestContainer(t)
	tc, err := foundationtemporal.NewTestClient(c.HostPort)
	require.NoError(t, err)
	defer tc.Close()

	handler := &testActionHandler{
		actionType: "test_action",
		result:     map[string]any{"status": "done"},
	}
	taskQueue := fmt.Sprintf("test-integration-%s", t.Name())
	w := startTestWorker(t, tc, taskQueue, handler)
	defer w.Stop()

	graph, _ := wfLinearGraph("test_action")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "integration-test-rule",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{"entity_id": "abc"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	run, err := tc.ExecuteWorkflow(ctx, temporalclient.StartWorkflowOptions{
		TaskQueue: taskQueue,
	}, ExecuteGraphWorkflow, input)
	require.NoError(t, err)
	require.NoError(t, run.Get(ctx, nil))
	require.Equal(t, 1, handler.called)
}

func TestIntegration_SequentialChain_RealServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c := foundationtemporal.GetTestContainer(t)
	tc, err := foundationtemporal.NewTestClient(c.HostPort)
	require.NoError(t, err)
	defer tc.Close()

	handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"val": "a"}}
	handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"val": "b"}}
	handler2 := &testActionHandler{actionType: "step_2", result: map[string]any{"val": "c"}}

	taskQueue := fmt.Sprintf("test-integration-%s", t.Name())
	w := startTestWorker(t, tc, taskQueue, handler0, handler1, handler2)
	defer w.Stop()

	graph, _ := wfLinearGraph("step_0", "step_1", "step_2")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "chain-integration",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	run, err := tc.ExecuteWorkflow(ctx, temporalclient.StartWorkflowOptions{
		TaskQueue: taskQueue,
	}, ExecuteGraphWorkflow, input)
	require.NoError(t, err)
	require.NoError(t, run.Get(ctx, nil))

	require.Equal(t, 1, handler0.called)
	require.Equal(t, 1, handler1.called)
	require.Equal(t, 1, handler2.called)
}

// =============================================================================
// Replay Determinism Tests
// =============================================================================

func TestReplay_SimpleSequential(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c := foundationtemporal.GetTestContainer(t)
	tc, err := foundationtemporal.NewTestClient(c.HostPort)
	require.NoError(t, err)
	defer tc.Close()

	// Step 1: Execute workflow to generate history.
	handler := &testActionHandler{
		actionType: "test_action",
		result:     map[string]any{"status": "replayed"},
	}
	taskQueue := fmt.Sprintf("test-replay-%s", t.Name())
	w := startTestWorker(t, tc, taskQueue, handler)
	defer w.Stop()

	graph, _ := wfLinearGraph("test_action")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "replay-test",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{"entity_id": "replay-entity"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	run, err := tc.ExecuteWorkflow(ctx, temporalclient.StartWorkflowOptions{
		TaskQueue: taskQueue,
	}, ExecuteGraphWorkflow, input)
	require.NoError(t, err)
	require.NoError(t, run.Get(ctx, nil))

	// Step 2: Fetch completed workflow history.
	history := fetchHistory(t, ctx, tc, run.GetID(), run.GetRunID())
	require.NotEmpty(t, history.Events, "history should have events")

	// Step 3: Replay with same workflow definitions.
	// Only workflow functions need registration (NOT activities — they're in the history).
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(ExecuteGraphWorkflow)
	replayer.RegisterWorkflow(ExecuteBranchUntilConvergence)

	err = replayer.ReplayWorkflowHistory(nil, history)
	require.NoError(t, err, "replay should not produce non-determinism error")
}

func TestReplay_SequentialChain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c := foundationtemporal.GetTestContainer(t)
	tc, err := foundationtemporal.NewTestClient(c.HostPort)
	require.NoError(t, err)
	defer tc.Close()

	handler0 := &testActionHandler{actionType: "step_0", result: map[string]any{"v": 0}}
	handler1 := &testActionHandler{actionType: "step_1", result: map[string]any{"v": 1}}
	handler2 := &testActionHandler{actionType: "step_2", result: map[string]any{"v": 2}}

	taskQueue := fmt.Sprintf("test-replay-%s", t.Name())
	w := startTestWorker(t, tc, taskQueue, handler0, handler1, handler2)
	defer w.Stop()

	graph, _ := wfLinearGraph("step_0", "step_1", "step_2")
	input := WorkflowInput{
		RuleID:      uuid.New(),
		RuleName:    "replay-chain",
		ExecutionID: uuid.New(),
		Graph:       graph,
		TriggerData: map[string]any{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	run, err := tc.ExecuteWorkflow(ctx, temporalclient.StartWorkflowOptions{
		TaskQueue: taskQueue,
	}, ExecuteGraphWorkflow, input)
	require.NoError(t, err)
	require.NoError(t, run.Get(ctx, nil))

	history := fetchHistory(t, ctx, tc, run.GetID(), run.GetRunID())
	require.NotEmpty(t, history.Events, "history should have events")

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(ExecuteGraphWorkflow)
	replayer.RegisterWorkflow(ExecuteBranchUntilConvergence)

	err = replayer.ReplayWorkflowHistory(nil, history)
	require.NoError(t, err, "sequential chain replay should be deterministic")
}

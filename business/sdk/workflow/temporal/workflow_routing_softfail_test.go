package temporal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// reserveAlertGraph builds: start -> reserve
//
//	reserve --(success)------------> continue
//	reserve --(insufficient_stock)--> alert
//
// It models the over-order pipeline: reserve_inventory routes a stock shortfall
// to an alert via a typed output-port edge, and continues normally on success.
func reserveAlertGraph(reserveType, continueType, alertType string) GraphDefinition {
	reserveID := deterministicUUID("reserve")
	continueID := deterministicUUID("continue")
	alertID := deterministicUUID("alert")

	actions := []ActionNode{
		{ID: reserveID, Name: "reserve", ActionType: reserveType, Config: json.RawMessage(`{}`), IsActive: true},
		{ID: continueID, Name: "continue", ActionType: continueType, Config: json.RawMessage(`{}`), IsActive: true},
		{ID: alertID, Name: "alert", ActionType: alertType, Config: json.RawMessage(`{}`), IsActive: true},
	}

	edges := []ActionEdge{
		{ID: deterministicUUID("e-start"), SourceActionID: nil, TargetActionID: reserveID, EdgeType: EdgeTypeStart, SortOrder: 1},
		{ID: deterministicUUID("e-success"), SourceActionID: &reserveID, TargetActionID: continueID, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("success"), SortOrder: 1},
		{ID: deterministicUUID("e-insufficient"), SourceActionID: &reserveID, TargetActionID: alertID, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("insufficient_stock"), SortOrder: 2},
	}

	return GraphDefinition{Actions: actions, Edges: edges}
}

// TestRouting_SoftInsufficientStock_RoutesToAlert proves the premise the
// reserve_inventory rework relies on: when the handler returns a SOFT
// "insufficient_stock" output (a result map, no error), the graph routes to the
// alert action and the workflow run COMPLETES — it does not fail. This is what
// makes over-orders alertable without crashing the run, and contrasts with
// hard-erroring at the executor (which would fail the run / record it failed).
func TestRouting_SoftInsufficientStock_RoutesToAlert(t *testing.T) {
	reserve := &testActionHandler{
		actionType: "reserve_inventory",
		result:     map[string]any{"output": "insufficient_stock", "status": "failed"},
	}
	cont := &testActionHandler{actionType: "continue_action", result: map[string]any{"output": "success"}}
	alert := &testActionHandler{actionType: "create_alert", result: map[string]any{"output": "success"}}

	env := setupTestEnv(t, newTestRegistry(reserve, cont, alert))

	input := WorkflowInput{
		RuleID:      deterministicUUID("rule-insufficient"),
		RuleName:    "soft-insufficient-routes-to-alert",
		ExecutionID: deterministicUUID("exec-insufficient"),
		Graph:       reserveAlertGraph("reserve_inventory", "continue_action", "create_alert"),
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError(), "a soft insufficient_stock output must route, not fail the run")
	require.Equal(t, 1, reserve.called)
	require.Equal(t, 1, alert.called, "shortfall must route to the alert action")
	require.Equal(t, 0, cont.called, "shortfall must NOT continue down the success path")
}

// TestRouting_Success_SkipsAlert is the discriminating negative: the same graph,
// but reserve returns "success" — so it must take the success edge to continue
// and must NOT fire the alert. Proves the typed output ports actually
// discriminate (the routing isn't just "always follow").
func TestRouting_Success_SkipsAlert(t *testing.T) {
	reserve := &testActionHandler{
		actionType: "reserve_inventory",
		result:     map[string]any{"output": "success", "status": "success"},
	}
	cont := &testActionHandler{actionType: "continue_action", result: map[string]any{"output": "success"}}
	alert := &testActionHandler{actionType: "create_alert", result: map[string]any{"output": "success"}}

	env := setupTestEnv(t, newTestRegistry(reserve, cont, alert))

	input := WorkflowInput{
		RuleID:      deterministicUUID("rule-success"),
		RuleName:    "success-skips-alert",
		ExecutionID: deterministicUUID("exec-success"),
		Graph:       reserveAlertGraph("reserve_inventory", "continue_action", "create_alert"),
		TriggerData: map[string]any{},
	}

	env.ExecuteWorkflow(ExecuteGraphWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, 1, reserve.called)
	require.Equal(t, 1, cont.called, "success must continue down the success path")
	require.Equal(t, 0, alert.called, "success must NOT fire the alert")
}

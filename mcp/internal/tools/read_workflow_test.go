package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestGetWorkflow_Success(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-1":         `{"id":"wf-1","name":"test workflow","trigger_type":"on_create"}`,
		"/v1/workflow/rules/wf-1/actions": `[{"id":"a1","name":"notify","action_type":"send_email"}]`,
		"/v1/workflow/rules/wf-1/edges":   `[{"id":"e1","target_action_id":"a1","edge_type":"start"}]`,
	}

	session, ctx := setupToolTest(t,
		pathRouter(mockRoutes),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "get_workflow", map[string]any{"id": "wf-1"})

	if result.IsError {
		t.Fatalf("get_workflow returned error: %s", getTextContent(t, result))
	}

	// Verify the merged response contains rule, actions, edges, and summary.
	text := getTextContent(t, result)
	var merged map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &merged); err != nil {
		t.Fatalf("failed to parse merged result: %v", err)
	}

	for _, key := range []string{"rule", "actions", "edges", "summary"} {
		if _, ok := merged[key]; !ok {
			t.Errorf("merged result missing key %q", key)
		}
	}
}

func TestGetWorkflow_MissingID(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterWorkflowReadTools,
	)

	// Pass empty string for id to bypass SDK required check but trigger handler validation.
	result := callTool(t, session, ctx, "get_workflow", map[string]any{"id": ""})
	if !result.IsError {
		t.Error("get_workflow should return error when id is empty")
	}
}

func TestGetWorkflow_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "get_workflow", map[string]any{"id": "wf-1"})
	if !result.IsError {
		t.Error("get_workflow should return error on API failure")
	}
}

func TestListWorkflows_Success(t *testing.T) {
	mockData := `[{"id":"wf-1","name":"notify on order"},{"id":"wf-2","name":"approve expenses"}]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/workflow/rules": mockData}),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "list_workflows", nil)

	if result.IsError {
		t.Fatalf("list_workflows returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != mockData {
		t.Errorf("got %q, want %q", text, mockData)
	}
}

func TestListActionTemplates_Success(t *testing.T) {
	mockData := `[{"id":"t1","name":"email notification","is_active":true}]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{"/v1/workflow/templates": mockData}),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "list_action_templates", nil)

	if result.IsError {
		t.Fatalf("list_action_templates returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != mockData {
		t.Errorf("got %q, want %q", text, mockData)
	}
}

func TestListWorkflows_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "list_workflows", nil)
	if !result.IsError {
		t.Error("list_workflows should return error on API failure")
	}
}

func TestExplainWorkflowNode_Success(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-1/actions": `[
			{"id":"a1","name":"check_inventory","action_type":"condition"},
			{"id":"a2","name":"send_email","action_type":"send_email"}
		]`,
		"/v1/workflow/rules/wf-1/edges": `[
			{"id":"e1","target_action_id":"a1","edge_type":"start"},
			{"id":"e2","source_action_id":"a1","target_action_id":"a2","edge_type":"always","source_output":"success"}
		]`,
		"/v1/workflow/action-types/condition/schema": `{"name":"condition","description":"Evaluate a condition"}`,
	}

	session, ctx := setupToolTest(t,
		pathRouter(mockRoutes),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "explain_workflow_node", map[string]any{
		"rule_id":    "wf-1",
		"identifier": "check_inventory",
	})

	if result.IsError {
		t.Fatalf("explain_workflow_node returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var explanation map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &explanation); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	for _, key := range []string{"action", "depth_from_start", "incoming_from", "outgoing_to"} {
		if _, ok := explanation[key]; !ok {
			t.Errorf("result missing key %q", key)
		}
	}
}

func TestExplainWorkflowNode_ByID(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-1/actions": `[{"id":"a1","name":"check_inventory","action_type":"condition"}]`,
		"/v1/workflow/rules/wf-1/edges":   `[{"id":"e1","target_action_id":"a1","edge_type":"start"}]`,
	}

	session, ctx := setupToolTest(t,
		pathRouter(mockRoutes),
		tools.RegisterWorkflowReadTools,
	)

	// Look up by UUID instead of name.
	result := callTool(t, session, ctx, "explain_workflow_node", map[string]any{
		"rule_id":    "wf-1",
		"identifier": "a1",
	})

	if result.IsError {
		t.Fatalf("explain_workflow_node returned error: %s", getTextContent(t, result))
	}
}

func TestExplainWorkflowNode_NotFound(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-1/actions": `[{"id":"a1","name":"check_inventory"}]`,
		"/v1/workflow/rules/wf-1/edges":   `[]`,
	}

	session, ctx := setupToolTest(t,
		pathRouter(mockRoutes),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "explain_workflow_node", map[string]any{
		"rule_id":    "wf-1",
		"identifier": "nonexistent",
	})

	if !result.IsError {
		t.Error("explain_workflow_node should return error for unknown action")
	}
}

func TestExplainWorkflowPath_Success(t *testing.T) {
	// Workflow: start -> check_inventory -> [sufficient] fulfill_order
	//                                    -> [insufficient] send_notification -> create_alert
	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-1/actions": `[
			{"id":"a1","name":"check_inventory","action_type":"condition"},
			{"id":"a2","name":"fulfill_order","action_type":"fulfill"},
			{"id":"a3","name":"send_notification","action_type":"send_notification"},
			{"id":"a4","name":"create_alert","action_type":"create_alert"}
		]`,
		"/v1/workflow/rules/wf-1/edges": `[
			{"id":"e1","target_action_id":"a1","edge_type":"start"},
			{"id":"e2","source_action_id":"a1","target_action_id":"a2","edge_type":"output","source_output":"sufficient","edge_order":1},
			{"id":"e3","source_action_id":"a1","target_action_id":"a3","edge_type":"output","source_output":"insufficient","edge_order":2},
			{"id":"e4","source_action_id":"a3","target_action_id":"a4","edge_type":"always","source_output":"","edge_order":1}
		]`,
	}

	session, ctx := setupToolTest(t, pathRouter(mockRoutes), tools.RegisterWorkflowReadTools)

	// Trace from check_inventory following the "insufficient" output.
	result := callTool(t, session, ctx, "explain_workflow_path", map[string]any{
		"rule_id": "wf-1",
		"from":    "check_inventory",
		"output":  "insufficient",
	})

	if result.IsError {
		t.Fatalf("explain_workflow_path returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var pathRes map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &pathRes); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	for _, key := range []string{"starting_from", "output_followed", "path", "text_outline"} {
		if _, ok := pathRes[key]; !ok {
			t.Errorf("result missing key %q", key)
		}
	}

	// Verify path contains send_notification and create_alert but NOT fulfill_order.
	var steps []map[string]json.RawMessage
	if err := json.Unmarshal(pathRes["path"], &steps); err != nil {
		t.Fatalf("failed to parse path: %v", err)
	}
	if len(steps) != 2 {
		t.Errorf("expected 2 path steps, got %d", len(steps))
	}
}

func TestExplainWorkflowPath_FromStart(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-1/actions": `[
			{"id":"a1","name":"step_one","action_type":"send_email"},
			{"id":"a2","name":"step_two","action_type":"create_alert"}
		]`,
		"/v1/workflow/rules/wf-1/edges": `[
			{"id":"e1","target_action_id":"a1","edge_type":"start"},
			{"id":"e2","source_action_id":"a1","target_action_id":"a2","edge_type":"always","edge_order":1}
		]`,
	}

	session, ctx := setupToolTest(t, pathRouter(mockRoutes), tools.RegisterWorkflowReadTools)

	// Trace from start (no "from" parameter).
	result := callTool(t, session, ctx, "explain_workflow_path", map[string]any{
		"rule_id": "wf-1",
	})

	if result.IsError {
		t.Fatalf("explain_workflow_path returned error: %s", getTextContent(t, result))
	}

	text := getTextContent(t, result)
	var pathRes struct {
		StartingFrom string            `json:"starting_from"`
		Path         []json.RawMessage `json:"path"`
	}
	if err := json.Unmarshal([]byte(text), &pathRes); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if pathRes.StartingFrom != "(start)" {
		t.Errorf("expected starting_from to be '(start)', got %q", pathRes.StartingFrom)
	}
	if len(pathRes.Path) != 2 {
		t.Errorf("expected 2 path steps from start, got %d", len(pathRes.Path))
	}
}

func TestExplainWorkflowPath_OutputWithoutFrom(t *testing.T) {
	session, ctx := setupToolTest(t, staticHandler(`{}`), tools.RegisterWorkflowReadTools)

	result := callTool(t, session, ctx, "explain_workflow_path", map[string]any{
		"rule_id": "wf-1",
		"output":  "insufficient",
	})

	if !result.IsError {
		t.Error("explain_workflow_path should return error when output is given without from")
	}
}

func TestExplainWorkflowPath_MissingRuleID(t *testing.T) {
	session, ctx := setupToolTest(t, staticHandler(`{}`), tools.RegisterWorkflowReadTools)

	result := callTool(t, session, ctx, "explain_workflow_path", map[string]any{
		"rule_id": "",
	})

	if !result.IsError {
		t.Error("explain_workflow_path should return error when rule_id is empty")
	}
}

func TestExplainWorkflowNode_MissingParams(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "explain_workflow_node", map[string]any{
		"rule_id":    "",
		"identifier": "",
	})

	if !result.IsError {
		t.Error("explain_workflow_node should return error when params are empty")
	}
}

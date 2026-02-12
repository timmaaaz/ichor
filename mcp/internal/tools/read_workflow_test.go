package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestGetWorkflow_Success(t *testing.T) {
	mockRoutes := map[string]string{
		"/v1/workflow/rules/wf-1":         `{"id":"wf-1","name":"test workflow","trigger_type":"on_create"}`,
		"/v1/workflow/rules/wf-1/actions": `[{"id":"a1","action_type":"send_email"}]`,
		"/v1/workflow/rules/wf-1/edges":   `[{"id":"e1","source_action_id":"a1"}]`,
	}

	session, ctx := setupToolTest(t,
		pathRouter(mockRoutes),
		tools.RegisterWorkflowReadTools,
	)

	result := callTool(t, session, ctx, "get_workflow", map[string]any{"id": "wf-1"})

	if result.IsError {
		t.Fatalf("get_workflow returned error: %s", getTextContent(t, result))
	}

	// Verify the merged response contains rule, actions, and edges.
	text := getTextContent(t, result)
	var merged map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &merged); err != nil {
		t.Fatalf("failed to parse merged result: %v", err)
	}

	for _, key := range []string{"rule", "actions", "edges"} {
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

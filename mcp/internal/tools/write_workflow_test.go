package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestValidateWorkflow_Success(t *testing.T) {
	validationResult := `{"valid":true,"errors":[]}`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/workflow/rules/full?dry_run=true": validationResult,
		}),
		tools.RegisterWorkflowWriteTools,
	)

	result := callTool(t, session, ctx, "validate_workflow", map[string]any{
		"workflow": map[string]any{
			"rule":    map[string]any{"name": "test"},
			"actions": []any{},
			"edges":   []any{},
		},
	})

	if result.IsError {
		t.Fatalf("validate_workflow returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != validationResult {
		t.Errorf("got %q, want %q", text, validationResult)
	}
}

func TestValidateWorkflow_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterWorkflowWriteTools,
	)

	result := callTool(t, session, ctx, "validate_workflow", map[string]any{
		"workflow": map[string]any{"rule": map[string]any{"name": "test"}},
	})
	if !result.IsError {
		t.Error("validate_workflow should return error on API failure")
	}
}

func TestCreateWorkflow_SkipValidation(t *testing.T) {
	createResult := `{"id":"wf-new","name":"test workflow"}`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/workflow/rules/full": createResult,
		}),
		tools.RegisterWorkflowWriteTools,
	)

	result := callTool(t, session, ctx, "create_workflow", map[string]any{
		"workflow": map[string]any{
			"rule":    map[string]any{"name": "test"},
			"actions": []any{},
			"edges":   []any{},
		},
		"validate": false,
	})

	if result.IsError {
		t.Fatalf("create_workflow returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != createResult {
		t.Errorf("got %q, want %q", text, createResult)
	}
}

func TestCreateWorkflow_WithValidation_Passes(t *testing.T) {
	createResult := `{"id":"wf-new","name":"test workflow"}`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/workflow/rules/full?dry_run=true": `{"valid":true,"errors":[]}`,
			"/v1/workflow/rules/full":              createResult,
		}),
		tools.RegisterWorkflowWriteTools,
	)

	// Explicitly pass validate=true.
	result := callTool(t, session, ctx, "create_workflow", map[string]any{
		"workflow": map[string]any{
			"rule":    map[string]any{"name": "test"},
			"actions": []any{},
			"edges":   []any{},
		},
		"validate": true,
	})

	if result.IsError {
		t.Fatalf("create_workflow returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != createResult {
		t.Errorf("got %q, want %q", text, createResult)
	}
}

func TestCreateWorkflow_ValidationFails(t *testing.T) {
	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/workflow/rules/full?dry_run=true": `{"valid":false,"errors":["missing start edge","cycle detected"]}`,
		}),
		tools.RegisterWorkflowWriteTools,
	)

	result := callTool(t, session, ctx, "create_workflow", map[string]any{
		"workflow": map[string]any{
			"rule":    map[string]any{"name": "invalid"},
			"actions": []any{},
			"edges":   []any{},
		},
		"validate": true,
	})

	// Should NOT be an error result (validation failure is a normal response, not a tool error).
	if result.IsError {
		t.Fatalf("create_workflow should not return IsError for validation failure")
	}

	text := getTextContent(t, result)
	var resp map[string]any
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if created, ok := resp["created"].(bool); !ok || created {
		t.Error("expected created=false in response")
	}
	if errors, ok := resp["validation_errors"].([]any); !ok || len(errors) != 2 {
		t.Errorf("expected 2 validation errors, got %v", resp["validation_errors"])
	}
}

func TestCreateWorkflow_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterWorkflowWriteTools,
	)

	result := callTool(t, session, ctx, "create_workflow", map[string]any{
		"workflow": map[string]any{"rule": map[string]any{"name": "test"}},
		"validate": false,
	})
	if !result.IsError {
		t.Error("create_workflow should return error on API failure")
	}
}

func TestUpdateWorkflow_Success(t *testing.T) {
	updateResult := `{"id":"wf-1","name":"updated workflow"}`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/workflow/rules/full?dry_run=true": `{"valid":true,"errors":[]}`,
			"/v1/workflow/rules/wf-1/full":         updateResult,
		}),
		tools.RegisterWorkflowWriteTools,
	)

	result := callTool(t, session, ctx, "update_workflow", map[string]any{
		"id": "wf-1",
		"workflow": map[string]any{
			"rule":    map[string]any{"name": "updated"},
			"actions": []any{},
			"edges":   []any{},
		},
		"validate": true,
	})

	if result.IsError {
		t.Fatalf("update_workflow returned error: %s", getTextContent(t, result))
	}
	if text := getTextContent(t, result); text != updateResult {
		t.Errorf("got %q, want %q", text, updateResult)
	}
}

func TestUpdateWorkflow_MissingID(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterWorkflowWriteTools,
	)

	// Pass empty string for id to trigger handler validation.
	result := callTool(t, session, ctx, "update_workflow", map[string]any{
		"id":       "",
		"workflow": map[string]any{"rule": map[string]any{"name": "test"}},
		"validate": false,
	})
	if !result.IsError {
		t.Error("update_workflow should return error when id is empty")
	}
}

func TestUpdateWorkflow_ValidationFails(t *testing.T) {
	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/workflow/rules/full?dry_run=true": `{"valid":false,"errors":["invalid graph"]}`,
		}),
		tools.RegisterWorkflowWriteTools,
	)

	result := callTool(t, session, ctx, "update_workflow", map[string]any{
		"id": "wf-1",
		"workflow": map[string]any{
			"rule":    map[string]any{"name": "bad"},
			"actions": []any{},
			"edges":   []any{},
		},
		"validate": true,
	})

	if result.IsError {
		t.Fatal("update_workflow should not return IsError for validation failure")
	}

	text := getTextContent(t, result)
	var resp map[string]any
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if updated, ok := resp["updated"].(bool); !ok || updated {
		t.Error("expected updated=false in response")
	}
}

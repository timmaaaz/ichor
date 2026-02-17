package tools_test

import (
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestPageActionReadTools_GetPageActions_Success(t *testing.T) {
	response := `{"buttons":[{"id":"b1","action_type":"button"}],"dropdowns":[],"separators":[]}`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/config/page-configs/actions/pc-123": response,
		}),
		tools.RegisterPageActionReadTools,
	)

	result := callTool(t, session, ctx, "get_page_actions", map[string]any{
		"page_config_id": "pc-123",
	})

	if result.IsError {
		t.Errorf("get_page_actions returned error: %s", getTextContent(t, result))
	}
	text := getTextContent(t, result)
	if text != response {
		t.Errorf("got %q, want %q", text, response)
	}
}

func TestPageActionReadTools_GetPageAction_Success(t *testing.T) {
	response := `{"id":"act-1","action_type":"button","button":{"label":"Create","variant":"default"}}`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/config/page-actions/act-1": response,
		}),
		tools.RegisterPageActionReadTools,
	)

	result := callTool(t, session, ctx, "get_page_action", map[string]any{
		"action_id": "act-1",
	})

	if result.IsError {
		t.Errorf("get_page_action returned error: %s", getTextContent(t, result))
	}
	text := getTextContent(t, result)
	if text != response {
		t.Errorf("got %q, want %q", text, response)
	}
}

func TestPageActionReadTools_MissingRequiredArgs(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterPageActionReadTools,
	)

	// Empty string passes SDK schema validation but triggers handler validation.
	tests := []struct {
		toolName string
		args     map[string]any
	}{
		{"get_page_actions", map[string]any{"page_config_id": ""}},
		{"get_page_action", map[string]any{"action_id": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := callTool(t, session, ctx, tt.toolName, tt.args)
			if !result.IsError {
				t.Errorf("%s should return error when required args missing", tt.toolName)
			}
		})
	}
}

func TestPageActionReadTools_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterPageActionReadTools,
	)

	tests := []struct {
		toolName string
		args     map[string]any
	}{
		{"get_page_actions", map[string]any{"page_config_id": "abc-123"}},
		{"get_page_action", map[string]any{"action_id": "abc-123"}},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := callTool(t, session, ctx, tt.toolName, tt.args)
			if !result.IsError {
				t.Errorf("%s should return error for 500 response", tt.toolName)
			}
		})
	}
}

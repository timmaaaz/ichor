package tools_test

import (
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestReadUITools_ListSuccess(t *testing.T) {
	tests := []struct {
		toolName string
		path     string
		response string
	}{
		{"list_pages", "/v1/config/page-configs/all", `[{"id":"p1","name":"dashboard"}]`},
		{"list_forms", "/v1/config/forms", `[{"id":"f1","name":"user_form"}]`},
		{"list_table_configs", "/v1/data/configs/all", `[{"id":"t1","name":"users_table"}]`},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			session, ctx := setupToolTest(t,
				pathRouter(map[string]string{tt.path: tt.response}),
				tools.RegisterUIReadTools,
			)

			result := callTool(t, session, ctx, tt.toolName, nil)

			if result.IsError {
				t.Errorf("%s returned error: %s", tt.toolName, getTextContent(t, result))
			}
			text := getTextContent(t, result)
			if text != tt.response {
				t.Errorf("got %q, want %q", text, tt.response)
			}
		})
	}
}

func TestReadUITools_GetByID(t *testing.T) {
	tests := []struct {
		toolName string
		argName  string
		argValue string
		path     string
		response string
	}{
		{"get_page_config", "id", "abc-123", "/v1/config/page-configs/id/abc-123", `{"id":"abc-123","name":"dashboard"}`},
		{"get_page_content", "page_config_id", "abc-123", "/v1/config/page-configs/content/abc-123", `[{"id":"c1","content_type":"table"}]`},
		{"get_table_config", "id", "abc-123", "/v1/data/id/abc-123", `{"id":"abc-123","name":"users_table"}`},
		{"get_form_definition", "id", "abc-123", "/v1/config/forms/abc-123/full", `{"id":"abc-123","name":"user_form","fields":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.toolName+"_by_id", func(t *testing.T) {
			session, ctx := setupToolTest(t,
				pathRouter(map[string]string{tt.path: tt.response}),
				tools.RegisterUIReadTools,
			)

			result := callTool(t, session, ctx, tt.toolName, map[string]any{
				tt.argName: tt.argValue,
			})

			if result.IsError {
				t.Errorf("%s returned error: %s", tt.toolName, getTextContent(t, result))
			}
			text := getTextContent(t, result)
			if text != tt.response {
				t.Errorf("got %q, want %q", text, tt.response)
			}
		})
	}
}

func TestReadUITools_GetByName(t *testing.T) {
	tests := []struct {
		toolName string
		name     string
		path     string
		response string
	}{
		{"get_page_config", "dashboard", "/v1/config/page-configs/name/dashboard", `{"id":"p1","name":"dashboard"}`},
		{"get_table_config", "users_table", "/v1/data/name/users_table", `{"id":"t1","name":"users_table"}`},
		{"get_form_definition", "user_form", "/v1/config/forms/name/user_form/full", `{"id":"f1","name":"user_form","fields":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.toolName+"_by_name", func(t *testing.T) {
			session, ctx := setupToolTest(t,
				pathRouter(map[string]string{tt.path: tt.response}),
				tools.RegisterUIReadTools,
			)

			result := callTool(t, session, ctx, tt.toolName, map[string]any{
				"name": tt.name,
			})

			if result.IsError {
				t.Errorf("%s returned error: %s", tt.toolName, getTextContent(t, result))
			}
			text := getTextContent(t, result)
			if text != tt.response {
				t.Errorf("got %q, want %q", text, tt.response)
			}
		})
	}
}

func TestReadUITools_MissingRequiredArgs(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterUIReadTools,
	)

	tests := []struct {
		toolName string
		args     map[string]any
	}{
		{"get_page_config", nil},              // Neither id nor name
		{"get_page_config", map[string]any{}}, // Empty args
		{"get_table_config", nil},             // Neither id nor name
		{"get_form_definition", nil},          // Neither id nor name
		// page_config_id is required in schema, pass empty string to test handler validation.
		{"get_page_content", map[string]any{"page_config_id": ""}},
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

func TestReadUITools_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterUIReadTools,
	)

	tests := []struct {
		toolName string
		args     map[string]any
	}{
		{"list_pages", nil},
		{"list_forms", nil},
		{"list_table_configs", nil},
		{"get_page_config", map[string]any{"id": "abc-123"}},
		{"get_page_content", map[string]any{"page_config_id": "abc-123"}},
		{"get_table_config", map[string]any{"id": "abc-123"}},
		{"get_form_definition", map[string]any{"id": "abc-123"}},
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

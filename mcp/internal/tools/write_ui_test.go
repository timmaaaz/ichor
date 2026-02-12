package tools_test

import (
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestUIWriteTools_Create_Success(t *testing.T) {
	tests := []struct {
		toolName string
		argKey   string
		argValue map[string]any
		path     string
		response string
	}{
		{"create_page_config", "config", map[string]any{"name": "new page", "module": "sales"}, "/v1/config/page-configs", `{"id":"pc-1","name":"new page"}`},
		{"create_page_content", "content", map[string]any{"content_type": "table", "page_config_id": "pc-1"}, "/v1/config/page-content", `{"id":"cnt-1","content_type":"table"}`},
		{"create_form", "form", map[string]any{"name": "user_form", "entity": "users"}, "/v1/config/forms", `{"id":"f-1","name":"new form"}`},
		{"add_form_field", "field", map[string]any{"form_id": "f-1", "field_type": "text", "name": "email"}, "/v1/config/form-fields", `{"id":"ff-1","field_type":"text"}`},
		{"create_table_config", "config", map[string]any{"name": "users_table", "config": map[string]any{"data_sources": []any{}}}, "/v1/data", `{"id":"tc-1","name":"new table"}`},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			session, ctx := setupToolTest(t,
				pathRouter(map[string]string{tt.path: tt.response}),
				tools.RegisterUIWriteTools,
			)

			result := callTool(t, session, ctx, tt.toolName, map[string]any{
				tt.argKey: tt.argValue,
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

func TestUIWriteTools_Update_Success(t *testing.T) {
	tests := []struct {
		toolName string
		argKey   string
		argValue map[string]any
		path     string
		response string
	}{
		{"update_page_config", "config", map[string]any{"name": "updated page"}, "/v1/config/page-configs/id/abc-123", `{"id":"abc-123","name":"updated page"}`},
		{"update_page_content", "content", map[string]any{"label": "New Label"}, "/v1/config/page-content/abc-123", `{"id":"abc-123","content_type":"form"}`},
		{"update_table_config", "config", map[string]any{"name": "updated table"}, "/v1/data/abc-123", `{"id":"abc-123","name":"updated table"}`},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			session, ctx := setupToolTest(t,
				pathRouter(map[string]string{tt.path: tt.response}),
				tools.RegisterUIWriteTools,
			)

			result := callTool(t, session, ctx, tt.toolName, map[string]any{
				"id":      "abc-123",
				tt.argKey: tt.argValue,
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

func TestUIWriteTools_Update_MissingID(t *testing.T) {
	session, ctx := setupToolTest(t,
		staticHandler(`{}`),
		tools.RegisterUIWriteTools,
	)

	tests := []struct {
		toolName string
		argKey   string
	}{
		{"update_page_config", "config"},
		{"update_page_content", "content"},
		{"update_table_config", "config"},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			// Pass empty string for id to trigger handler validation.
			result := callTool(t, session, ctx, tt.toolName, map[string]any{
				"id":      "",
				tt.argKey: map[string]any{"name": "test"},
			})
			if !result.IsError {
				t.Errorf("%s should return error when id is empty", tt.toolName)
			}
		})
	}
}

func TestUIWriteTools_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterUIWriteTools,
	)

	tests := []struct {
		toolName string
		args     map[string]any
	}{
		{"create_page_config", map[string]any{"config": map[string]any{"name": "test"}}},
		{"update_page_config", map[string]any{"id": "abc", "config": map[string]any{"name": "test"}}},
		{"create_page_content", map[string]any{"content": map[string]any{"content_type": "table"}}},
		{"update_page_content", map[string]any{"id": "abc", "content": map[string]any{"label": "test"}}},
		{"create_form", map[string]any{"form": map[string]any{"name": "test"}}},
		{"add_form_field", map[string]any{"field": map[string]any{"field_type": "text"}}},
		{"create_table_config", map[string]any{"config": map[string]any{"name": "test"}}},
		{"update_table_config", map[string]any{"id": "abc", "config": map[string]any{"name": "test"}}},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			result := callTool(t, session, ctx, tt.toolName, tt.args)
			if !result.IsError {
				t.Errorf("%s should return error on API failure", tt.toolName)
			}
		})
	}
}

package tools_test

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestDiscoveryTools_Success(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		path     string
		response string
	}{
		{"discover_config_surfaces", "discover_config_surfaces", "/v1/agent/catalog", `[{"name":"page_configs","endpoints":{"list":"/v1/config/page-configs/all"}}]`},
		{"discover_action_types", "discover_action_types", "/v1/workflow/action-types", `[{"type":"send_email","category":"notification"}]`},
		{"discover_field_types", "discover_field_types", "/v1/config/form-field-types", `[{"type":"text","label":"Text Input"}]`},
		{"discover_trigger_types", "discover_trigger_types", "/v1/workflow/trigger-types", `["on_create","on_update","on_delete"]`},
		{"discover_entity_types", "discover_entity_types", "/v1/workflow/entity-types", `["orders","products","users"]`},
		{"discover_entities", "discover_entities", "/v1/workflow/entities", `[{"entity":"orders","schema":"sales"}]`},
		{"discover_content_types", "discover_content_types", "/v1/config/schemas/content-types", `[{"type":"table"},{"type":"form"},{"type":"chart"}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, ctx := setupToolTest(t,
				pathRouter(map[string]string{tt.path: tt.response}),
				tools.RegisterDiscoveryTools,
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

func TestDiscoveryTools_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterDiscoveryTools,
	)

	toolNames := []string{
		"discover_config_surfaces",
		"discover_action_types",
		"discover_field_types",
		"discover_trigger_types",
		"discover_entity_types",
		"discover_entities",
		"discover_content_types",
	}

	for _, tn := range toolNames {
		t.Run(tn, func(t *testing.T) {
			result := callTool(t, session, ctx, tn, nil)
			if !result.IsError {
				t.Errorf("%s should return error for 500 response", tn)
			}
		})
	}
}

func TestDiscoveryTools_ReturnsJSONContent(t *testing.T) {
	mockJSON := `{"surfaces":["pages","forms","tables"]}`

	session, ctx := setupToolTest(t,
		staticHandler(mockJSON),
		tools.RegisterDiscoveryTools,
	)

	result := callTool(t, session, ctx, "discover_config_surfaces", nil)

	if result.IsError {
		t.Fatal("unexpected error")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", result.Content[0])
	}
	if tc.Text != mockJSON {
		t.Errorf("content text = %q, want %q", tc.Text, mockJSON)
	}
}

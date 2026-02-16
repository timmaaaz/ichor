package tools_test

import (
	"testing"

	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

func TestPageActionDiscoveryTools_Success(t *testing.T) {
	response := `[{"type":"button","name":"Button"},{"type":"dropdown","name":"Dropdown Menu"},{"type":"separator","name":"Separator"}]`

	session, ctx := setupToolTest(t,
		pathRouter(map[string]string{
			"/v1/config/schemas/page-action-types": response,
		}),
		tools.RegisterPageActionDiscoveryTools,
	)

	result := callTool(t, session, ctx, "discover_page_action_types", nil)

	if result.IsError {
		t.Errorf("discover_page_action_types returned error: %s", getTextContent(t, result))
	}
	text := getTextContent(t, result)
	if text != response {
		t.Errorf("got %q, want %q", text, response)
	}
}

func TestPageActionDiscoveryTools_APIError(t *testing.T) {
	session, ctx := setupToolTest(t,
		errorHandler(500),
		tools.RegisterPageActionDiscoveryTools,
	)

	result := callTool(t, session, ctx, "discover_page_action_types", nil)
	if !result.IsError {
		t.Errorf("discover_page_action_types should return error for 500 response")
	}
}

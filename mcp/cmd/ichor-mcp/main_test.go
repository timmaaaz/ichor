package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
	"github.com/timmaaaz/ichor/mcp/internal/prompts"
	"github.com/timmaaaz/ichor/mcp/internal/resources"
	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

// setupIntegrationTest creates a full MCP server with all tools, resources, and prompts
// registered against a mock HTTP backend. Returns a connected client session.
func setupIntegrationTest(t *testing.T, handler http.Handler) (*mcp.ClientSession, context.Context) {
	t.Helper()

	mock := httptest.NewServer(handler)
	t.Cleanup(mock.Close)

	ichorClient := client.New(mock.URL, "test-token")

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ichor-mcp-test",
		Version: "0.0.1",
	}, nil)

	tools.RegisterDiscoveryTools(server, ichorClient)
	tools.RegisterUIReadTools(server, ichorClient)
	tools.RegisterWorkflowReadTools(server, ichorClient)
	tools.RegisterSearchTools(server, ichorClient)
	tools.RegisterWorkflowWriteTools(server, ichorClient)
	tools.RegisterUIWriteTools(server, ichorClient)
	tools.RegisterValidationTools(server, ichorClient)
	tools.RegisterAnalysisTools(server, ichorClient)
	prompts.RegisterPrompts(server, ichorClient)
	resources.RegisterResources(server, ichorClient)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		server.Connect(ctx, serverTransport, nil)
	}()

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return session, ctx
}

func TestMCPServer_ToolsAndResources(t *testing.T) {
	// Mock Ichor API that returns empty JSON for all requests.
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer mock.Close()

	ichorClient := client.New(mock.URL, "test-token")

	// Create and configure the MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ichor-mcp-test",
		Version: "0.0.1",
	}, nil)

	tools.RegisterDiscoveryTools(server, ichorClient)
	tools.RegisterUIReadTools(server, ichorClient)
	tools.RegisterWorkflowReadTools(server, ichorClient)
	tools.RegisterSearchTools(server, ichorClient)
	tools.RegisterWorkflowWriteTools(server, ichorClient)
	tools.RegisterUIWriteTools(server, ichorClient)
	tools.RegisterValidationTools(server, ichorClient)
	tools.RegisterAnalysisTools(server, ichorClient)
	prompts.RegisterPrompts(server, ichorClient)
	resources.RegisterResources(server, ichorClient)

	// Connect using in-memory transport.
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the server connection in the background.
	serverDone := make(chan error, 1)
	go func() {
		_, err := server.Connect(ctx, serverTransport, nil)
		serverDone <- err
	}()

	// Create a client and connect.
	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	// List tools.
	toolResult, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	expectedTools := []string{
		// Discovery tools
		"discover_config_surfaces",
		"discover_action_types",
		"discover_field_types",
		"discover_trigger_types",
		"discover_entity_types",
		"discover_entities",
		"discover_content_types",
		// UI read tools
		"get_page_config",
		"get_page_content",
		"get_table_config",
		"get_form_definition",
		"list_pages",
		"list_forms",
		"list_table_configs",
		// Workflow read tools
		"get_workflow",
		"list_workflows",
		"list_action_templates",
		// Search tools
		"search_database_schema",
		"search_enums",
		// Workflow write tools
		"validate_workflow",
		"create_workflow",
		"update_workflow",
		// UI write tools
		"create_page_config",
		"update_page_config",
		"create_page_content",
		"update_page_content",
		"create_form",
		"add_form_field",
		"create_table_config",
		"update_table_config",
		// Validation tools
		"validate_table_config",
		// Analysis tools
		"analyze_workflow",
		"suggest_templates",
		"show_cascade",
	}

	toolNames := make(map[string]bool)
	for _, tool := range toolResult.Tools {
		toolNames[tool.Name] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("expected tool %q not found in tool list", expected)
		}
	}

	if len(toolResult.Tools) != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), len(toolResult.Tools))
		for _, tool := range toolResult.Tools {
			t.Logf("  registered tool: %s", tool.Name)
		}
	}

	// List resources.
	resourceResult, err := session.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}

	expectedResources := []string{
		"config://catalog",
		"config://action-types",
		"config://field-types",
		"config://table-config-schema",
		"config://layout-schema",
	}

	resourceURIs := make(map[string]bool)
	for _, res := range resourceResult.Resources {
		resourceURIs[res.URI] = true
	}

	for _, expected := range expectedResources {
		if !resourceURIs[expected] {
			t.Errorf("expected resource %q not found in resource list", expected)
		}
	}

	// List resource templates.
	templateResult, err := session.ListResourceTemplates(ctx, nil)
	if err != nil {
		t.Fatalf("list resource templates: %v", err)
	}

	expectedTemplates := []string{
		"config://db/{schema}/{table}",
		"config://enums/{schema}/{name}",
	}

	templateURIs := make(map[string]bool)
	for _, tmpl := range templateResult.ResourceTemplates {
		templateURIs[tmpl.URITemplate] = true
	}

	for _, expected := range expectedTemplates {
		if !templateURIs[expected] {
			t.Errorf("expected resource template %q not found", expected)
		}
	}

	// List prompts.
	promptResult, err := session.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("list prompts: %v", err)
	}

	expectedPrompts := []string{
		"build-workflow",
		"configure-page",
		"design-form",
	}

	promptNames := make(map[string]bool)
	for _, p := range promptResult.Prompts {
		promptNames[p.Name] = true
	}

	for _, expected := range expectedPrompts {
		if !promptNames[expected] {
			t.Errorf("expected prompt %q not found in prompt list", expected)
		}
	}

	if len(promptResult.Prompts) != len(expectedPrompts) {
		t.Errorf("expected %d prompts, got %d", len(expectedPrompts), len(promptResult.Prompts))
	}

	// Call a discovery tool to verify it works end-to-end.
	callResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "discover_config_surfaces",
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if callResult.IsError {
		t.Error("discover_config_surfaces returned error")
	}
	if len(callResult.Content) == 0 {
		t.Error("discover_config_surfaces returned no content")
	}
}

// TestMCPServer_EndToEnd_AllCategories calls at least one tool from each category
// to verify end-to-end registration and request routing.
func TestMCPServer_EndToEnd_AllCategories(t *testing.T) {
	// Mock API with realistic responses per path.
	routes := map[string]string{
		// Discovery
		"/v1/agent/catalog":                `[{"name":"page_configs"}]`,
		"/v1/workflow/action-types":         `[{"type":"send_email"}]`,
		"/v1/config/form-field-types":       `[{"type":"text"}]`,
		"/v1/workflow/trigger-types":         `["on_create"]`,
		"/v1/workflow/entity-types":          `["orders"]`,
		"/v1/workflow/entities":              `[{"entity":"orders"}]`,
		"/v1/config/schemas/content-types":   `[{"type":"table"}]`,

		// UI Read
		"/v1/config/page-configs/all":                  `[{"id":"p1"}]`,
		"/v1/config/page-configs/id/p1":                `{"id":"p1","name":"dashboard"}`,
		"/v1/config/page-configs/content/p1":           `[{"id":"c1"}]`,
		"/v1/data/configs/all":                         `[{"id":"t1"}]`,
		"/v1/data/id/t1":                               `{"id":"t1","name":"users"}`,
		"/v1/config/forms":                             `[{"id":"f1"}]`,
		"/v1/config/forms/f1/full":                     `{"id":"f1","fields":[]}`,

		// Workflow Read
		"/v1/workflow/rules":                `[{"id":"wf-1"}]`,
		"/v1/workflow/rules/wf-1":           `{"id":"wf-1","name":"test"}`,
		"/v1/workflow/rules/wf-1/actions":   `[{"id":"a1","action_type":"send_email"}]`,
		"/v1/workflow/rules/wf-1/edges":     `[{"id":"e1"}]`,
		"/v1/workflow/templates":             `[{"id":"t1"}]`,
		"/v1/workflow/templates/active":      `[{"id":"t1","is_active":true}]`,

		// Search
		"/v1/introspection/schemas":                          `["core","hr"]`,
		"/v1/introspection/schemas/core/tables":              `["users","roles"]`,
		"/v1/introspection/tables/core/users/columns":        `[{"name":"id"}]`,
		"/v1/introspection/tables/core/users/relationships":  `[{"column":"role_id"}]`,
		"/v1/introspection/enums/core":                       `["role_type"]`,
		"/v1/config/enums/core/role_type/options":            `[{"value":"admin"}]`,

		// Write (all POST/PUT return mock created/updated response)
		"/v1/workflow/rules/full":                     `{"id":"wf-new"}`,
		"/v1/workflow/rules/full?dry_run=true":        `{"valid":true,"errors":[]}`,
		"/v1/config/page-configs":                     `{"id":"pc-new"}`,
		"/v1/config/page-content":                     `{"id":"cnt-new"}`,
		"/v1/config/forms/new":                        `{"id":"f-new"}`,
		"/v1/config/form-fields":                      `{"id":"ff-new"}`,
		"/v1/data":                                    `{"id":"tc-new"}`,
		"/v1/data/validate":                           `{"valid":true}`,

		// Resources
		"/v1/config/schemas/table-config":  `{"type":"object"}`,
		"/v1/config/schemas/layout":        `{"type":"object"}`,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		if resp, ok := routes[path]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(resp))
			return
		}
		// Fallback: try without query.
		if resp, ok := routes[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(resp))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})

	session, ctx := setupIntegrationTest(t, handler)

	// Helper to call a tool and verify success.
	callAndVerify := func(name string, args map[string]any) {
		t.Helper()
		result, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		})
		if err != nil {
			t.Errorf("CallTool(%s): %v", name, err)
			return
		}
		if result.IsError {
			tc, _ := result.Content[0].(*mcp.TextContent)
			t.Errorf("%s returned error: %s", name, tc.Text)
			return
		}
		if len(result.Content) == 0 {
			t.Errorf("%s returned no content", name)
		}
	}

	// ===== Discovery (1 tool) =====
	callAndVerify("discover_config_surfaces", nil)

	// ===== UI Read (3 tools) =====
	callAndVerify("list_pages", nil)
	callAndVerify("get_page_config", map[string]any{"id": "p1"})
	callAndVerify("get_page_content", map[string]any{"page_config_id": "p1"})

	// ===== Workflow Read (2 tools) =====
	callAndVerify("list_workflows", nil)
	callAndVerify("get_workflow", map[string]any{"id": "wf-1"})

	// ===== Search (2 tools, all modes) =====
	callAndVerify("search_database_schema", nil)                                           // schemas
	callAndVerify("search_database_schema", map[string]any{"schema": "core"})              // tables
	callAndVerify("search_database_schema", map[string]any{"schema": "core", "table": "users"}) // columns
	callAndVerify("search_enums", map[string]any{"schema": "core"})                        // enum types
	callAndVerify("search_enums", map[string]any{"schema": "core", "name": "role_type"})   // enum values

	// ===== Workflow Write (2 tools) =====
	callAndVerify("validate_workflow", map[string]any{
		"workflow": map[string]any{
			"rule":    map[string]any{"name": "test"},
			"actions": []any{},
			"edges":   []any{},
		},
	})
	callAndVerify("create_workflow", map[string]any{
		"workflow": map[string]any{
			"rule":    map[string]any{"name": "test"},
			"actions": []any{},
			"edges":   []any{},
		},
		"validate": false,
	})

	// ===== UI Write (2 tools) =====
	callAndVerify("create_page_config", map[string]any{
		"config": map[string]any{"name": "new page", "module": "sales"},
	})
	callAndVerify("add_form_field", map[string]any{
		"field": map[string]any{"form_id": "f-1", "field_type": "text", "name": "email"},
	})

	// ===== Validation (1 tool) =====
	callAndVerify("validate_table_config", map[string]any{
		"config": map[string]any{"name": "test_table", "config": map[string]any{"data_sources": []any{}}},
	})

	// ===== Analysis (2 tools) =====
	callAndVerify("analyze_workflow", map[string]any{"id": "wf-1"})
	callAndVerify("suggest_templates", map[string]any{"use_case": "notify on low inventory"})
	callAndVerify("show_cascade", map[string]any{"entity": "orders"})

	// ===== Resources (2 static + 2 templates) =====
	resourceTests := []struct {
		name string
		uri  string
	}{
		{"catalog", "config://catalog"},
		{"action-types", "config://action-types"},
		{"db-template", "config://db/core/users"},
		{"enum-template", "config://enums/core/role_type"},
	}
	for _, rt := range resourceTests {
		t.Run("resource_"+rt.name, func(t *testing.T) {
			result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
				URI: rt.uri,
			})
			if err != nil {
				t.Errorf("ReadResource(%s): %v", rt.uri, err)
				return
			}
			if len(result.Contents) == 0 {
				t.Errorf("ReadResource(%s) returned no content", rt.uri)
			}
		})
	}

	// ===== Prompts (all 3) =====
	promptTests := []struct {
		name string
		args map[string]string
	}{
		{"build-workflow", map[string]string{"trigger": "on_create", "entity": "orders"}},
		{"configure-page", map[string]string{"entity": "products"}},
		{"design-form", map[string]string{"entity": "invoices"}},
	}
	for _, pt := range promptTests {
		t.Run("prompt_"+pt.name, func(t *testing.T) {
			result, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
				Name:      pt.name,
				Arguments: pt.args,
			})
			if err != nil {
				t.Errorf("GetPrompt(%s): %v", pt.name, err)
				return
			}
			if len(result.Messages) == 0 {
				t.Errorf("GetPrompt(%s) returned no messages", pt.name)
			}
			// Verify the prompt text contains the entity argument.
			tc, ok := result.Messages[0].Content.(*mcp.TextContent)
			if !ok {
				t.Errorf("expected TextContent, got %T", result.Messages[0].Content)
				return
			}
			entity := pt.args["entity"]
			if entity == "" {
				entity = pt.args["trigger"]
			}
			if !strings.Contains(tc.Text, entity) {
				t.Errorf("prompt text should contain %q", entity)
			}
		})
	}
}

// TestMCPServer_EndToEnd_ErrorHandling verifies that tools return IsError
// when the backend API is unavailable.
func TestMCPServer_EndToEnd_ErrorHandling(t *testing.T) {
	errorServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"backend unavailable"}`))
	})

	session, ctx := setupIntegrationTest(t, errorServer)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"discover_config_surfaces", nil},
		{"list_pages", nil},
		{"list_workflows", nil},
		{"search_database_schema", nil},
		{"get_workflow", map[string]any{"id": "wf-1"}},
		{"validate_table_config", map[string]any{"config": map[string]any{"name": "test"}}},
		{"analyze_workflow", map[string]any{"id": "wf-1"}},
		{"show_cascade", map[string]any{"entity": "orders"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      tt.name,
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("CallTool(%s): %v", tt.name, err)
			}
			if !result.IsError {
				tc, _ := result.Content[0].(*mcp.TextContent)
				t.Errorf("%s should return IsError=true on backend failure, got content: %s", tt.name, tc.Text)
			}
		})
	}

	// Verify resources also fail gracefully.
	_, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "config://catalog"})
	if err == nil {
		t.Error("ReadResource should return error on backend failure")
	}
}

// TestMCPServer_ToolContent_IsJSON verifies that tool responses contain valid JSON.
func TestMCPServer_ToolContent_IsJSON(t *testing.T) {
	session, ctx := setupIntegrationTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"test-123","name":"example"}`))
	}))

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "discover_config_surfaces",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	// Verify the content is valid JSON.
	if !json.Valid([]byte(tc.Text)) {
		t.Errorf("tool response is not valid JSON: %s", tc.Text)
	}
}

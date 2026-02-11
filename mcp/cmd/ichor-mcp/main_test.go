package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
	"github.com/timmaaaz/ichor/mcp/internal/prompts"
	"github.com/timmaaaz/ichor/mcp/internal/resources"
	"github.com/timmaaaz/ichor/mcp/internal/tools"
)

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

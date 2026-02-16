// Package tools provides MCP tool handlers that wrap the Ichor REST API.
package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterDiscoveryTools adds all discovery-related tools to the MCP server.
func RegisterDiscoveryTools(s *mcp.Server, c *client.Client) {
	RegisterWorkflowDiscoveryTools(s, c)
	RegisterTablesDiscoveryTools(s, c)
}

// RegisterWorkflowDiscoveryTools adds workflow-related discovery tools.
func RegisterWorkflowDiscoveryTools(s *mcp.Server, c *client.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_action_types",
		Description: "List all workflow action types with their JSON config schemas, output ports, categories, and metadata. Use this to understand what actions are available when building workflows.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetActionTypes(ctx)
		if err != nil {
			return errorResult("Failed to fetch action types: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_trigger_types",
		Description: "List the available trigger types for workflow rules (e.g., on_create, on_update, on_delete).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetTriggerTypes(ctx)
		if err != nil {
			return errorResult("Failed to fetch trigger types: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_entity_types",
		Description: "List the available entity types that can trigger workflows.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetEntityTypes(ctx)
		if err != nil {
			return errorResult("Failed to fetch entity types: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_entities",
		Description: "List the specific entities registered for workflow triggers.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetEntities(ctx)
		if err != nil {
			return errorResult("Failed to fetch entities: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

// RegisterTablesDiscoveryTools adds tables/UI-related discovery tools,
// including page action type discovery.
func RegisterTablesDiscoveryTools(s *mcp.Server, c *client.Client) {
	RegisterPageActionDiscoveryTools(s, c)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_config_surfaces",
		Description: "List all configurable surfaces in Ichor (pages, forms, tables, workflows, alerts, permissions, etc.) with their CRUD endpoint URLs, discovery links, and constraints.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetCatalog(ctx)
		if err != nil {
			return errorResult("Failed to fetch catalog: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_field_types",
		Description: "List all form field types (text, number, dropdown, date, etc.) with their JSON config schemas. Use this to understand what field types are available when building forms.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetFieldTypes(ctx)
		if err != nil {
			return errorResult("Failed to fetch field types: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_content_types",
		Description: "List valid content types for page content blocks (table, form, chart, tabs, container, text) with their requirements and capabilities.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetContentTypes(ctx)
		if err != nil {
			return errorResult("Failed to fetch content types: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

// jsonResult creates a CallToolResult with JSON text content.
func jsonResult(data json.RawMessage) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}
}

// errorResult creates a CallToolResult indicating an error.
func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}
}

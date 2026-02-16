package tools

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterPageWriteTools adds page-level mutation tools to the MCP server.
func RegisterPageWriteTools(s *mcp.Server, c *client.Client) {
	// create_page_config — create a new page configuration.
	type CreatePageConfigArgs struct {
		Config json.RawMessage `json:"config"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_page_config",
		Description: "Create a new page configuration. A page config is the top-level container that holds content blocks and actions.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"config": {Type: "object", Description: "Page config payload with name and other metadata"},
			},
			Required: []string{"config"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreatePageConfigArgs) (*mcp.CallToolResult, any, error) {
		data, err := c.CreatePageConfig(ctx, args.Config)
		if err != nil {
			return errorResult("Failed to create page config: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// update_page_config — update an existing page configuration.
	type UpdatePageConfigArgs struct {
		ID     string          `json:"id"`
		Config json.RawMessage `json:"config"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_page_config",
		Description: "Update an existing page configuration.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"id":     {Type: "string", Description: "UUID of the page config to update"},
				"config": {Type: "object", Description: "Updated page config payload"},
			},
			Required: []string{"id", "config"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdatePageConfigArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" {
			return errorResult("id is required"), nil, nil
		}
		data, err := c.UpdatePageConfig(ctx, args.ID, args.Config)
		if err != nil {
			return errorResult("Failed to update page config: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// create_page_content — add a content block to a page.
	type CreatePageContentArgs struct {
		Content json.RawMessage `json:"content"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_page_content",
		Description: "Add a content block (table, form, chart, container, tabs, or text) to a page config. Use discover_content_types to see available types and their requirements.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"content": {Type: "object", Description: "Page content payload with content_type, layout, and references"},
			},
			Required: []string{"content"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreatePageContentArgs) (*mcp.CallToolResult, any, error) {
		data, err := c.CreatePageContent(ctx, args.Content)
		if err != nil {
			return errorResult("Failed to create page content: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// update_page_content — update an existing content block.
	type UpdatePageContentArgs struct {
		ID      string          `json:"id"`
		Content json.RawMessage `json:"content"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_page_content",
		Description: "Update an existing page content block (change layout, label, visibility).",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"id":      {Type: "string", Description: "UUID of the content block to update"},
				"content": {Type: "object", Description: "Updated content payload (label, layout, visibility)"},
			},
			Required: []string{"id", "content"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdatePageContentArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" {
			return errorResult("id is required"), nil, nil
		}
		data, err := c.UpdatePageContent(ctx, args.ID, args.Content)
		if err != nil {
			return errorResult("Failed to update page content: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

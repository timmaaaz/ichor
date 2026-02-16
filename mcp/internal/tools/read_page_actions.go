package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterPageActionReadTools adds page action read tools to the MCP server.
func RegisterPageActionReadTools(s *mcp.Server, c *client.Client) {
	// get_page_actions — get all actions for a page config, grouped by type.
	type GetPageActionsArgs struct {
		PageConfigID string `json:"page_config_id" jsonschema:"UUID of the page config to get actions for,required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_page_actions",
		Description: "Get all page actions (buttons, dropdowns, separators) for a page config, grouped by type. Use discover_page_action_types to see available action types and their fields.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetPageActionsArgs) (*mcp.CallToolResult, any, error) {
		if args.PageConfigID == "" {
			return errorResult("page_config_id is required"), nil, nil
		}
		data, err := c.GetPageActions(ctx, args.PageConfigID)
		if err != nil {
			return errorResult("Failed to fetch page actions: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// get_page_action — get a single page action by ID with full type-specific details.
	type GetPageActionArgs struct {
		ActionID string `json:"action_id" jsonschema:"UUID of the page action,required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_page_action",
		Description: "Get a single page action by ID with full type-specific details (button config, dropdown items, etc.).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetPageActionArgs) (*mcp.CallToolResult, any, error) {
		if args.ActionID == "" {
			return errorResult("action_id is required"), nil, nil
		}
		data, err := c.GetPageAction(ctx, args.ActionID)
		if err != nil {
			return errorResult("Failed to fetch page action: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

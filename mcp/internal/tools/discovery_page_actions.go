package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterPageActionDiscoveryTools adds page action discovery tools to the MCP server.
func RegisterPageActionDiscoveryTools(s *mcp.Server, c *client.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_page_action_types",
		Description: "List all page action types (button, dropdown, separator) with their fields, required fields, and enum values. Use this to understand what action types are available when adding actions to a page.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetPageActionTypes(ctx)
		if err != nil {
			return errorResult("Failed to fetch page action types: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

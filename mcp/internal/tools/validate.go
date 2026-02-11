package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterValidationTools adds validation-only tools to the MCP server.
func RegisterValidationTools(s *mcp.Server, c *client.Client) {
	// validate_table_config — validate a table config without saving.
	type ValidateTableConfigArgs struct {
		Config json.RawMessage `json:"config" jsonschema:"Table config JSONB payload to validate,required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "validate_table_config",
		Description: "Validate a table/widget configuration without saving it. Returns validation errors if the config is invalid (missing column types, invalid data source, etc.).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ValidateTableConfigArgs) (*mcp.CallToolResult, any, error) {
		data, err := c.ValidateTableConfig(ctx, args.Config)
		if err != nil {
			return errorResult("Validation failed: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

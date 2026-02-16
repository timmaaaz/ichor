package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterContentBlockReadTools adds content block read tools (tables, forms)
// to the MCP server.
func RegisterContentBlockReadTools(s *mcp.Server, c *client.Client) {
	// get_table_config — get a table config by ID or name.
	type GetTableConfigArgs struct {
		ID   string `json:"id,omitempty" jsonschema:"UUID of the table config. Provide either id or name."`
		Name string `json:"name,omitempty" jsonschema:"Name of the table config. Provide either id or name."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_table_config",
		Description: "Get a table/widget configuration by ID or name, including its full JSONB config (data sources, columns, visual settings).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetTableConfigArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" && args.Name == "" {
			return errorResult("Either 'id' or 'name' must be provided"), nil, nil
		}
		var data []byte
		var err error
		if args.ID != "" {
			data, err = c.GetTableConfig(ctx, args.ID)
		} else {
			data, err = c.GetTableConfigByName(ctx, args.Name)
		}
		if err != nil {
			return errorResult("Failed to fetch table config: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// get_form_definition — get a form with all its fields.
	type GetFormArgs struct {
		ID   string `json:"id,omitempty" jsonschema:"UUID of the form. Provide either id or name."`
		Name string `json:"name,omitempty" jsonschema:"Name of the form. Provide either id or name."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_form_definition",
		Description: "Get a form definition with all its fields, field types, validation rules, and configuration. Returns the full form structure needed to render or modify it.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetFormArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" && args.Name == "" {
			return errorResult("Either 'id' or 'name' must be provided"), nil, nil
		}
		var data []byte
		var err error
		if args.ID != "" {
			data, err = c.GetFormFull(ctx, args.ID)
		} else {
			data, err = c.GetFormByNameFull(ctx, args.Name)
		}
		if err != nil {
			return errorResult("Failed to fetch form: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// list_table_configs — list all table configs.
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_table_configs",
		Description: "List all table/widget configurations in the system.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetTableConfigs(ctx)
		if err != nil {
			return errorResult("Failed to list table configs: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// list_forms — list all forms.
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_forms",
		Description: "List all form definitions in the system.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetForms(ctx)
		if err != nil {
			return errorResult("Failed to list forms: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

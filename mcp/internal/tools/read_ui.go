package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterUIReadTools adds read-only UI config tools to the MCP server.
func RegisterUIReadTools(s *mcp.Server, c *client.Client) {
	// get_page_config — get a page config by ID or name.
	type GetPageConfigArgs struct {
		ID   string `json:"id,omitempty" jsonschema:"UUID of the page config. Provide either id or name."`
		Name string `json:"name,omitempty" jsonschema:"Name of the page config. Provide either id or name."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_page_config",
		Description: "Get a page configuration by ID or name, including its metadata.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetPageConfigArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" && args.Name == "" {
			return errorResult("Either 'id' or 'name' must be provided"), nil, nil
		}
		var data []byte
		var err error
		if args.ID != "" {
			data, err = c.GetPageConfig(ctx, args.ID)
		} else {
			data, err = c.GetPageConfigByName(ctx, args.Name)
		}
		if err != nil {
			return errorResult("Failed to fetch page config: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// get_page_content — get all content blocks for a page config.
	type GetPageContentArgs struct {
		PageConfigID string `json:"page_config_id" jsonschema:"UUID of the page config to get content for,required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_page_content",
		Description: "Get all content blocks (tables, forms, charts, containers) for a page config, including their layout settings.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetPageContentArgs) (*mcp.CallToolResult, any, error) {
		if args.PageConfigID == "" {
			return errorResult("page_config_id is required"), nil, nil
		}
		data, err := c.GetPageContent(ctx, args.PageConfigID)
		if err != nil {
			return errorResult("Failed to fetch page content: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

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

	// list_pages — list all page configs.
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_pages",
		Description: "List all page configurations in the system.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetPageConfigs(ctx)
		if err != nil {
			return errorResult("Failed to list pages: " + err.Error()), nil, nil
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
}

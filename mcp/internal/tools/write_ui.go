package tools

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterUIWriteTools adds UI config mutation tools to the MCP server.
func RegisterUIWriteTools(s *mcp.Server, c *client.Client) {
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

	// create_form — create a new form definition.
	type CreateFormArgs struct {
		Form json.RawMessage `json:"form"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_form",
		Description: "Create a new form definition for data entry. After creating, add fields with add_form_field.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"form": {Type: "object", Description: "Form payload with name, entity, and metadata"},
			},
			Required: []string{"form"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateFormArgs) (*mcp.CallToolResult, any, error) {
		data, err := c.CreateForm(ctx, args.Form)
		if err != nil {
			return errorResult("Failed to create form: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// add_form_field — add a field to a form.
	type AddFormFieldArgs struct {
		Field json.RawMessage `json:"field"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_form_field",
		Description: "Add a field to a form. Use discover_field_types to see available types and their config schemas.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"field": {Type: "object", Description: "Form field payload with form_id, field_type, name, and config"},
			},
			Required: []string{"field"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args AddFormFieldArgs) (*mcp.CallToolResult, any, error) {
		data, err := c.CreateFormField(ctx, args.Field)
		if err != nil {
			return errorResult("Failed to create form field: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// create_table_config — create a new table/widget config.
	type CreateTableConfigArgs struct {
		Config json.RawMessage `json:"config"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_table_config",
		Description: "Create a new table/widget configuration. The config JSONB schema is available via the config://table-config-schema resource.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"config": {Type: "object", Description: "Table config payload with name and JSONB config"},
			},
			Required: []string{"config"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateTableConfigArgs) (*mcp.CallToolResult, any, error) {
		data, err := c.CreateTableConfig(ctx, args.Config)
		if err != nil {
			return errorResult("Failed to create table config: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// update_table_config — update a table/widget config.
	type UpdateTableConfigArgs struct {
		ID     string          `json:"id"`
		Config json.RawMessage `json:"config"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_table_config",
		Description: "Update an existing table/widget configuration.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"id":     {Type: "string", Description: "UUID of the table config to update"},
				"config": {Type: "object", Description: "Updated table config payload"},
			},
			Required: []string{"id", "config"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateTableConfigArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" {
			return errorResult("id is required"), nil, nil
		}
		data, err := c.UpdateTableConfig(ctx, args.ID, args.Config)
		if err != nil {
			return errorResult("Failed to update table config: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

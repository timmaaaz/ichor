package tools

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterContentBlockWriteTools adds content block mutation tools (forms, tables)
// to the MCP server.
func RegisterContentBlockWriteTools(s *mcp.Server, c *client.Client) {
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

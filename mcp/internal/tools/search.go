package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterSearchTools adds database introspection tools to the MCP server.
func RegisterSearchTools(s *mcp.Server, c *client.Client) {
	// search_database_schema — browse database structure.
	type SearchDBArgs struct {
		Schema string `json:"schema,omitempty" jsonschema:"PostgreSQL schema name (e.g. 'core', 'inventory'). If omitted, lists all schemas."`
		Table  string `json:"table,omitempty" jsonschema:"Table name. If provided with schema, returns columns and relationships."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_database_schema",
		Description: "Browse the database schema structure. Call with no args to list schemas, with schema to list tables, or with schema+table to get columns and relationships.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchDBArgs) (*mcp.CallToolResult, any, error) {
		// No args: list all schemas.
		if args.Schema == "" {
			data, err := c.GetSchemas(ctx)
			if err != nil {
				return errorResult("Failed to list schemas: " + err.Error()), nil, nil
			}
			return jsonResult(data), nil, nil
		}

		// Schema only: list tables.
		if args.Table == "" {
			data, err := c.GetTables(ctx, args.Schema)
			if err != nil {
				return errorResult("Failed to list tables: " + err.Error()), nil, nil
			}
			return jsonResult(data), nil, nil
		}

		// Schema + table: get columns and relationships.
		columns, err := c.GetColumns(ctx, args.Schema, args.Table)
		if err != nil {
			return errorResult("Failed to get columns: " + err.Error()), nil, nil
		}

		relationships, err := c.GetRelationships(ctx, args.Schema, args.Table)
		if err != nil {
			return errorResult("Failed to get relationships: " + err.Error()), nil, nil
		}

		result := map[string]json.RawMessage{
			"columns":       columns,
			"relationships": relationships,
		}
		data, err := json.Marshal(result)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to marshal result: %v", err)), nil, nil
		}

		return jsonResult(data), nil, nil
	})

	// search_enums — browse PostgreSQL enum types.
	type SearchEnumArgs struct {
		Schema string `json:"schema" jsonschema:"PostgreSQL schema name (e.g. 'core', 'inventory'),required"`
		Name   string `json:"name,omitempty" jsonschema:"Enum type name. If omitted, lists all enum types in the schema."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_enums",
		Description: "Browse PostgreSQL enum types. With schema only, lists all enum types. With schema+name, returns enum values with human-friendly labels.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchEnumArgs) (*mcp.CallToolResult, any, error) {
		if args.Schema == "" {
			return errorResult("schema is required"), nil, nil
		}

		if args.Name == "" {
			data, err := c.GetEnumTypes(ctx, args.Schema)
			if err != nil {
				return errorResult("Failed to list enum types: " + err.Error()), nil, nil
			}
			return jsonResult(data), nil, nil
		}

		data, err := c.GetEnumOptions(ctx, args.Schema, args.Name)
		if err != nil {
			return errorResult("Failed to get enum values: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}

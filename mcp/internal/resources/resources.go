// Package resources provides MCP resource handlers that expose Ichor
// configuration data as addressable resources.
package resources

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterResources adds static and template resources to the MCP server.
func RegisterResources(s *mcp.Server, c *client.Client) {
	// config://catalog — static catalog of all configurable surfaces.
	s.AddResource(&mcp.Resource{
		URI:         "config://catalog",
		Name:        "Config Surface Catalog",
		Description: "Complete catalog of all configurable surfaces in Ichor with CRUD endpoints, discovery URLs, and constraints.",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		data, err := c.GetCatalog(ctx)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "config://catalog", MIMEType: "application/json", Text: string(data)},
			},
		}, nil
	})

	// config://action-types — all workflow action types with schemas and ports.
	s.AddResource(&mcp.Resource{
		URI:         "config://action-types",
		Name:        "Workflow Action Types",
		Description: "All 17 workflow action types with JSON config schemas, output ports, categories, and metadata.",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		data, err := c.GetActionTypes(ctx)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "config://action-types", MIMEType: "application/json", Text: string(data)},
			},
		}, nil
	})

	// config://field-types — all form field types with config schemas.
	s.AddResource(&mcp.Resource{
		URI:         "config://field-types",
		Name:        "Form Field Types",
		Description: "All form field types with JSON config schemas describing their configuration options.",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		data, err := c.GetFieldTypes(ctx)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "config://field-types", MIMEType: "application/json", Text: string(data)},
			},
		}, nil
	})

	// config://table-config-schema — JSON schema for table config JSONB.
	s.AddResource(&mcp.Resource{
		URI:         "config://table-config-schema",
		Name:        "Table Config JSON Schema",
		Description: "JSON schema describing the structure of table_configs.config JSONB column.",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		data, err := c.GetTableConfigSchema(ctx)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "config://table-config-schema", MIMEType: "application/json", Text: string(data)},
			},
		}, nil
	})

	// config://layout-schema — JSON schema for page content layout JSONB.
	s.AddResource(&mcp.Resource{
		URI:         "config://layout-schema",
		Name:        "Layout Config JSON Schema",
		Description: "JSON schema describing the structure of page_content.layout JSONB column.",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		data, err := c.GetLayoutSchema(ctx)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "config://layout-schema", MIMEType: "application/json", Text: string(data)},
			},
		}, nil
	})

	// config://db/{schema}/{table} — resource template for table introspection.
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "config://db/{schema}/{table}",
		Name:        "Database Table Schema",
		Description: "Columns and relationships for a database table. Use schema and table parameters.",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Parse URI: config://db/{schema}/{table}
		uri := req.Params.URI
		schema, table, err := parseDBResourceURI(uri)
		if err != nil {
			return nil, err
		}

		columns, colErr := c.GetColumns(ctx, schema, table)
		if colErr != nil {
			return nil, colErr
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: uri, MIMEType: "application/json", Text: string(columns)},
			},
		}, nil
	})

	// config://enums/{schema}/{name} — resource template for enum values.
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "config://enums/{schema}/{name}",
		Name:        "Enum Options",
		Description: "Enum values with human-friendly labels for a specific PostgreSQL enum type.",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI
		schema, name, err := parseEnumResourceURI(uri)
		if err != nil {
			return nil, err
		}

		data, apiErr := c.GetEnumOptions(ctx, schema, name)
		if apiErr != nil {
			return nil, apiErr
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: uri, MIMEType: "application/json", Text: string(data)},
			},
		}, nil
	})
}

// parseDBResourceURI extracts schema and table from "config://db/{schema}/{table}".
func parseDBResourceURI(uri string) (string, string, error) {
	const prefix = "config://db/"
	if len(uri) <= len(prefix) {
		return "", "", mcp.ResourceNotFoundError(uri)
	}
	rest := uri[len(prefix):]
	for i, c := range rest {
		if c == '/' {
			schema := rest[:i]
			table := rest[i+1:]
			if schema == "" || table == "" {
				return "", "", mcp.ResourceNotFoundError(uri)
			}
			return schema, table, nil
		}
	}
	return "", "", mcp.ResourceNotFoundError(uri)
}

// parseEnumResourceURI extracts schema and name from "config://enums/{schema}/{name}".
func parseEnumResourceURI(uri string) (string, string, error) {
	const prefix = "config://enums/"
	if len(uri) <= len(prefix) {
		return "", "", mcp.ResourceNotFoundError(uri)
	}
	rest := uri[len(prefix):]
	for i, c := range rest {
		if c == '/' {
			schema := rest[:i]
			name := rest[i+1:]
			if schema == "" || name == "" {
				return "", "", mcp.ResourceNotFoundError(uri)
			}
			return schema, name, nil
		}
	}
	return "", "", mcp.ResourceNotFoundError(uri)
}

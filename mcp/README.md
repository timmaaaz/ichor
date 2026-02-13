# Ichor MCP Server

Standalone MCP (Model Context Protocol) server that wraps the Ichor REST API, providing LLM agents with tools to discover, read, search, create, and analyze Ichor configurations.

## Overview

The MCP server is a separate Go module (`mcp/`) that connects to a running Ichor API instance via HTTP. It exposes Ichor's capabilities through the MCP protocol, which is supported by Claude Desktop, Ollama, and other MCP-capable clients.

**Module**: `github.com/timmaaaz/ichor/mcp`
**Transport**: stdio (JSON-RPC over stdin/stdout)
**SDK**: `github.com/modelcontextprotocol/go-sdk` v1.3.0

## Running

### With the Dev Cluster (recommended)

Start the KIND cluster first (`make dev-up` or `make dev-bounce`), then:

```bash
make mcp
```

This auto-fetches a token from the auth service and starts the MCP server connected to the Ichor API at `localhost:8080`. Requires `jq` installed.

### Manual

```bash
cd mcp
go run ./cmd/ichor-mcp/ --token $ICHOR_TOKEN
```

The default `--api-url` is `http://localhost:8080` (matching the KIND cluster). Override with `--api-url` for other environments.

Token can also be set via environment variable:

```bash
export ICHOR_TOKEN=your-bearer-token
go run ./cmd/ichor-mcp/
```

### Claude Desktop Configuration

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "ichor": {
      "command": "go",
      "args": ["run", "./cmd/ichor-mcp/", "--token", "YOUR_TOKEN"],
      "cwd": "/path/to/ichor/ichor/mcp"
    }
  }
}
```

## Architecture

```
mcp/
├── cmd/ichor-mcp/
│   ├── main.go              # Entry point: flags, server setup, stdio transport
│   └── main_test.go         # Integration test (InMemoryTransports, full tool/resource/prompt verification)
├── internal/
│   ├── client/
│   │   ├── ichor.go         # HTTP client wrapping all Ichor REST endpoints
│   │   └── ichor_test.go    # Client tests (auth headers, error handling, path correctness)
│   ├── tools/
│   │   ├── discovery.go     # 7 discovery tools + jsonResult/errorResult helpers
│   │   ├── read_ui.go       # 7 UI read tools
│   │   ├── read_workflow.go # 3 workflow read tools
│   │   ├── search.go        # 2 search tools (DB introspection, enums)
│   │   ├── write_workflow.go# 3 workflow write tools (with validation-first pattern)
│   │   ├── write_ui.go      # 8 UI write tools
│   │   ├── validate.go      # 1 validation tool
│   │   └── analysis.go      # 3 analysis/advisory tools
│   ├── resources/
│   │   ├── resources.go     # 5 static resources + 2 resource templates
│   │   └── resources_test.go# URI parsing tests
│   └── prompts/
│       └── prompts.go       # 3 guided prompts (workflow, page, form building)
├── go.mod
└── go.sum
```

### Data Flow

```
LLM Agent ──stdio──► MCP Server ──HTTP──► Ichor REST API ──► PostgreSQL
```

The MCP server is a thin translation layer. Every tool and resource handler calls the Ichor HTTP client, which makes authenticated REST calls to the running Ichor service. No direct database access.

## Complete Tool Inventory (33 tools)

### Discovery (7) — `tools/discovery.go`

| Tool | Args | Description |
|------|------|-------------|
| `discover_config_surfaces` | none | List all configurable surfaces with CRUD endpoints |
| `discover_action_types` | none | All 17 workflow action types with schemas and output ports |
| `discover_field_types` | none | All form field types with config schemas |
| `discover_trigger_types` | none | Workflow trigger types (on_create, on_update, on_delete) |
| `discover_entity_types` | none | Entity types that can trigger workflows |
| `discover_entities` | none | Specific entities registered for triggers |
| `discover_content_types` | none | Valid page content block types |

### UI Read (7) — `tools/read_ui.go`

| Tool | Args | Description |
|------|------|-------------|
| `get_page_config` | `id` or `name` | Page config by ID or name |
| `get_page_content` | `page_config_id` | All content blocks for a page |
| `get_table_config` | `id` or `name` | Table/widget config with full JSONB |
| `get_form_definition` | `id` or `name` | Form with all fields, types, validation |
| `list_pages` | none | All page configs |
| `list_forms` | none | All form definitions |
| `list_table_configs` | none | All table/widget configs |

### Workflow Read (3) — `tools/read_workflow.go`

| Tool | Args | Description |
|------|------|-------------|
| `get_workflow` | `id` | Full workflow: rule + actions + edges merged |
| `list_workflows` | none | All workflow rules |
| `list_action_templates` | none | All reusable action templates |

### Search (2) — `tools/search.go`

| Tool | Args | Description |
|------|------|-------------|
| `search_database_schema` | `schema?`, `table?` | Progressive: no args→schemas, schema→tables, schema+table→columns+relationships |
| `search_enums` | `schema`, `name?` | Schema→enum types, schema+name→values with labels |

### Workflow Write (3) — `tools/write_workflow.go`

| Tool | Args | Description |
|------|------|-------------|
| `validate_workflow` | `workflow` (JSON) | Dry-run validation without saving |
| `create_workflow` | `workflow` (JSON), `validate?` | Create with auto-validation (default: validates first) |
| `update_workflow` | `id`, `workflow` (JSON), `validate?` | Update with auto-validation |

### UI Write (8) — `tools/write_ui.go`

| Tool | Args | Description |
|------|------|-------------|
| `create_page_config` | `config` (JSON) | Create page config |
| `update_page_config` | `id`, `config` (JSON) | Update page config |
| `create_page_content` | `content` (JSON) | Add content block to page |
| `update_page_content` | `id`, `content` (JSON) | Update content block |
| `create_form` | `form` (JSON) | Create form definition |
| `add_form_field` | `field` (JSON) | Add field to form |
| `create_table_config` | `config` (JSON) | Create table/widget config |
| `update_table_config` | `id`, `config` (JSON) | Update table/widget config |

### Validation (1) — `tools/validate.go`

| Tool | Args | Description |
|------|------|-------------|
| `validate_table_config` | `config` (JSON) | Validate table config without saving |

### Analysis (3) — `tools/analysis.go`

| Tool | Args | Description |
|------|------|-------------|
| `analyze_workflow` | `id` | Complexity scoring, branching analysis, suggestions |
| `suggest_templates` | `use_case` (text) | Suggest action templates for a use case |
| `show_cascade` | `entity` | Show which workflows trigger on entity changes |

## Resources (5 static + 2 templates)

### Static Resources — `resources/resources.go`

| URI | Description |
|-----|-------------|
| `config://catalog` | Config surface catalog |
| `config://action-types` | Workflow action types with schemas |
| `config://field-types` | Form field types with schemas |
| `config://table-config-schema` | JSON Schema for table config JSONB |
| `config://layout-schema` | JSON Schema for page layout JSONB |

### Resource Templates

| URI Template | Description |
|-------------|-------------|
| `config://db/{schema}/{table}` | Columns for a database table |
| `config://enums/{schema}/{name}` | Enum values with labels |

## Prompts (3)

| Prompt | Args | Description |
|--------|------|-------------|
| `build-workflow` | `trigger`, `entity` | Guided workflow building with action types and instructions |
| `configure-page` | `entity` | Guided page layout configuration |
| `design-form` | `entity` | Guided form design with field type mapping |

Each prompt pre-loads relevant context from the Ichor API (available action types, field types, content types) so the agent has the information it needs to guide the user.

## Patterns

### Tool Registration

Tools use the Go SDK's generic `mcp.AddTool` which auto-generates JSON schemas from struct tags:

```go
type GetWorkflowArgs struct {
    ID string `json:"id" jsonschema:"UUID of the workflow rule,required"`
}

mcp.AddTool(s, &mcp.Tool{
    Name:        "get_workflow",
    Description: "Get a workflow rule by ID...",
}, func(ctx context.Context, req *mcp.CallToolRequest, args GetWorkflowArgs) (*mcp.CallToolResult, any, error) {
    // args.ID is already parsed and validated by the SDK
    data, err := c.GetWorkflowRule(ctx, args.ID)
    if err != nil {
        return errorResult("Failed: " + err.Error()), nil, nil
    }
    return jsonResult(data), nil, nil
})
```

**Key conventions**:
- Args struct defined inline before each tool registration
- `jsonschema:"description,required"` tag for schema generation
- No-arg tools use `_ struct{}` as the args parameter
- `errorResult()` returns `IsError: true` with text message
- `jsonResult()` wraps `json.RawMessage` as `TextContent`

### Validation-First Writes

Write tools for workflows auto-validate before committing:

```go
// 1. Call dry-run endpoint
valResult, err := c.ValidateWorkflow(ctx, args.Workflow)
// 2. Parse result, abort if invalid
if !result.Valid {
    return jsonResult(validationErrorResponse), nil, nil
}
// 3. Only then call the actual create endpoint
data, err := c.CreateWorkflow(ctx, args.Workflow)
```

### HTTP Client

`internal/client/ichor.go` wraps every Ichor REST endpoint. Pattern:

```go
func (c *Client) GetSomething(ctx context.Context, id string) (json.RawMessage, error) {
    return c.get(ctx, "/v1/path/to/"+id)
}
```

All methods return `json.RawMessage` — the MCP server passes API responses through as-is without deserializing into Go structs.

### Resource URI Parsing

Resource templates parse URIs manually since the SDK doesn't extract template parameters:

```go
func parseDBResourceURI(uri string) (schema, table string, err error) {
    // Parses "config://db/{schema}/{table}"
}
```

## How To Add a New Tool

1. **Decide which file** it belongs in based on category (discovery, read_ui, read_workflow, search, write_ui, write_workflow, validate, analysis). Create a new file if it doesn't fit.

2. **Add the client method** in `internal/client/ichor.go`:
   ```go
   func (c *Client) GetNewThing(ctx context.Context, id string) (json.RawMessage, error) {
       return c.get(ctx, "/v1/path/to/"+id)
   }
   ```

3. **Add the tool** in the appropriate `Register*Tools` function:
   ```go
   type NewToolArgs struct {
       ID string `json:"id" jsonschema:"description,required"`
   }
   mcp.AddTool(s, &mcp.Tool{
       Name:        "new_tool_name",
       Description: "What this tool does.",
   }, func(ctx context.Context, req *mcp.CallToolRequest, args NewToolArgs) (*mcp.CallToolResult, any, error) {
       data, err := c.GetNewThing(ctx, args.ID)
       if err != nil {
           return errorResult("Failed: " + err.Error()), nil, nil
       }
       return jsonResult(data), nil, nil
   })
   ```

4. **Register in main.go** if you created a new `Register*Tools` function (existing functions are already wired).

5. **Update the integration test** in `cmd/ichor-mcp/main_test.go`:
   - Add the tool name to `expectedTools`
   - If new `Register*Tools` function, add the call in the test setup

6. **Run tests**:
   ```bash
   cd mcp && go test ./... -count=1
   ```

## How To Add a New Resource

1. **Add client method** if needed (same as tools).

2. **Add in `resources/resources.go`**:
   ```go
   s.AddResource(&mcp.Resource{
       URI:         "config://new-resource",
       Name:        "Display Name",
       Description: "What this resource provides.",
   }, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
       data, err := c.GetNewThing(ctx)
       if err != nil {
           return nil, err
       }
       return &mcp.ReadResourceResult{
           Contents: []*mcp.ResourceContents{
               {URI: "config://new-resource", MIMEType: "application/json", Text: string(data)},
           },
       }, nil
   })
   ```

3. **Update the integration test** — add URI to `expectedResources`.

## How To Add a New Prompt

1. **Add in `prompts/prompts.go`**:
   ```go
   s.AddPrompt(&mcp.Prompt{
       Name:        "prompt-name",
       Description: "What this prompt guides.",
       Arguments: []*mcp.PromptArgument{
           {Name: "arg", Description: "What this arg is", Required: true},
       },
   }, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
       arg := req.Params.Arguments["arg"]
       // Fetch context from API...
       return &mcp.GetPromptResult{
           Description: "...",
           Messages: []*mcp.PromptMessage{
               {Role: "user", Content: &mcp.TextContent{Text: contextText}},
           },
       }, nil
   })
   ```

2. **Register in main.go** if not already called.

3. **Update the integration test** — add name to `expectedPrompts`.

## Testing

```bash
# All tests
cd mcp && go test ./... -count=1

# Verbose
cd mcp && go test ./... -v -count=1
```

The integration test (`main_test.go`) spins up a mock HTTP server, wires the full MCP server in-process using `mcp.NewInMemoryTransports()`, connects an MCP client, and verifies all tools, resources, templates, and prompts are registered. It also calls `discover_config_surfaces` end-to-end to verify the full request flow.

## Decisions

- **Passthrough JSON**: The client returns `json.RawMessage` everywhere. The MCP server doesn't deserialize API responses into Go structs — it passes them through as text. This avoids maintaining duplicate type definitions and means new API fields are automatically exposed.

- **No SSE transport**: Only stdio is implemented. SSE can be added later if needed for web-based MCP clients, but stdio covers Claude Desktop and Ollama.

- **Validation-first writes**: Write tools call dry-run endpoints before committing. This gives agents a chance to fix errors before making changes. Can be disabled with `validate: false`.

- **Separate Go module**: The MCP server is `mcp/go.mod`, not part of the main Ichor module. This keeps the MCP SDK dependency out of the main binary and allows independent versioning.

# mcp

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared [mcp]=MCP-server
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## ModuleBoundary [mcp]

```
mcp/ (separate Go module)
  ├── main.go              CLI entrypoint (--token, --base-url, --context flags)
  ├── client/client.go     HTTP client wrapper
  ├── tools/               tool registration files (one per domain)
  │   └── register.go      aggregator — calls RegisterXyzTools for each domain
  ├── resources/           5 static + 2 template resources (config:// URIs)
  └── prompts/             3 guided prompts (build-workflow, configure-page, design-form)
```

key facts:
  - Standalone MCP server exposing Ichor capabilities to LLM agents (Claude Desktop, Cursor, etc.)
  - Separate Go module (`mcp/go.mod`) with zero imports of Ichor packages
  - Communicates exclusively via Ichor REST API
  - Transport: stdio (JSON-RPC over stdin/stdout)
  - Auth: Bearer JWT token set once at startup, forwarded on every HTTP call
  - 39 tools available; context modes select subsets
  - Imports only: MCP SDK (`github.com/mark3labs/mcp-go`)
  - New API fields are automatically exposed: all client methods return `json.RawMessage` (passthrough, no deserialization)

---

## Client [mcp]

file: mcp/client/client.go

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client   // timeout: 30s
}
```

key facts:
  - auth: `Authorization: Bearer {token}` on every request
  - Token source: `--token` CLI flag or `ICHOR_TOKEN` environment variable
  - Single identity for all tool calls — no per-request auth switching
  - No caching, no state, no streaming

---

## ToolRegistration [mcp]

```go
// Per-domain registration function
func RegisterXyzTools(s *mcp.Server, c *client.Client) {
    mcp.AddTool(s, &mcp.Tool{
        Name:        toolconstant,
        Description: "...",
    }, func(ctx context.Context, req *mcp.CallToolRequest, args ArgsStruct) (*mcp.CallToolResult, any, error) {
        result, err := c.SomeMethod(ctx, ...)
        if err != nil { return errorResult(err.Error()) }
        return jsonResult(result)
    })
}
```

key facts:
  - Args struct uses `json:"field" jsonschema:"description,required"` tags
  - MCP SDK auto-generates JSON Schema from struct tags
  - No-arg tools use `_ struct{}` parameter

---

## ContextModes [mcp]

Selected via `--context` flag at startup:

| Mode       | Tools | Description                                      |
|------------|-------|--------------------------------------------------|
| `all`      | 39    | all discovery, UI read/write, workflow, search   |
| `workflow` | 16    | 5 discovery + 3 read + 3 write + 2 search + 3 analysis |
| `tables`   | 20    | 2 discovery + 7 read + 2 search + 8 write + 1 validate |

Tool Count by Category (39 total):

| Category           | Count | Examples                                          |
|--------------------|-------|---------------------------------------------------|
| Discovery (workflow) | 5   | DiscoverEntityTypes, DiscoverActionTypes          |
| Discovery (tables)   | 2   | DiscoverConfigSurfaces, DiscoverFieldTypes        |
| Workflow read        | 5   | GetWorkflow, ListWorkflows, ListWorkflowRules     |
| Workflow write       | 3   | ValidateWorkflow, CreateWorkflow, UpdateWorkflow  |
| Workflow analysis    | 3   | AnalyzeWorkflow, SuggestTemplates, ShowCascade    |
| UI read              | 7   | GetPageConfig, GetTableConfig, GetFormDef         |
| UI write             | 8   | CreatePageConfig, UpdatePageConfig, CreateForm    |
| Search               | 2   | SearchDatabaseSchema, SearchEnums                 |
| Validation           | 1   | PreviewWorkflow                                   |
| Page actions         | 3   | ApplyColumnChange, ApplyFilterChange, etc.        |

---

## ResourcesAndPrompts [mcp]

resources (config:// URIs):
  5 static: workflow schema, tables schema, DB introspection, catalog, field types
  2 templates: context-specific schema subsets

prompts (3 guided workflows):
  build-workflow     — pre-loads discovery context for workflow creation
  configure-page     — pre-loads page/table config context
  design-form        — pre-loads form field schema context

---

## AgentInfrastructure [api]

Five dedicated API packages exist in Ichor for agent self-discovery:
  catalogapi          — tool catalog listing
  referenceapi        — enum/entity reference data
  formfieldschemaapi  — field schema definitions
  configschemaapi     — config type schemas
  introspectionapi    — DB schema introspection

key facts:
  - Consumed by MCP resources and by the in-app agent chat (ToolIndex RAG)
  - Validation-first write pattern: create_workflow and update_workflow call dry-run validate endpoint before committing (see docs/arch/workflow-save.md for DAG validation rules)

---

## ⚠ Adding a new MCP tool

  mcp/tools/{domain}_tools.go        (add RegisterXyzTools function with mcp.AddTool call)
  mcp/tools/register.go              (call new RegisterXyzTools in aggregator)
  mcp/client/client.go               (add HTTP method if new Ichor endpoint needed)
  business/sdk/toolcatalog/          (add constant to Ichor toolcatalog for parity
                                      if tool also exists in agent-chat)
  mcp/main.go                        (add tool to appropriate context mode filter(s))

## ⚠ Changing context modes

  mcp/main.go   (context mode filter maps — add/remove tool names per mode)
  README.md     (update tool count table in docs)

## ⚠ Changing MCP client auth

  mcp/client/client.go               (auth header injection)
  mcp/main.go                        (--token flag / ICHOR_TOKEN env var parsing)
  Note: token is a single shared identity; per-user auth would require
  passing credentials per tool call (not currently supported by stdio transport)

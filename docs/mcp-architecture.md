# MCP Architecture Within Ichor

How the MCP server fits into the Ichor system, including the full request flow, authentication chain, API endpoint mapping, and deployment model.

## System Context

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           Developer Machine                              │
│                                                                          │
│  ┌─────────────┐    stdio     ┌──────────────┐   HTTP :8080  ┌────────┐ │
│  │ Claude      │◄────────────►│  ichor-mcp   │─────────────►│ Ichor  │ │
│  │ Desktop /   │  JSON-RPC    │  (make mcp)  │  Bearer Auth  │ API    │ │
│  │ Ollama /    │              └──────────────┘               │ (K8s)  │ │
│  │ Claude Code │                                             │        │ │
│  └─────────────┘                                             │        │ │
│                                                              │        │ │
│  ┌─────────────┐              ┌──────────────┐               │        │ │
│  │ Frontend    │─────────────►│              │               │        │ │
│  │ (Browser)   │  HTTP :8080  │              │               │        │ │
│  └─────────────┘              │              │      SQL      │        │ │
│                               │              │◄─────────────►│ PG     │ │
│  ┌─────────────┐              │              │               │        │ │
│  │ Workflow    │◄────────────►│              │               └────────┘ │
│  │ Worker      │   Temporal   │              │                          │
│  └─────────────┘              └──────────────┘                          │
│                               (KIND cluster)                            │
└──────────────────────────────────────────────────────────────────────────┘
```

The MCP server is a **sidecar process** — it runs alongside the Ichor API, not inside it. It's a separate Go binary (`mcp/cmd/ichor-mcp/`) with its own Go module (`mcp/go.mod`) and its own dependency tree. It has no direct database access and no shared memory with the Ichor process.

The only link between them is HTTP: the MCP server makes authenticated REST calls to the same API that the frontend and other clients use.

## Request Flow

When an LLM agent calls an MCP tool, the request passes through 5 layers:

```
1. Agent         →  MCP protocol (JSON-RPC over stdio)
2. MCP Server    →  Parses tool call, extracts typed args
3. HTTP Client   →  Makes authenticated REST call to Ichor API
4. Ichor API     →  Standard route → middleware → handler → app → business → db
5. Response      →  JSON flows back through each layer unchanged
```

### Concrete Example: `get_workflow`

```
Agent calls: tools/call { name: "get_workflow", arguments: { id: "abc-123" } }
                                    │
                                    ▼
        MCP SDK deserializes args into GetWorkflowArgs{ID: "abc-123"}
                                    │
                                    ▼
        Handler calls 3 HTTP client methods:
          c.GetWorkflowRule(ctx, "abc-123")       →  GET /v1/workflow/rules/abc-123
          c.GetWorkflowRuleActions(ctx, "abc-123") →  GET /v1/workflow/rules/abc-123/actions
          c.GetWorkflowRuleEdges(ctx, "abc-123")   →  GET /v1/workflow/rules/abc-123/edges
                                    │
                                    ▼
        Ichor API processes each request through:
          authen middleware (JWT validation)
          authorize middleware (role + permissions check)
          ruleapi/edgeapi handler
          workflowapp (app layer)
          workflowbus (business layer)
          workflowdb (PostgreSQL)
                                    │
                                    ▼
        MCP handler merges 3 JSON responses into:
          { "rule": {...}, "actions": [...], "edges": [...] }
                                    │
                                    ▼
        Returns as MCP TextContent (raw JSON string)
```

### Concrete Example: `create_workflow` (with validation)

```
Agent calls: tools/call { name: "create_workflow", arguments: { workflow: {...} } }
                                    │
                                    ▼
        1. Validation-first: POST /v1/workflow/rules/full?dry_run=true
           Ichor runs full graph validation (DAG check, action types, output ports, edge types)
           Returns: { "valid": true/false, "errors": [...] }
                                    │
                                    ▼
        2. If valid=false → return validation errors to agent (no write happens)
           If valid=true  → POST /v1/workflow/rules/full (actual create)
                                    │
                                    ▼
        Ichor API processes through workflowsaveapi → workflowsaveapp
          (single transaction: creates rule + actions + edges atomically)
                                    │
                                    ▼
        Returns created workflow to agent
```

## Authentication Chain

```
┌──────────┐         ┌──────────┐         ┌──────────────┐
│ Agent    │  token   │ MCP      │  Bearer  │ Ichor API    │
│ config   │────────►│ Server   │─────────►│ middleware   │
│ (flag/   │         │ (stored  │  header   │ (JWT verify) │
│  env)    │         │  once)   │          │              │
└──────────┘         └──────────┘         └──────────────┘
```

1. The MCP server receives a bearer token at startup (`--token` flag or `ICHOR_TOKEN` env var)
2. Every HTTP request includes `Authorization: Bearer {token}` header
3. Ichor's `mid.Authenticate()` middleware validates the JWT
4. Ichor's `mid.Authorize()` middleware checks role permissions per endpoint

The MCP server runs with a **single identity** — whatever user the token belongs to. It doesn't have per-agent or per-request auth. The token determines what the agent can see and do.

**Permission requirements by endpoint category**:

| MCP Tool Category | Ichor Auth Requirement |
|-------------------|----------------------|
| Discovery tools | Admin role (referenceapi, introspectionapi) |
| UI read tools | Depends on endpoint (page configs: any, table configs: any) |
| Workflow read tools | Admin role |
| Search tools | Admin role (introspection) |
| Write tools | Same as underlying CRUD endpoint |
| Enum options | Any authenticated user |

## API Endpoint Mapping

Every MCP tool and resource maps to one or more Ichor REST endpoints. The MCP server adds no business logic — it's a translation layer.

### MCP Tool → Ichor API Endpoint

| MCP Tool | HTTP Method | Ichor Endpoint | Ichor Package |
|----------|------------|----------------|---------------|
| **Discovery** | | | |
| `discover_config_surfaces` | GET | `/v1/agent/catalog` | `catalogapi` |
| `discover_action_types` | GET | `/v1/workflow/action-types` | `referenceapi` |
| `discover_field_types` | GET | `/v1/config/form-field-types` | `formfieldschemaapi` |
| `discover_trigger_types` | GET | `/v1/workflow/trigger-types` | `referenceapi` |
| `discover_entity_types` | GET | `/v1/workflow/entity-types` | `referenceapi` |
| `discover_entities` | GET | `/v1/workflow/entities` | `referenceapi` |
| `discover_content_types` | GET | `/v1/config/schemas/content-types` | `configschemaapi` |
| **UI Read** | | | |
| `get_page_config` | GET | `/v1/config/page-configs/id/{id}` or `/name/{name}` | `pageconfigapi` |
| `get_page_content` | GET | `/v1/config/page-configs/content/{id}` | `pagecontentapi` |
| `get_table_config` | GET | `/v1/data/id/{id}` or `/name/{name}` | `dataapi` |
| `get_form_definition` | GET | `/v1/config/forms/{id}/full` or `/name/{name}/full` | `formapi` |
| `list_pages` | GET | `/v1/config/page-configs/all` | `pageconfigapi` |
| `list_forms` | GET | `/v1/config/forms` | `formapi` |
| `list_table_configs` | GET | `/v1/data/configs/all` | `dataapi` |
| **Workflow Read** | | | |
| `get_workflow` | GET | `/v1/workflow/rules/{id}` + `/actions` + `/edges` | `ruleapi`, `edgeapi` |
| `list_workflows` | GET | `/v1/workflow/rules` | `ruleapi` |
| `list_action_templates` | GET | `/v1/workflow/templates` | `referenceapi` |
| **Search** | | | |
| `search_database_schema` | GET | `/v1/introspection/schemas` → `/tables` → `/columns` + `/relationships` | `introspectionapi` |
| `search_enums` | GET | `/v1/introspection/enums/{schema}` or `/v1/config/enums/{schema}/{name}/options` | `introspectionapi` |
| **Workflow Write** | | | |
| `validate_workflow` | POST | `/v1/workflow/rules/full?dry_run=true` | `workflowsaveapi` |
| `create_workflow` | POST | `/v1/workflow/rules/full` (+ dry_run) | `workflowsaveapi` |
| `update_workflow` | PUT | `/v1/workflow/rules/{id}/full` (+ dry_run) | `workflowsaveapi` |
| **UI Write** | | | |
| `create_page_config` | POST | `/v1/config/page-configs` | `pageconfigapi` |
| `update_page_config` | PUT | `/v1/config/page-configs/id/{id}` | `pageconfigapi` |
| `create_page_content` | POST | `/v1/config/page-content` | `pagecontentapi` |
| `update_page_content` | PUT | `/v1/config/page-content/{id}` | `pagecontentapi` |
| `create_form` | POST | `/v1/config/forms` | `formapi` |
| `add_form_field` | POST | `/v1/config/form-fields` | `formfieldapi` |
| `create_table_config` | POST | `/v1/data` | `dataapi` |
| `update_table_config` | PUT | `/v1/data/{id}` | `dataapi` |
| **Validation** | | | |
| `validate_table_config` | POST | `/v1/data/validate` | `dataapi` |
| **Analysis** | | | |
| `analyze_workflow` | GET | `/v1/workflow/rules/{id}` + `/actions` + `/edges` | `ruleapi`, `edgeapi` |
| `suggest_templates` | GET | `/v1/workflow/templates/active` + `/action-types` | `referenceapi` |
| `show_cascade` | GET | `/v1/workflow/rules` | `ruleapi` |

### MCP Resource → Ichor API Endpoint

| MCP Resource | Ichor Endpoint | Ichor Package |
|-------------|----------------|---------------|
| `config://catalog` | `GET /v1/agent/catalog` | `catalogapi` |
| `config://action-types` | `GET /v1/workflow/action-types` | `referenceapi` |
| `config://field-types` | `GET /v1/config/form-field-types` | `formfieldschemaapi` |
| `config://table-config-schema` | `GET /v1/config/schemas/table-config` | `configschemaapi` |
| `config://layout-schema` | `GET /v1/config/schemas/layout` | `configschemaapi` |
| `config://db/{schema}/{table}` | `GET /v1/introspection/tables/{schema}/{table}/columns` | `introspectionapi` |
| `config://enums/{schema}/{name}` | `GET /v1/config/enums/{schema}/{name}/options` | `introspectionapi` |

## The Agent Infrastructure Layer

Five API packages exist specifically to make Ichor self-describing for agents. These were built before the MCP server and work independently of it — the frontend or any HTTP client can use them too.

```
┌─────────────────────────────────────────────────────────────────┐
│                    Ichor API Service                             │
│                                                                  │
│  ┌─────────────────── Agent Infrastructure ──────────────────┐  │
│  │                                                            │  │
│  │  catalogapi          "What can I configure?"               │  │
│  │  ├── GET /v1/agent/catalog                                 │  │
│  │  └── 12 config surfaces with endpoints + constraints       │  │
│  │                                                            │  │
│  │  referenceapi        "What workflow building blocks exist?" │  │
│  │  ├── GET /v1/workflow/action-types (17 types + schemas)    │  │
│  │  ├── GET /v1/workflow/trigger-types                        │  │
│  │  └── GET /v1/workflow/entity-types                         │  │
│  │                                                            │  │
│  │  formfieldschemaapi  "What form fields can I use?"         │  │
│  │  ├── GET /v1/config/form-field-types (14 types + schemas)  │  │
│  │  └── GET /v1/config/form-field-types/{type}/schema         │  │
│  │                                                            │  │
│  │  configschemaapi     "What shape is the JSONB config?"     │  │
│  │  ├── GET /v1/config/schemas/table-config                   │  │
│  │  ├── GET /v1/config/schemas/layout                         │  │
│  │  └── GET /v1/config/schemas/content-types                  │  │
│  │                                                            │  │
│  │  introspectionapi    "What's in the database?"             │  │
│  │  ├── GET /v1/introspection/schemas → tables → columns      │  │
│  │  └── GET /v1/introspection/enums                           │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─────────────────── Domain CRUD APIs ──────────────────────┐  │
│  │  pageconfigapi, pagecontentapi, formapi, formfieldapi,    │  │
│  │  dataapi, ruleapi, edgeapi, workflowsaveapi, alertapi,    │  │
│  │  actionapi, templateapi, ...                               │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─────────────────── Dry-Run Support ───────────────────────┐  │
│  │  workflowsaveapi: POST /v1/workflow/rules/full?dry_run=true│  │
│  │  dataapi:          POST /v1/data/validate                  │  │
│  └────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

The agent infrastructure answers the questions an agent needs before it can make changes:
1. **What exists?** → catalog, list endpoints
2. **What are the rules?** → schemas, validation, constraints
3. **Will this work?** → dry-run endpoints
4. **Do it** → CRUD endpoints

## Typical Agent Workflow

A well-behaved agent follows a discover → read → validate → write sequence:

```
1. discover_config_surfaces          "What can I configure?"
        │
2. discover_action_types             "What actions are available?"
   discover_field_types              "What field types exist?"
        │
3. search_database_schema            "What does the orders table look like?"
   get_workflow / get_form            "What already exists?"
        │
4. validate_workflow                  "Will this graph be valid?"
        │
5. create_workflow / create_form      "Save it"
```

The MCP prompts (`build-workflow`, `configure-page`, `design-form`) encode this sequence — they pre-load discovery data into the prompt context so the agent starts at step 3 instead of step 1.

## Deployment Model

### Development: Local Sidecar via `make mcp`

```
Developer machine:
  ├── KIND cluster (make dev-up)
  │   ├── Ichor API pod ──► port-forward ──► localhost:8080
  │   ├── Auth Service pod ──► port-forward ──► localhost:6000
  │   ├── Workflow Worker pod
  │   └── PostgreSQL pod
  │
  ├── ichor-mcp (local binary, stdio, launched by make mcp)
  │   └── connects to Ichor API at localhost:8080
  │
  └── Claude Desktop / Ollama / Claude Code
      └── connects to ichor-mcp via stdio
```

The standard development workflow:

```bash
# 1. Start the cluster (if not already running)
make dev-up          # or make dev-bounce for a fresh start

# 2. Start the MCP server
make mcp             # auto-fetches token from auth service, starts ichor-mcp
```

`make mcp` does two things:
1. Fetches a JWT from the auth service at `localhost:6000` (same as `make token`, using the default admin user)
2. Runs `go run ./cmd/ichor-mcp/` with that token, connected to `localhost:8080`

The MCP server is a **local process** — same pattern as `make run`, `make token`, and `make users`. It's not part of `make dev-bounce` or `make dev-update-apply` because it doesn't run in the cluster.

### Services in the Ichor Stack

| Service | Container | Port | Protocol | In K8s? |
|---------|-----------|------|----------|---------|
| Ichor API | `dockerfile.ichor` | 8080 | HTTP/REST | Yes |
| Auth Service | `dockerfile.auth` | 6000 | HTTP/REST | Yes |
| Metrics | `dockerfile.metrics` | — | Prometheus | Yes |
| Workflow Worker | `dockerfile.workflow-worker` | 4001 (health) | Temporal | Yes |
| **MCP Server** | **none** | **stdio** | **JSON-RPC** | **No — local** |

The MCP server is intentionally not containerized because:
- stdio transport requires the process to run on the same machine as the MCP client
- MCP clients (Claude Desktop, Ollama) expect to launch the server as a subprocess
- There's no network listener — no port to expose

### Why Not in the Cluster?

The MCP protocol uses stdio (stdin/stdout) as its transport. The MCP client (Claude Desktop, Ollama, Claude Code) launches the MCP server as a child process and communicates over pipes. This is fundamentally incompatible with running in a pod — there's no network port to expose, no service to route to, no ingress to configure.

This is the same reason you don't put `make token` or `make pgcli` in the cluster — they're developer tools that run locally and talk to the cluster over port-forwarded connections.

### Future: SSE Transport

If an SSE (Server-Sent Events) transport is added, the MCP server could be containerized and run as a service, allowing remote MCP clients to connect over HTTP. The SDK supports this but it's not implemented yet.

## Module Boundary

```
ichor/                          ← Main Go module (github.com/timmaaaz/ichor)
├── api/                        ← API layer (Ichor service)
├── app/                        ← App layer
├── business/                   ← Business layer
├── foundation/                 ← Foundation
│
└── mcp/                        ← SEPARATE Go module (github.com/timmaaaz/ichor/mcp)
    ├── go.mod                  ← Own dependency tree
    ├── cmd/ichor-mcp/          ← Own binary
    └── internal/               ← Own packages
```

The MCP module (`mcp/go.mod`) does **not** import any Ichor packages. The only dependency is the MCP SDK (`github.com/modelcontextprotocol/go-sdk`). This means:

- Changes to Ichor business logic don't require rebuilding the MCP server
- The MCP SDK dependency doesn't pollute the main Ichor binary
- The two can be versioned independently
- The MCP server can be tested with a mock HTTP server (no Ichor code needed)

The trade-off: if Ichor API response shapes change, the MCP server won't know until runtime (since it passes through `json.RawMessage` without deserializing). This is intentional — it avoids maintaining duplicate Go type definitions.

## What the MCP Server Does NOT Do

- **No direct database access** — always goes through the Ichor API
- **No business logic** — validation, authorization, and constraints are enforced by Ichor
- **No caching** — every tool call makes fresh HTTP requests
- **No state** — no sessions, no conversation memory, no request correlation
- **No streaming** — all responses are complete JSON (no SSE or WebSocket)
- **No multi-tenancy** — single token, single identity for all tool calls

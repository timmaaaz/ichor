# MCP & Agent Infrastructure — Frontend Integration Guide

Everything the frontend needs to understand and work with the Ichor MCP server and agent-friendly API infrastructure.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [How It Works](#how-it-works)
- [Running the MCP Server](#running-the-mcp-server)
- [Authentication](#authentication)
- [API Endpoint Reference](#api-endpoint-reference)
  - [Config Surface Catalog](#1-config-surface-catalog)
  - [Table Config & Layout JSON Schemas](#2-table-config--layout-json-schemas)
  - [Form Field Type Discovery](#3-form-field-type-discovery)
  - [Workflow Action Type Discovery](#4-workflow-action-type-discovery)
  - [Database Introspection](#5-database-introspection)
  - [Dry-Run Validation](#6-dry-run-validation)
- [MCP Tools (33 total)](#mcp-tools-33-total)
  - [Discovery Tools](#discovery-tools-7)
  - [UI Read Tools](#ui-read-tools-7)
  - [Workflow Read Tools](#workflow-read-tools-3)
  - [Search Tools](#search-tools-2)
  - [Workflow Write Tools](#workflow-write-tools-3)
  - [UI Write Tools](#ui-write-tools-8)
  - [Validation Tools](#validation-tools-1)
  - [Analysis Tools](#analysis-tools-3)
- [MCP Resources (7 total)](#mcp-resources-7-total)
- [MCP Prompts (3 total)](#mcp-prompts-3-total)
- [Response Shapes](#response-shapes)
- [Typical Agent Workflow](#typical-agent-workflow)
- [Config Surface Catalog (Full List)](#config-surface-catalog-full-list)
- [Workflow Action Types (Full List)](#workflow-action-types-full-list)
- [Form Field Types (Full List)](#form-field-types-full-list)

---

## Architecture Overview

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
│  ┌─────────────┐                                             │        │ │
│  │ Frontend    │─────── HTTP :8080 ─────────────────────────►│        │ │
│  │ (Browser)   │                                             │        │ │
│  └─────────────┘                                             └────────┘ │
└──────────────────────────────────────────────────────────────────────────┘
```

**Key points:**

- The MCP server is a **separate Go binary** (`mcp/cmd/ichor-mcp/`) with its own Go module (`mcp/go.mod`)
- It's a **thin translation layer** — it makes authenticated HTTP calls to the same Ichor REST API the frontend uses
- No direct database access, no shared memory with the Ichor process
- Transport: **stdio** (JSON-RPC over stdin/stdout) — launched as a subprocess by the MCP client
- The "agent infrastructure" REST endpoints exist independently of the MCP server — **the frontend can call them directly too**

---

## How It Works

When an LLM agent calls an MCP tool, the request flows through:

```
1. Agent         →  MCP protocol (JSON-RPC over stdio)
2. MCP Server    →  Parses tool call, extracts typed args
3. HTTP Client   →  Makes authenticated REST call to Ichor API
4. Ichor API     →  route → middleware → handler → app → business → db
5. Response      →  JSON flows back through each layer unchanged
```

The MCP server passes API responses through as raw JSON (`json.RawMessage`) — it does not deserialize into Go structs. This means new API fields are automatically exposed without MCP server changes.

---

## Running the MCP Server

### With the Dev Cluster (recommended)

```bash
# Start KIND cluster first
make dev-up

# Start MCP server (auto-fetches token)
make mcp
```

### Manual

```bash
cd mcp
go run ./cmd/ichor-mcp/ --token $ICHOR_TOKEN
```

### Configuration

| Setting | Flag | Env Var | Default |
|---------|------|---------|---------|
| Bearer token | `--token` | `ICHOR_TOKEN` | (required) |
| API base URL | `--api-url` | — | `http://localhost:8080` |

### Claude Desktop Integration

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

---

## Authentication

All requests (both MCP and direct API) use JWT Bearer tokens:

```
Authorization: Bearer {token}
Content-Type: application/json
```

The MCP server receives its token at startup and sends it on every request. It runs with a **single identity** — whatever user the token belongs to.

**Permission requirements by endpoint category:**

| Endpoint Category | Auth Requirement |
|-------------------|-----------------|
| Discovery endpoints | Admin role |
| UI read endpoints | Varies (most: any authenticated) |
| Workflow read endpoints | Admin role |
| Introspection (search) | Admin role |
| Write endpoints | Same as underlying CRUD endpoint |
| Enum options | Any authenticated user |

---

## API Endpoint Reference

These are the REST endpoints that the agent infrastructure exposes. The frontend can call these directly — they don't require the MCP server.

### 1. Config Surface Catalog

**`GET /v1/agent/catalog`**

Returns a list of all 12 configurable surfaces in the system with their CRUD endpoints, discovery URLs, and constraints.

**Response shape:**
```json
[
  {
    "name": "Page Configs",
    "description": "Page-level configuration containers...",
    "category": "ui",
    "endpoints": {
      "list": "GET /v1/config/page-configs/all",
      "get": "GET /v1/config/page-configs/id/{config_id}",
      "create": "POST /v1/config/page-configs",
      "update": "PUT /v1/config/page-configs/id/{config_id}",
      "delete": "DELETE /v1/config/page-configs/id/{config_id}"
    },
    "discovery_url": "GET /v1/config/page-configs/all",
    "constraints": [
      "name must be unique",
      "supports import/export via POST /v1/config/page-configs/export and /import"
    ]
  }
]
```

### 2. Table Config & Layout JSON Schemas

| Endpoint | Returns |
|----------|---------|
| `GET /v1/config/schemas/table-config` | JSON Schema for `config.table_configs.config` JSONB column |
| `GET /v1/config/schemas/layout` | JSON Schema for `config.page_content.layout` JSONB column |
| `GET /v1/config/schemas/content-types` | Valid content types with descriptions and requirements |

**Table config schema covers:**
- `data_source[]` — query definitions (table, schema, columns, joins, filters, sort, metrics, group_by)
- `visual_settings.columns` — per-column config (type, header, width, format, editability, links, lookups)
- `visual_settings.conditional_formatting[]` — rule-based styling
- `visual_settings.row_actions[]` / `table_actions[]` — UI actions
- `visual_settings.pagination` — page sizes and defaults
- `permissions` — role-based access
- Chart types (when `visualization: "chart"`)

**Layout schema covers:**
- `colSpan` / `gridCols` — `ResponsiveValue` with breakpoints (default, sm, md, lg, xl, 2xl)
- `containerType` — grid-12, flex, stack, tab, accordion, section, grid
- `gap` — Tailwind CSS gap classes

### 3. Form Field Type Discovery

| Endpoint | Returns |
|----------|---------|
| `GET /v1/config/form-field-types` | All field types with names, descriptions, and config schemas |
| `GET /v1/config/form-field-types/{type}/schema` | Single field type's config schema |

**Response shape:**
```json
[
  {
    "type": "dropdown",
    "name": "Dropdown",
    "description": "Select from entity records...",
    "config_schema": {
      "type": "object",
      "properties": {
        "entity": { "type": "string" },
        "label_column": { "type": "string" },
        "value_column": { "type": "string" }
      },
      "required": ["entity", "label_column", "value_column"]
    }
  }
]
```

### 4. Workflow Action Type Discovery

| Endpoint | Returns |
|----------|---------|
| `GET /v1/workflow/action-types` | All 17 action types with schemas and output ports |
| `GET /v1/workflow/action-types/{type}/schema` | Single action type's config schema |
| `GET /v1/workflow/trigger-types` | Available trigger types |
| `GET /v1/workflow/entity-types` | Entity types for triggers |
| `GET /v1/workflow/entities` | Registered entities |
| `GET /v1/workflow/templates` | All action templates |
| `GET /v1/workflow/templates/active` | Active templates only |

**Action type response shape:**
```json
[
  {
    "type": "evaluate_condition",
    "name": "Evaluate Condition",
    "description": "Evaluates a boolean expression to branch workflow execution",
    "category": "control",
    "supports_manual_execution": false,
    "is_async": false,
    "config_schema": { ... },
    "output_ports": [
      { "name": "true", "description": "Condition evaluated to true" },
      { "name": "false", "description": "Condition evaluated to false" }
    ]
  }
]
```

**Trigger type response shape:**
```json
[
  {
    "id": "uuid",
    "name": "on_create",
    "description": "Fires when an entity is created",
    "is_active": true
  }
]
```

**Entity type response shape:**
```json
[
  {
    "id": "uuid",
    "name": "standard",
    "description": "Standard entity type",
    "is_active": true
  }
]
```

**Entity response shape:**
```json
[
  {
    "id": "uuid",
    "name": "orders",
    "entity_type_id": "uuid",
    "schema_name": "sales",
    "is_active": true
  }
]
```

**Action template response shape:**
```json
[
  {
    "id": "uuid",
    "name": "Low Stock Alert",
    "description": "Creates alert when inventory falls below threshold",
    "actionType": "create_alert",
    "icon": "alert-triangle",
    "defaultConfig": { ... },
    "createdDate": "2026-01-15T10:00:00Z",
    "createdBy": "uuid",
    "isActive": true,
    "deactivatedBy": "uuid"
  }
]
```

### 5. Database Introspection

Progressive browsing of the PostgreSQL schema. All read-only.

| Endpoint | Returns |
|----------|---------|
| `GET /v1/introspection/schemas` | All PostgreSQL schemas |
| `GET /v1/introspection/schemas/{schema}/tables` | Tables in a schema |
| `GET /v1/introspection/tables/{schema}/{table}/columns` | Column definitions |
| `GET /v1/introspection/tables/{schema}/{table}/relationships` | Foreign key relationships |
| `GET /v1/introspection/tables/{schema}/{table}/referencing-tables` | Tables that reference this one |
| `GET /v1/introspection/enums/{schema}` | Enum types in a schema |
| `GET /v1/introspection/enums/{schema}/{name}` | Values of a specific enum |
| `GET /v1/config/enums/{schema}/{name}/options` | Enum values merged with human-friendly labels |

**Schema response:**
```json
[{ "name": "core" }, { "name": "sales" }, { "name": "inventory" }]
```

**Table response:**
```json
[{ "schema": "sales", "name": "orders", "row_count_estimate": 1500 }]
```

**Column response:**
```json
[
  {
    "name": "id",
    "data_type": "uuid",
    "is_nullable": false,
    "is_primary_key": true,
    "default_value": "gen_random_uuid()",
    "is_foreign_key": false
  },
  {
    "name": "customer_id",
    "data_type": "uuid",
    "is_nullable": false,
    "is_primary_key": false,
    "default_value": "",
    "is_foreign_key": true,
    "referenced_schema": "sales",
    "referenced_table": "customers",
    "referenced_column": "id"
  }
]
```

**Relationship response:**
```json
[
  {
    "foreign_key_name": "fk_orders_customer_id",
    "column_name": "customer_id",
    "referenced_schema": "sales",
    "referenced_table": "customers",
    "referenced_column": "id",
    "relationship_type": "many-to-one"
  }
]
```

**Referencing table response:**
```json
[
  {
    "schema": "sales",
    "table": "order_line_items",
    "foreign_key_column": "order_id",
    "constraint_name": "fk_line_items_order_id"
  }
]
```

**Enum type response:**
```json
[
  { "name": "order_status", "schema": "sales", "values": ["draft", "pending", "approved", "shipped"] }
]
```

**Enum value response:**
```json
[{ "value": "draft", "sort_order": 1 }, { "value": "pending", "sort_order": 2 }]
```

**Enum option response (with labels):**
```json
[
  { "value": "draft", "label": "Draft", "sort_order": 1 },
  { "value": "pending", "label": "Pending Review", "sort_order": 2 }
]
```

### 6. Dry-Run Validation

Validate without saving — same request body as regular save, add `?dry_run=true`.

| Endpoint | Purpose |
|----------|---------|
| `POST /v1/workflow/rules/full?dry_run=true` | Validate new workflow |
| `PUT /v1/workflow/rules/{id}/full?dry_run=true` | Validate workflow update |
| `POST /v1/data/validate` | Validate table config JSONB |

**Workflow validation response:**
```json
{
  "valid": true,
  "errors": [],
  "action_count": 5,
  "edge_count": 6
}
```

**Failure example:**
```json
{
  "valid": false,
  "errors": ["graph contains a cycle", "action 'check_stock' has no outgoing edges"],
  "action_count": 5,
  "edge_count": 6
}
```

---

## MCP Tools (33 total)

### Discovery Tools (7)

| Tool | Input | REST Endpoint | Description |
|------|-------|---------------|-------------|
| `discover_config_surfaces` | (none) | `GET /v1/agent/catalog` | List all 12 configurable surfaces with CRUD endpoints |
| `discover_action_types` | (none) | `GET /v1/workflow/action-types` | All 17 workflow action types with JSON schemas + output ports |
| `discover_field_types` | (none) | `GET /v1/config/form-field-types` | All form field types with config schemas |
| `discover_trigger_types` | (none) | `GET /v1/workflow/trigger-types` | Workflow trigger types (on_create, on_update, on_delete) |
| `discover_entity_types` | (none) | `GET /v1/workflow/entity-types` | Entity types that can trigger workflows |
| `discover_entities` | (none) | `GET /v1/workflow/entities` | Specific entities registered for triggers |
| `discover_content_types` | (none) | `GET /v1/config/schemas/content-types` | Valid page content block types (table, form, chart, tabs, container, text) |

### UI Read Tools (7)

| Tool | Input | REST Endpoint | Description |
|------|-------|---------------|-------------|
| `get_page_config` | `id` (optional), `name` (optional) | `GET /v1/config/page-configs/id/{id}` or `/name/{name}` | Get page config by ID or name |
| `get_page_content` | `page_config_id` (required UUID) | `GET /v1/config/page-configs/content/{id}` | Get all content blocks for a page |
| `get_table_config` | `id` (optional), `name` (optional) | `GET /v1/data/id/{id}` or `/name/{name}` | Get table/widget config with full JSONB |
| `get_form_definition` | `id` (optional), `name` (optional) | `GET /v1/config/forms/{id}/full` or `/name/{name}/full` | Get form with all fields, types, validation |
| `list_pages` | (none) | `GET /v1/config/page-configs/all` | List all page configs |
| `list_forms` | (none) | `GET /v1/config/forms` | List all form definitions |
| `list_table_configs` | (none) | `GET /v1/data/configs/all` | List all table/widget configs |

### Workflow Read Tools (3)

| Tool | Input | REST Endpoint | Description |
|------|-------|---------------|-------------|
| `get_workflow` | `id` (required UUID) | 3 calls: `GET /v1/workflow/rules/{id}` + `/actions` + `/edges` | Get workflow with full action graph. Returns merged JSON: `{ rule, actions, edges }` |
| `list_workflows` | (none) | `GET /v1/workflow/rules` | List all workflow automation rules |
| `list_action_templates` | (none) | `GET /v1/workflow/templates` | List all reusable action templates |

### Search Tools (2)

| Tool | Input | REST Endpoint | Behavior |
|------|-------|---------------|----------|
| `search_database_schema` | `schema` (optional), `table` (optional) | Progressive: no args → schemas, schema → tables, schema+table → columns + relationships | Database browsing |
| `search_enums` | `schema` (required), `name` (optional) | Schema only → enum types, schema+name → values with labels | Enum browsing |

### Workflow Write Tools (3)

| Tool | Input | REST Endpoint | Description |
|------|-------|---------------|-------------|
| `validate_workflow` | `workflow` (required JSON) | `POST /v1/workflow/rules/full?dry_run=true` | Dry-run validation, returns `{ valid, errors, action_count, edge_count }` |
| `create_workflow` | `workflow` (required JSON), `validate` (optional bool, default: true) | Validates first, then `POST /v1/workflow/rules/full` | Create with auto-validation |
| `update_workflow` | `id` (required UUID), `workflow` (required JSON), `validate` (optional bool, default: true) | Validates first, then `PUT /v1/workflow/rules/{id}/full` | Update with auto-validation |

### UI Write Tools (8)

| Tool | Input | REST Endpoint | Description |
|------|-------|---------------|-------------|
| `create_page_config` | `config` (required JSON) | `POST /v1/config/page-configs` | Create page config |
| `update_page_config` | `id` (required UUID), `config` (required JSON) | `PUT /v1/config/page-configs/id/{id}` | Update page config |
| `create_page_content` | `content` (required JSON) | `POST /v1/config/page-content` | Add content block to page |
| `update_page_content` | `id` (required UUID), `content` (required JSON) | `PUT /v1/config/page-content/{id}` | Update content block |
| `create_form` | `form` (required JSON) | `POST /v1/config/forms` | Create form definition |
| `add_form_field` | `field` (required JSON) | `POST /v1/config/form-fields` | Add field to form |
| `create_table_config` | `config` (required JSON) | `POST /v1/data` | Create table/widget config |
| `update_table_config` | `id` (required UUID), `config` (required JSON) | `PUT /v1/data/{id}` | Update table/widget config |

### Validation Tools (1)

| Tool | Input | REST Endpoint | Description |
|------|-------|---------------|-------------|
| `validate_table_config` | `config` (required JSON) | `POST /v1/data/validate` | Validate table config JSONB without saving |

### Analysis Tools (3)

| Tool | Input | REST Endpoint | Returns |
|------|-------|---------------|---------|
| `analyze_workflow` | `id` (required UUID) | `GET /v1/workflow/rules/{id}` + actions + edges | `{ rule, action_count, edge_count, action_type_counts, max_branching, suggestions }` |
| `suggest_templates` | `use_case` (required text) | `GET /v1/workflow/templates/active` + action-types | `{ use_case, available_templates, available_action_types, guidance }` |
| `show_cascade` | `entity` (required text) | `GET /v1/workflow/rules` | `{ entity, triggered_rules, rule_count, message }` |

---

## MCP Resources (7 total)

### Static Resources (5)

| URI | Name | MIME Type | REST Endpoint |
|-----|------|-----------|---------------|
| `config://catalog` | Config Surface Catalog | application/json | `GET /v1/agent/catalog` |
| `config://action-types` | Workflow Action Types | application/json | `GET /v1/workflow/action-types` |
| `config://field-types` | Form Field Types | application/json | `GET /v1/config/form-field-types` |
| `config://table-config-schema` | Table Config JSON Schema | application/json | `GET /v1/config/schemas/table-config` |
| `config://layout-schema` | Layout Config JSON Schema | application/json | `GET /v1/config/schemas/layout` |

### Resource Templates (2)

| URI Template | Name | Parameters | REST Endpoint |
|--------------|------|------------|---------------|
| `config://db/{schema}/{table}` | Database Table Schema | schema, table | `GET /v1/introspection/tables/{schema}/{table}/columns` |
| `config://enums/{schema}/{name}` | Enum Options | schema, name | `GET /v1/config/enums/{schema}/{name}/options` |

---

## MCP Prompts (3 total)

### `build-workflow`

Guide the user through building a workflow automation rule.

**Arguments:**
- `trigger` (required) — Trigger type (e.g., `on_create`, `on_update`, `on_delete`)
- `entity` (required) — Entity type (e.g., `orders`, `products`)

**Pre-loaded context:** Action types with schemas, trigger types

**Guides the agent through:** Graph design, start edges, action connections, output port routing, DAG validation rules

### `configure-page`

Guide the user through configuring a page layout.

**Arguments:**
- `entity` (required) — Entity type (e.g., `orders`, `products`)

**Pre-loaded context:** Content types, field types

**Guides the agent through:** Page config creation, content blocks (tables, forms, charts, containers, tabs), layout (colSpan, gridCols, gap, containerType, responsive breakpoints)

### `design-form`

Guide the user through designing a data entry form.

**Arguments:**
- `entity` (required) — Entity type (e.g., `orders`, `products`)

**Pre-loaded context:** Field types with schemas

**Guides the agent through:** Database schema exploration, column-to-field-type mapping, validation configuration, field type selection

---

## Response Shapes

### Workflow Save Request (`POST /v1/workflow/rules/full`)

```json
{
  "name": "Low Stock Alert",
  "description": "Notify when inventory drops below threshold",
  "is_active": true,
  "entity_id": "uuid-of-entity",
  "trigger_type_id": "uuid-of-trigger-type",
  "trigger_conditions": {},
  "canvas_layout": {},
  "actions": [
    {
      "id": null,
      "name": "check_stock",
      "description": "Check current stock level",
      "action_type": "check_inventory",
      "action_config": { "warehouse_id": "..." },
      "is_active": true
    }
  ],
  "edges": [
    {
      "source_action_id": "",
      "target_action_id": "temp:0",
      "edge_type": "start",
      "source_output": "",
      "edge_order": 0
    },
    {
      "source_action_id": "temp:0",
      "target_action_id": "temp:1",
      "edge_type": "sequence",
      "source_output": "insufficient",
      "edge_order": 0
    }
  ]
}
```

**Edge conventions:**
- `source_action_id: ""` with `edge_type: "start"` = start edge
- `target_action_id: "temp:N"` = reference to actions array index N (for new actions)
- `edge_type`: `start`, `sequence`, or `always`
- `source_output`: output port name from the source action (e.g., `success`, `failure`, `true`, `false`, `sufficient`, `insufficient`)

### Workflow Save Response

```json
{
  "id": "uuid",
  "name": "Low Stock Alert",
  "description": "...",
  "is_active": true,
  "entity_id": "uuid",
  "trigger_type_id": "uuid",
  "trigger_conditions": {},
  "canvas_layout": {},
  "actions": [
    {
      "id": "uuid",
      "name": "check_stock",
      "description": "...",
      "action_type": "check_inventory",
      "action_config": { ... },
      "is_active": true
    }
  ],
  "edges": [
    {
      "id": "uuid",
      "source_action_id": "",
      "target_action_id": "uuid",
      "edge_type": "start",
      "source_output": "",
      "edge_order": 0
    }
  ],
  "created_date": "2026-01-15T10:00:00Z",
  "updated_date": "2026-01-15T10:00:00Z"
}
```

### Action type validation

Each `action_type` must be one of:
`allocate_inventory`, `check_inventory`, `check_reorder_point`, `commit_allocation`, `create_alert`, `create_entity`, `delay`, `evaluate_condition`, `log_audit_entry`, `lookup_entity`, `release_reservation`, `reserve_inventory`, `seek_approval`, `send_email`, `send_notification`, `transition_status`, `update_field`

### Edge type validation

Each `edge_type` must be one of: `start`, `sequence`, `always`

---

## Typical Agent Workflow

A well-behaved agent follows a **discover → read → validate → write** sequence:

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

---

## Config Surface Catalog (Full List)

All 12 surfaces returned by `GET /v1/agent/catalog`:

### UI Surfaces

| Surface | CRUD Endpoints | Discovery URL |
|---------|---------------|---------------|
| **Page Configs** | `GET/POST/PUT/DELETE /v1/config/page-configs/...` | `GET /v1/config/page-configs/all` |
| **Page Content** | `GET/POST/PUT/DELETE /v1/config/page-content/...` | `GET /v1/config/schemas/content-types` |
| **Page Actions** | `GET/POST/PUT/DELETE /v1/config/page-actions/...` | — |
| **Table Configs** | `GET/POST/PUT/DELETE /v1/data/...` | `GET /v1/config/schemas/table-config` |
| **Forms** | `GET/POST/PUT/DELETE /v1/config/forms/...` | `GET /v1/config/form-field-types` |
| **Form Fields** | `GET/POST/PUT/DELETE /v1/config/form-fields/...` | `GET /v1/config/form-field-types` |

### Workflow Surfaces

| Surface | Key Endpoints | Discovery URL |
|---------|--------------|---------------|
| **Workflow Rules** | `POST /v1/workflow/rules/full` (full save) | `GET /v1/workflow/action-types` |
| **Action Templates** | `GET /v1/workflow/templates` (read-only) | `GET /v1/workflow/templates/active` |
| **Alerts** | `GET /v1/workflow/alerts`, `POST .../acknowledge`, `POST .../dismiss` | — |

### System Surfaces

| Surface | Key Endpoints | Discovery URL |
|---------|--------------|---------------|
| **Action Permissions** | `POST /v1/workflow/actions/{type}/execute` | — |
| **Enum Labels** | `GET /v1/config/enums/{schema}/{name}/options` | `GET /v1/introspection/enums/{schema}` |
| **Database Introspection** | `GET /v1/introspection/schemas` (progressive) | `GET /v1/introspection/schemas` |

---

## Workflow Action Types (Full List)

All 17 types returned by `GET /v1/workflow/action-types`:

| Type | Category | Output Ports | Async? |
|------|----------|-------------|--------|
| `allocate_inventory` | inventory | success, failure | yes |
| `check_inventory` | inventory | sufficient, insufficient | no |
| `check_reorder_point` | inventory | ok, needs_reorder | no |
| `commit_allocation` | inventory | success, failure | no |
| `create_alert` | communication | success, failure | no |
| `create_entity` | data | success, failure | no |
| `delay` | control | success | no |
| `evaluate_condition` | control | true, false | no |
| `log_audit_entry` | data | success, failure | no |
| `lookup_entity` | data | found, not_found, failure | no |
| `release_reservation` | inventory | success, failure | no |
| `reserve_inventory` | inventory | success, failure | no |
| `seek_approval` | approval | success, failure | no |
| `send_email` | communication | success, failure | yes |
| `send_notification` | communication | success, failure | no |
| `transition_status` | data | success, invalid_transition, failure | no |
| `update_field` | data | success, failure | no |

**Output ports** determine which edge to follow after an action completes. Edges reference these by name in `source_output`.

---

## Form Field Types (Full List)

All 14 types returned by `GET /v1/config/form-field-types`:

| Type | Description | Key Config Properties |
|------|-------------|----------------------|
| `text` | Single-line text input | defaultValue, validation (min/max length) |
| `textarea` | Multi-line text | defaultValue, validation (min/max length) |
| `number` | Numeric input | min, max, step, precision |
| `currency` | Money input | precision, step |
| `percent` | Percentage input | precision, min, max |
| `date` | Date picker | validation (minDate, maxDate, mustBeFuture) |
| `datetime` | Date + time picker | same as date |
| `time` | Time picker | validation |
| `dropdown` | Select from entity records | **entity** (required), **label_column** (required), **value_column** (required), displayColumns, autoPopulate |
| `smart-combobox` | Searchable dropdown | same as dropdown + tableConfigName |
| `enum` | PostgreSQL enum values | **enumName** (required) |
| `checkbox` | Boolean toggle | defaultValue |
| `hidden` | Auto-populated hidden field | defaultValue, defaultValueCreate, defaultValueUpdate |
| `lineitems` | Inline child entities | **entity** (required), **parentField** (required), **fields[]** (required), minItems, maxItems |

Each field type has a JSON Schema for its config JSONB — retrieve via `GET /v1/config/form-field-types/{type}/schema`.

---

## Complete MCP → REST Endpoint Mapping

For quick reference, every MCP tool and the exact Ichor REST endpoint(s) it calls:

| MCP Tool | HTTP Method | Ichor Endpoint |
|----------|------------|----------------|
| `discover_config_surfaces` | GET | `/v1/agent/catalog` |
| `discover_action_types` | GET | `/v1/workflow/action-types` |
| `discover_field_types` | GET | `/v1/config/form-field-types` |
| `discover_trigger_types` | GET | `/v1/workflow/trigger-types` |
| `discover_entity_types` | GET | `/v1/workflow/entity-types` |
| `discover_entities` | GET | `/v1/workflow/entities` |
| `discover_content_types` | GET | `/v1/config/schemas/content-types` |
| `get_page_config` | GET | `/v1/config/page-configs/id/{id}` or `/name/{name}` |
| `get_page_content` | GET | `/v1/config/page-configs/content/{id}` |
| `get_table_config` | GET | `/v1/data/id/{id}` or `/name/{name}` |
| `get_form_definition` | GET | `/v1/config/forms/{id}/full` or `/name/{name}/full` |
| `list_pages` | GET | `/v1/config/page-configs/all` |
| `list_forms` | GET | `/v1/config/forms` |
| `list_table_configs` | GET | `/v1/data/configs/all` |
| `get_workflow` | GET | `/v1/workflow/rules/{id}` + `/actions` + `/edges` |
| `list_workflows` | GET | `/v1/workflow/rules` |
| `list_action_templates` | GET | `/v1/workflow/templates` |
| `search_database_schema` | GET | `/v1/introspection/schemas` → `/tables` → `/columns` + `/relationships` |
| `search_enums` | GET | `/v1/introspection/enums/{schema}` or `/v1/config/enums/{schema}/{name}/options` |
| `validate_workflow` | POST | `/v1/workflow/rules/full?dry_run=true` |
| `create_workflow` | POST | `/v1/workflow/rules/full` (+ dry_run) |
| `update_workflow` | PUT | `/v1/workflow/rules/{id}/full` (+ dry_run) |
| `create_page_config` | POST | `/v1/config/page-configs` |
| `update_page_config` | PUT | `/v1/config/page-configs/id/{id}` |
| `create_page_content` | POST | `/v1/config/page-content` |
| `update_page_content` | PUT | `/v1/config/page-content/{id}` |
| `create_form` | POST | `/v1/config/forms` |
| `add_form_field` | POST | `/v1/config/form-fields` |
| `create_table_config` | POST | `/v1/data` |
| `update_table_config` | PUT | `/v1/data/{id}` |
| `validate_table_config` | POST | `/v1/data/validate` |
| `analyze_workflow` | GET | `/v1/workflow/rules/{id}` + `/actions` + `/edges` |
| `suggest_templates` | GET | `/v1/workflow/templates/active` + `/action-types` |
| `show_cascade` | GET | `/v1/workflow/rules` |

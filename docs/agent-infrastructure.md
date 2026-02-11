# Agent Infrastructure

API endpoints that make Ichor self-describing for LLM agents, MCP tooling, and frontend code generators. These endpoints expose metadata about what's configurable, what shapes the data takes, and what constraints apply.

## Overview

The agent infrastructure consists of 5 API packages spread across the codebase:

| Package | Route Prefix | Purpose |
|---------|-------------|---------|
| `catalogapi` | `/v1/agent/catalog` | "What can I configure?" — lists all surfaces with endpoints |
| `configschemaapi` | `/v1/config/schemas/` | JSON schemas for opaque JSONB columns (table config, layout) |
| `formfieldschemaapi` | `/v1/config/form-field-types` | Field type discovery with config schemas |
| `referenceapi` | `/v1/workflow/` | Action type, trigger type, entity type discovery |
| `introspectionapi` | `/v1/introspection/` | Database schema browsing (tables, columns, relationships, enums) |

Plus dry-run validation on the workflow save API:

| Package | Route | Purpose |
|---------|-------|---------|
| `workflowsaveapi` | `POST /v1/workflow/rules/full?dry_run=true` | Validate workflow graph without saving |

## Config Surface Catalog

**Package**: `api/domain/http/agentapi/catalogapi/`
**Endpoint**: `GET /v1/agent/catalog`

Returns a static list of all 12 configurable surfaces in the system. Each entry includes:
- CRUD endpoint URLs
- Discovery/schema URLs
- Human-readable constraint summaries

### Surfaces

| Surface | Category | Key Endpoints |
|---------|----------|--------------|
| Page Configs | ui | CRUD at `/v1/config/page-configs/` |
| Page Content | ui | CRUD at `/v1/config/page-content/` |
| Page Actions | ui | CRUD at `/v1/config/page-actions/` |
| Table Configs | ui | CRUD at `/v1/data/` |
| Forms | ui | CRUD at `/v1/config/forms/` |
| Form Fields | ui | CRUD at `/v1/config/form-fields/` |
| Workflow Rules | workflow | Full save at `/v1/workflow/rules/full` |
| Action Templates | workflow | Read-only at `/v1/workflow/templates` |
| Alerts | workflow | Read + ack/dismiss at `/v1/workflow/alerts/` |
| Action Permissions | system | Execute at `/v1/workflow/actions/` |
| Enum Labels | system | Options at `/v1/config/enums/` |
| Database Introspection | system | Read-only at `/v1/introspection/` |

### Key Files

| File | Purpose |
|------|---------|
| `catalogapi/catalogapi.go` | Handler + static catalog data (12 `ConfigSurface` entries) |
| `catalogapi/model.go` | `ConfigSurface`, `Endpoints`, `Catalog` types |
| `catalogapi/route.go` | Route registration |

### Adding a New Surface

Edit `catalogapi/catalogapi.go` and add a new `ConfigSurface` to the `catalog` slice. No database changes needed — the catalog is a Go literal.

## Table Config & Layout JSON Schemas

**Package**: `api/domain/http/config/configschemaapi/`

Exposes machine-readable JSON schemas for JSONB columns that are otherwise opaque to agents.

### Endpoints

| Route | Returns |
|-------|---------|
| `GET /v1/config/schemas/table-config` | JSON Schema for `config.table_configs.config` JSONB |
| `GET /v1/config/schemas/layout` | JSON Schema for `config.page_content.layout` JSONB |
| `GET /v1/config/schemas/content-types` | Valid content types with descriptions and requirements |

### Table Config Schema

The table config JSONB is deeply nested. The schema covers:
- `data_source[]` — query definitions with table, schema, columns, joins, filters, sort, metrics, group_by
- `visual_settings.columns` — per-column config (type, header, width, format, editability, links, lookups)
- `visual_settings.conditional_formatting[]` — rule-based styling
- `visual_settings.row_actions[]` / `table_actions[]` — UI actions
- `visual_settings.pagination` — page sizes and defaults
- `permissions` — role-based access
- Chart types (when `visualization: "chart"`)

Source of truth for the schema: `business/sdk/tablebuilder/model.go`

### Layout Schema

Describes the responsive layout system used by page content blocks:
- `colSpan` / `gridCols` — `ResponsiveValue` with breakpoints (default, sm, md, lg, xl, 2xl)
- `containerType` — grid-12, flex, stack, tab, accordion, section, grid
- `gap` — Tailwind CSS gap classes

Source of truth: `business/domain/config/pagecontentbus/model.go`

### Key Files

| File | Purpose |
|------|---------|
| `configschemaapi/schemas/table_config.json` | Table config JSON Schema with `$defs` |
| `configschemaapi/schemas/layout.json` | Layout JSON Schema |
| `configschemaapi/schemas/content_types.json` | Content types list |
| `configschemaapi/schemas.go` | `//go:embed` loader + `init()` parser |
| `configschemaapi/configschemaapi.go` | 3 handlers |

### Updating Schemas

When the Go model structs change, update the corresponding JSON schema files in `schemas/`. Run `go test ./api/domain/http/config/configschemaapi/` to verify schemas are still valid JSON.

## Form Field Type Discovery

**Package**: `api/domain/http/config/formfieldschemaapi/`

Exposes all form field types with their config JSON schemas so agents know what configuration each field type accepts.

### Endpoints

| Route | Returns |
|-------|---------|
| `GET /v1/config/form-field-types` | All field types with names, descriptions, and config schemas |
| `GET /v1/config/form-field-types/{type}/schema` | Single field type's config schema |

### Field Types

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
| `dropdown` | Select from entity | entity, labelColumn, valueColumn, displayColumns, autoPopulate |
| `smart-combobox` | Searchable dropdown | same as dropdown + tableConfigName |
| `enum` | PostgreSQL enum | enumName (required) |
| `checkbox` | Boolean toggle | defaultValue |
| `hidden` | Auto-populated | defaultValue, defaultValueCreate, defaultValueUpdate |
| `lineitems` | Inline child entities | entity, parentField, fields[], minItems, maxItems |

### Key Files

| File | Purpose |
|------|---------|
| `formfieldschemaapi/schemas/*.json` | Per-field-type JSON schema files |
| `formfieldschemaapi/fieldtypes.go` | Metadata map + `init()` loader |
| `formfieldschemaapi/formfieldschemaapi.go` | 2 handlers |

Source of truth for schemas: `business/domain/config/formfieldbus/model.go` (FormFieldConfig, DropdownConfig, LineItemsFieldConfig)

## Workflow Action Type Discovery

**Package**: `api/domain/http/workflow/referenceapi/`

Exposes all 17 workflow action types with their JSON config schemas, output ports, categories, and metadata.

### Endpoints

| Route | Returns |
|-------|---------|
| `GET /v1/workflow/action-types` | All 17 action types with schemas and output ports |
| `GET /v1/workflow/action-types/{type}/schema` | Single action type's config schema |
| `GET /v1/workflow/trigger-types` | Available trigger types |
| `GET /v1/workflow/entity-types` | Entity types for triggers |
| `GET /v1/workflow/entities` | Registered entities |

### Action Types (17 total)

| Type | Category | Output Ports |
|------|----------|-------------|
| `allocate_inventory` | inventory | success, failure |
| `check_inventory` | inventory | sufficient, insufficient |
| `check_reorder_point` | inventory | ok, needs_reorder |
| `commit_allocation` | inventory | success, failure |
| `create_alert` | communication | success, failure |
| `create_entity` | data | success, failure |
| `delay` | control | success |
| `evaluate_condition` | control | true, false |
| `log_audit_entry` | data | success, failure |
| `lookup_entity` | data | found, not_found, failure |
| `release_reservation` | inventory | success, failure |
| `reserve_inventory` | inventory | success, failure |
| `seek_approval` | approval | success, failure |
| `send_email` | communication | success, failure |
| `send_notification` | communication | success, failure |
| `transition_status` | data | success, invalid_transition, failure |
| `update_field` | data | success, failure |

### Key Files

| File | Purpose |
|------|---------|
| `referenceapi/actionschemas.go` | Metadata map, `GetActionTypes()` merge with live registry |
| `referenceapi/schemas/*.json` | Per-action-type JSON config schemas |
| `referenceapi/model.go` | `ActionTypeInfo` with OutputPorts field |

## Dry-Run Validation

**Package**: `api/domain/http/workflow/workflowsaveapi/`

The workflow save endpoints support `?dry_run=true` to validate a workflow graph without committing to the database.

### Usage

```
POST /v1/workflow/rules/full?dry_run=true
PUT  /v1/workflow/rules/{id}/full?dry_run=true
```

### Response

```json
{
  "valid": true,
  "errors": [],
  "action_count": 5,
  "edge_count": 6
}
```

Or on failure:

```json
{
  "valid": false,
  "errors": ["graph contains a cycle", "action 'check_stock' has no outgoing edges"],
  "action_count": 5,
  "edge_count": 6
}
```

### Key Files

| File | Purpose |
|------|---------|
| `workflowsaveapi/workflowsaveapi.go` | Parses `dry_run` query param, routes to `DryRunValidate` |
| `workflowsaveapp/workflowsaveapp.go` | `DryRunValidate()` — runs full validation without DB write |
| `workflowsaveapp/model.go` | `ValidationResult` response type |

## Database Introspection

**Package**: `api/domain/http/introspectionapi/`

Read-only endpoints for browsing the PostgreSQL schema. Used by agents to understand table structure before building forms, table configs, or workflow rules.

### Endpoints

| Route | Returns |
|-------|---------|
| `GET /v1/introspection/schemas` | All PostgreSQL schemas |
| `GET /v1/introspection/schemas/{schema}/tables` | Tables in a schema |
| `GET /v1/introspection/tables/{schema}/{table}/columns` | Column definitions (name, type, nullable, default) |
| `GET /v1/introspection/tables/{schema}/{table}/relationships` | Foreign key relationships |
| `GET /v1/introspection/tables/{schema}/{table}/referencing-tables` | Tables that reference this one |
| `GET /v1/introspection/enums/{schema}` | Enum types in a schema |
| `GET /v1/introspection/enums/{schema}/{name}` | Values of a specific enum |
| `GET /v1/config/enums/{schema}/{name}/options` | Enum values merged with human-friendly labels |

## Decisions

- **Static catalog**: The catalog is a Go literal, not database-driven. Adding a new surface means editing `catalogapi.go`. This avoids a migration and keeps the catalog in sync with route registrations.

- **Embedded JSON schemas**: All JSON schema files use `//go:embed` and are loaded at init time. This means schema validation happens at startup, not per-request.

- **Passthrough architecture**: The MCP server wraps these endpoints but doesn't add logic — it calls the Ichor API and passes responses through. Changes to these endpoints automatically appear in the MCP server.

- **Dry-run as query param**: `?dry_run=true` uses the same request body as regular save. This means agents can take a payload they plan to submit and test it without a separate validation endpoint or request format.

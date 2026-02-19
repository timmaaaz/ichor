# Table Builder AI Assistant — Implementation Plan

> **Goal**: The AI assistant in the simple table builder page (`/admin/config/table-configs/:id`)
> progressively fills out a config — one atomic concern at a time — with validation and
> preview at every step. The LLM orchestrates intent; tools do the heavy lifting.

---

## System Overview (Current State)

### The Simple Table Builder Page

`QuickTableBuilder.vue` (mounted at `/admin/config/table-configs/:id`) holds:
- `configState: Ref<TableBuilderState>` — the live editor state (simplified builder format)
- Child panels: `BaseTableSelector`, `RelatedTablesSelector`, `SimpleColumnSelector`,
  `SimpleFilterBuilder`, `SimpleSortBuilder`
- `AssistantDrawer` — the AI chat panel
- `MutationPreview` (re-used from workflow) — the accept/reject card

**On every chat message**, the frontend bundles the entire current `TableBuilderState`
into the POST body as `context.state`. The LLM always has a fresh snapshot.

### The Two Wire Formats

These are DIFFERENT shapes. The LLM receives `TableBuilderState` but must output `Config`.

| TableBuilderState field         | Config wire format field                                  |
|---------------------------------|-----------------------------------------------------------|
| `baseTable.schema / .table`     | `data_source[0].schema` / `data_source[0].source`         |
| `dataSources[]`                 | `data_source[0].select.foreign_tables[]` (for joins)      |
| `columns[].alias` (or .column)  | `data_source[0].select.columns[].name`                    |
| `columns[]` (type, format, etc) | `visual_settings.columns` map, keyed by column name/alias |
| `filters[]`                     | `data_source[0].filters[]`                                |
| `joins[]`                       | `data_source[0].select.foreign_tables[]` (join shape)     |
| `sortBy[]`                      | `data_source[0].sort[]`                                   |

### Existing Tool Inventory

| Tool                     | What it does                                                        | Status        |
|--------------------------|---------------------------------------------------------------------|---------------|
| `get_table_config`       | Fetches full Config wire format by id or name                       | Exists, works |
| `preview_table_config`   | ValidateConfig() + emits SSE `table_config_preview` event           | Exists, works |
| `validate_table_config`  | Dry-run ValidateConfig() — no SSE, returns errors/warnings          | Exists, works |
| `update_table_config`    | PUT to backend directly (bypasses preview — should not be LLM-used) | Exists        |
| `search_database_schema` | Introspects real DB: schemas → tables → columns + relationships     | Exists, works |
| `discover_table_reference` | Static reference: PG type → visual type, format options, operators | Exists, works |
| `list_table_configs`     | Lists all saved configs                                             | Exists        |
| `create_table_config`    | Creates new config (frontend can't act on this yet)                 | Exists        |

### The Preview Flow (frontend)

```
AI calls preview_table_config
  → SSE table_config_preview fires: {description, config_id, name, config, is_update}
  → useTableBuilderMutations.showPreview() stores payload + shows card
  → User clicks Accept
    → applyMutations() → PUT /v1/data/{id} with full Config JSON
    → toast "Table config updated by assistant"
    → onReload() → loadConfig() → re-maps Config → configState (UI updates)
  → User clicks Reject
    → rejectPreview() → card dismisses, state unchanged
```

**Known gap in applyMutations**: Only handles `is_update: true`. Create path not implemented.
This is intentional — we work with existing configs only.

### Core Tools Wired (current, after Phase 1 of this worktree)

```go
coreToolsByContext["tables"] = ["get_table_config", "discover_table_reference", "preview_table_config"]
```

---

## What's Wrong With the Current Approach

1. **LLM constructs the full Config JSON in one shot** — error-prone, cognitively expensive
   for the LLM, produces no intermediate feedback to the user.

2. **No operation-specific tools** — the LLM has to know: how to map PG types, maintain
   the all-or-nothing ordering constraint, ensure filters reference selected columns,
   add hidden columns before filter columns, handle disambiguation aliases. All of this
   should be in tools, not LLM reasoning.

3. **validate_table_config exists but the prompt doesn't instruct the LLM to use it**
   before calling preview. The LLM goes straight to preview with potentially broken configs.

4. **No incremental build** — the user can't see the table taking shape step by step.
   Each accepted preview replaces the whole config. There's no "add one column at a time"
   flow with the UI reflecting each accepted step.

5. **System prompt doesn't sequence operations** — no per-operation playbooks, no guidance
   on what to do when an operation depends on a prior one (e.g., filter requires column).

---

## Phased Implementation Plan

### Phase 1 — Current State Documentation ✅ DONE
Already implemented (this worktree, feature/guided-table-creation):
- `tablesRoleBlock`: explains context format, modification workflow
- `tablesConstraintsGuidance`: wire format mapping, 6 key rules
- `coreToolsByContext["tables"]`: fixed to `[get_table_config, discover_table_reference, preview_table_config]`

**Files**: `chatapi/prompt.go`, `chatapi/chatapi.go`

---

### Phase 2 — Operation-Specific Backend Tools

**Goal**: New tools that perform one atomic operation each. Each tool:
1. Accepts the current `Config` JSON + operation-specific params
2. Applies the operation internally (handling all constraints)
3. Calls `ValidateConfig()` before returning
4. Returns `{valid, errors, warnings, config}` on success or `{valid: false, errors}` on failure

The LLM should call one of these operation tools, then call `preview_table_config` with
the returned `config` if valid.

#### Tool: `apply_column_change`
Adds or removes columns. Handles internally:
- PG type → visual type mapping (calls introspection API if type unknown)
- Adding to BOTH `select.columns` AND `visual_settings.columns`
- datetime format injection
- All-or-nothing ordering: if other visible columns have order, assigns next order value
- On remove: removes from both select.columns and visual_settings.columns

**Input schema:**
```json
{
  "config": "<current Config JSON>",
  "operation": "add" | "remove",
  "columns": [
    {
      "name": "column_name",
      "source_table": "table_name",
      "source_schema": "schema_name",
      "pg_type": "timestamp without time zone",
      "alias": "Created At",
      "hidden": false
    }
  ]
}
```

**Output:**
```json
{
  "valid": true,
  "config": { /* updated Config JSON */ },
  "warnings": [],
  "applied": ["column_name added to select.columns and visual_settings.columns"]
}
```
Or on failure:
```json
{
  "valid": false,
  "errors": [{ "field": "...", "message": "..." }],
  "config": null
}
```

**File**: `business/sdk/agenttools/executor.go` — new `handleApplyColumnChange()`
**Tool def**: `business/sdk/agenttools/definitions.go` — new tool in `TableToolDefinitions()`
**Catalog**: `business/sdk/toolcatalog/toolcatalog.go` — add `ApplyColumnChange` constant + GroupTables

#### Tool: `apply_filter_change`
Adds or removes filters. Handles internally:
- Checks that the filter column is in `select.columns`
- If not: adds as a hidden column first (auto-calls the column add logic)
- Validates operator is valid for the column's visual type
- On remove: removes filter only (does NOT remove the hidden column — that's explicit)

**Input schema:**
```json
{
  "config": "<current Config JSON>",
  "operation": "add" | "remove",
  "filter": {
    "column": "status",
    "operator": "eq",
    "value": "active",
    "label": "Status filter"
  }
}
```

**File**: `business/sdk/agenttools/executor.go` — new `handleApplyFilterChange()`

#### Tool: `apply_join_change`
Adds or removes foreign table joins. Handles internally:
- On add: builds the `foreign_tables` entry with correct relationship fields
- Disambiguation alias assignment when same table joined twice
- Optional: surfaces specified columns from the joined table (delegates to column add logic)
- On remove: removes the foreign table + all columns sourced from it

**Input schema:**
```json
{
  "config": "<current Config JSON>",
  "operation": "add" | "remove",
  "join": {
    "table": "users",
    "schema": "core",
    "join_type": "LEFT",
    "relationship_from": "orders.created_by",
    "relationship_to": "users.id",
    "columns_to_add": ["first_name", "last_name"]
  }
}
```

**File**: `business/sdk/agenttools/executor.go` — new `handleApplyJoinChange()`

#### Tool: `apply_sort_change`
Adds, removes, or reorders sort columns. Simple but important to isolate.

**Input schema:**
```json
{
  "config": "<current Config JSON>",
  "operation": "add" | "remove" | "set",
  "sort": [
    { "column": "created_date", "direction": "desc" }
  ]
}
```

**File**: `business/sdk/agenttools/executor.go` — new `handleApplySortChange()`

#### Notes
- All tools validate with `ValidateConfig()` before returning
- All tools return the same `{valid, errors, warnings, config}` shape
- The LLM calls tool → gets config back → calls `preview_table_config` with that config
- `preview_table_config` is still the SSE trigger — operation tools just produce validated configs

**Status**: ⬜ Pending

---

### Phase 3 — System Prompt Rewrite: Operation Playbooks

**Goal**: Replace `tablesConstraintsGuidance` with a new constant `tablesOperationsGuidance`
that gives the LLM per-operation step sequences and the validate-before-preview pattern.

#### The Validate-Preview Pattern (must be explicit)
```
For every change:
1. Call the appropriate operation tool (apply_column_change, apply_filter_change, etc.)
2. If tool returns valid: true → call preview_table_config with the returned config
3. If tool returns valid: false → explain errors to user, ask how to proceed
4. NEVER call preview_table_config with a config that hasn't come from an operation tool or get_table_config
```

#### Per-Operation Playbooks

**Adding columns:**
1. If column source unknown: `search_database_schema` with schema + table
2. Call `apply_column_change` with operation="add", columns=[{name, pg_type, source_table, source_schema}]
3. If valid → `preview_table_config`
4. Tell user what column was added and ask them to accept

**Adding a filter:**
1. Identify the column to filter on (it may already be in context.state.columns)
2. Call `apply_filter_change` with operation="add", filter={column, operator, value}
3. Tool will auto-add as hidden column if not already selected
4. If valid → `preview_table_config`

**Adding a join:**
1. `search_database_schema` on the target table to find columns and relationships
2. Call `apply_join_change` with operation="add", join={table, schema, join_type, relationship_from, relationship_to}
3. Optionally include columns_to_add if user wants specific columns from the joined table
4. If valid → `preview_table_config`

**Changing sort:**
1. Call `apply_sort_change` with the desired sort columns
2. If valid → `preview_table_config`

**Complex requests (e.g., "show inventory items with warehouse name, filter active only"):**
1. Decompose: base table + columns + maybe a join + maybe a filter
2. Handle in order: columns first, then joins (if needed), then filters
3. Aim for one preview per logical group, not one per individual column

#### Tone
- One operation at a time
- After each preview sent: "Preview ready — please accept or reject before we continue"
- Don't batch too many changes into a single config pass

**File**: `chatapi/prompt.go` — replace `tablesConstraintsGuidance` with `tablesOperationsGuidance`
**Status**: ⬜ Pending (depends on Phase 2 tool names being final)

---

### Phase 4 — Core Tools Update

Update `coreToolsByContext["tables"]` to include the new operation tools so they're
always available without RAG needing to surface them.

```go
coreToolsByContext["tables"] = []string{
    "get_table_config",
    "discover_table_reference",
    "apply_column_change",
    "apply_filter_change",
    "preview_table_config",
}
```

`apply_join_change` and `apply_sort_change` can be RAG-discoverable (less common operations).

**File**: `chatapi/chatapi.go`
**Status**: ⬜ Pending (depends on Phase 2)

---

### Phase 5 — Frontend: Live Preview in Builder UI

**Goal**: When the AI proposes a change via `table_config_preview`, the table builder page
should show the proposed config in the preview panel BEFORE the user accepts, so they can
see the table taking shape.

#### Current behavior
- Preview card appears with description text only
- Builder UI shows the CURRENT (last saved) state — not the proposed state
- User accepts blind (can read description but can't see the rendered table)

#### Desired behavior
- When `table_config_preview` fires: update a `previewConfig` ref in the builder with the
  proposed Config (but DO NOT save it yet)
- The table builder panels (columns list, filters list, etc.) show the proposed state
  highlighted/differentiated from the current state
- The live table preview (`ConfigPreview.vue`) rerenders with the proposed config
- User sees exactly what the table will look like before accepting
- On Accept: save + make previewConfig the new baseConfig
- On Reject: discard previewConfig, revert UI to last saved state

#### Files to modify (frontend — separate Vue PR)
- `QuickTableBuilder.vue`: add `previewConfig` ref, pass to preview-aware child components
- `useTableBuilderMutations.ts`: expose `previewConfig` from `showPreview()`
- `ConfigPreview.vue`: accept optional `overrideConfig` prop to render proposed state
- `SimpleColumnSelector.vue`: show proposed columns with "pending" indicator
- `SimpleFilterBuilder.vue`: show proposed filters with "pending" indicator

**Status**: ⬜ Pending (separate frontend PR, can be parallelized with Phase 2-4)

---

### Phase 6 — Validation Error UX (Frontend)

**Goal**: When an operation tool returns `valid: false`, the LLM explains the errors and
the frontend should surface them clearly.

Currently: LLM just describes errors in chat text. No structured display.

#### What to add
- The SSE stream already has a `tool_call_result` event type
- When `preview_table_config` returns errors, the frontend should:
  - Show an inline error card in the chat (not just text)
  - List each `ValidationError` with its field and message
  - Offer a "try again" or "edit manually" action

**Files**: `AssistantMessage.vue`, `AssistantDrawer.vue`
**Status**: ⬜ Pending (lower priority, can do after Phase 2-4)

---

## Key Decisions

### Why operation-specific tools instead of full-config manipulation?
The LLM is good at understanding intent ("add a filter for active status") but bad at
maintaining invariants across a 200-line JSON blob (ordering constraints, hidden columns
for filters, both select.columns and visual_settings.columns staying in sync).
Baking invariant maintenance into tools keeps the LLM doing what it's good at.

### Why is preview still a separate step?
Operation tools return a validated config but don't trigger the SSE event.
This gives the LLM control over WHEN to show a preview (e.g., it can call multiple
operation tools to build up a complex change, then preview the cumulative result).

### Why not one mega-tool with an operation enum?
Separate tools have clear, typed input schemas. Gemini Flash 2.5 performs better with
focused tool descriptions than a polymorphic params object.

### Why does the frontend need changes for live preview?
Without it, the user accepts blind. The builder UI reflecting proposed state before
accept is what makes the "filling in the page" experience feel natural. The backend
changes (Phases 2-4) can ship without Phase 5, but Phase 5 is what makes it great.

---

## File Index

| File | Concern |
|------|---------|
| `api/domain/http/agentapi/chatapi/prompt.go` | System prompts, operation playbooks |
| `api/domain/http/agentapi/chatapi/chatapi.go` | Core tools, tool routing |
| `business/sdk/agenttools/definitions.go` | Tool definitions + JSON schemas |
| `business/sdk/agenttools/executor.go` | Tool implementations |
| `business/sdk/toolcatalog/toolcatalog.go` | Tool group membership |
| `business/sdk/tablebuilder/model.go` | Config, DataSource, VisualSettings types |
| `business/sdk/tablebuilder/validation.go` | ValidateConfig() |
| `vue/.../QuickTableBuilder.vue` | Builder page, preview state |
| `vue/.../useTableBuilderMutations.ts` | Preview flow (accept/reject) |
| `vue/.../useTableBuilderAssistantContext.ts` | Context serialization |
| `vue/.../tableBuilderAssistant.ts` | SSE event types |

---

## Progress Tracker

| Phase | Description                         | Status      | Notes |
|-------|-------------------------------------|-------------|-------|
| 1     | Current state documented + prompts  | ✅ Done     | This worktree |
| 2     | Operation-specific backend tools    | ⬜ Pending  | 4 new tools in executor.go |
| 3     | System prompt operation playbooks   | ⬜ Pending  | Depends on Phase 2 tool names |
| 4     | Core tools update in chatapi.go     | ⬜ Pending  | Depends on Phase 2 |
| 5     | Frontend live preview in builder    | ⬜ Pending  | Separate Vue PR |
| 6     | Validation error UX (frontend)      | ⬜ Pending  | Lower priority |

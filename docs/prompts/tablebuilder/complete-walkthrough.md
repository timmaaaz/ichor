# Complete Table Builder Walkthrough

This document walks through ALL steps of building a table from scratch using the chat agent. Use this as your primary test script to validate the full table builder flow.

**Test subject**: `inventory.inventory_items` — a real table with columns, foreign keys, date fields, and boolean fields.

---

## Prerequisites

1. The app is running (local or dev cluster)
2. You have a valid auth token: `export TOKEN=$(make token)`
3. The table builder chat endpoint is accessible at `/v1/agent/chat`

---

## Phase 1: Discover the Table

**Goal**: Make the agent inspect the database schema before building anything.

### Message 1

```
I want to build a table for inventory.inventory_items.
What columns are available and what types are they?
```

**What happens:**
- Agent calls `search_database_schema` on `inventory.inventory_items`
- Agent lists the columns with their database types
- **No preview sent** — this is discovery only

**What to check:** The agent describes columns and mentions their types (varchar, integer, boolean, timestamp, uuid, etc.)

---

## Phase 2: Create the Base Table

**Goal**: Create the initial config with 4–5 core columns.

### Message 2

```
Great. Create a new table called "Inventory Stock View" with these columns:
- item_number
- quantity
- reorder_point
- unit_cost
- created_date (format as MM/dd/yyyy)
```

**What happens:**
1. Agent calls `search_database_schema` for pg_types (or reuses from Message 1)
2. Agent calls `apply_column_change` with all 5 columns
3. Agent calls `preview_table_config`
4. **Preview sent to UI** — you accept it

**What to check in the preview config:**
- `title: "Inventory Stock View"`
- All 5 columns in `data_source[0].select.columns`
- In `visual_settings.columns`:
  - `item_number`: type `string`
  - `quantity`: type `number`
  - `reorder_point`: type `number`
  - `unit_cost`: type `number`
  - `created_date`: type `datetime`, format `{ "format": "MM/dd/yyyy" }`
- No validation errors

**Accept the preview.** The frontend will call `create_table_config` and return a config ID.

---

## Phase 3: Add a Hidden ID Column

**Goal**: Add `warehouse_id` as a hidden column (needed for the join in Phase 4).

### Message 3

```
Add the warehouse_id column but keep it hidden — I need it for filtering but don't want to display it.
```

**What happens:**
1. Agent calls `get_table_config` (auto-fills config_id from context)
2. Agent calls `apply_column_change` to add `warehouse_id` with `hidden: true`
3. Agent calls `preview_table_config`

**What to check:**
- `warehouse_id` in `select.columns`
- `visual_settings.columns.warehouse_id.hidden: true`
- `visual_settings.columns.warehouse_id.type: "uuid"`

**Accept the preview.**

---

## Phase 4: Add a Join

**Goal**: Join `inventory.warehouses` and show the warehouse name.

### Message 4

```
Join the inventory.warehouses table on warehouse_id (left join).
Show the warehouse name. Call the column "Warehouse" in the header.
```

**What happens:**
1. Agent calls `get_table_config`
2. Agent calls `search_database_schema` on `inventory.warehouses`
3. Agent calls `apply_join_change`:
   - `operation: "add"`
   - `table: "warehouses"`, `schema: "inventory"`, `join_type: "left"`
   - `relationship_from: "warehouse_id"`, `relationship_to: "id"`
   - `columns_to_add: ["name"]`
4. Agent calls `preview_table_config`

**What to check:**
- `data_source[0].select.foreign_tables` has one entry for warehouses
- `warehouse_name` (or `name` with alias) in the foreign table's `columns`
- `visual_settings.columns.warehouse_name.header: "Warehouse"`
- `visual_settings.columns.warehouse_name.type: "string"`

**Accept the preview.**

---

## Phase 5: Add Filters

**Goal**: Add two filters — active items only, and positive quantity.

### Message 5

```
Add two filters:
1. Only show items where is_active is true
2. Only show items where quantity is greater than 0
```

**What happens:**
1. Agent calls `get_table_config`
2. Agent calls `apply_filter_change` for `is_active eq true`
   - `is_active` auto-added as hidden column if not already selected
3. Agent calls `apply_filter_change` again for `quantity gt 0`
   - Chained: second call uses config returned by first call
4. Agent calls `preview_table_config` once with both filters

**What to check:**
```json
"filters": [
  { "column": "is_active", "operator": "eq", "value": true },
  { "column": "quantity",  "operator": "gt", "value": 0 }
]
```

**Accept the preview.**

---

## Phase 6: Configure Sort

**Goal**: Sort by quantity ascending (low stock first).

### Message 6

```
Sort this table by quantity ascending so the lowest-stock items appear first.
```

**What happens:**
1. Agent calls `get_table_config`
2. Agent calls `apply_sort_change` with `operation="set"`, `sort=[{column: "quantity", direction: "asc"}]`
3. Agent calls `preview_table_config`

**What to check:**
```json
"sort": [{ "column": "quantity", "direction": "asc", "priority": 1 }]
```

**Accept the preview.**

---

## Phase 7: Set Column Display Order

**Goal**: Control the order columns appear in the table.

### Message 7

```
Set the display order for the columns:
1. item_number
2. warehouse (warehouse name)
3. quantity
4. reorder_point
5. unit_cost
6. created_date
```

**What happens:**
- Agent calls `apply_column_change` (or a combination of tools) to set `order` values in `visual_settings.columns`
- If `is_active` and `warehouse_id` are hidden, they don't need order values

**What to check:**
- All visible columns in `visual_settings.columns` have non-zero `order` values
- The `order` constraint: all-or-none — if any column has an order, ALL visible columns must have one

**Accept the preview.**

---

## Phase 8: Remove a Column

**Goal**: Remove `unit_cost` — we decided we don't need it.

### Message 8

```
Remove the unit_cost column.
```

**What happens:**
1. Agent calls `get_table_config`
2. Agent calls `apply_column_change` with `operation="remove"`, `columns=["unit_cost"]`
3. Agent calls `preview_table_config`

**What to check:**
- `unit_cost` NOT in `select.columns`
- `unit_cost` NOT in `visual_settings.columns`
- Column order values on remaining columns (still valid, possibly updated)

**Accept the preview.**

---

## Phase 9: Remove a Filter

**Goal**: Remove the `quantity > 0` filter.

### Message 9

```
Remove the filter on quantity.
```

**What happens:**
1. Agent calls `get_table_config`
2. Agent calls `apply_filter_change` with `operation="remove"`, `filter={column: "quantity"}`
3. Agent calls `preview_table_config`

**What to check:**
- Only one filter remains: `is_active eq true`

**Accept the preview.**

---

## Phase 10: Add a Dynamic Filter

**Goal**: Add a runtime-configurable filter so users can search by item_number.

### Message 10

```
Add a dynamic filter on item_number so users can search by item number at runtime.
Use ilike operator. Label it "Search by Item #".
```

**What to check:**
```json
{ "column": "item_number", "operator": "ilike", "dynamic": true, "label": "Search by Item #" }
```

**Accept the preview.**

---

## Phase 11: Final Validation

**Goal**: Confirm the final config is fully valid.

### Message 11

```
Can you validate the current table config and confirm everything looks correct?
```

**Expected behavior:**
- Agent calls `validate_table_config` (or `get_table_config` and reports)
- Agent summarizes: how many columns, what filters, what sort, what joins are configured
- **No validation errors**

---

## Final Config Checklist

After completing all phases, the config should have:

- [ ] `title: "Inventory Stock View"`
- [ ] `widget_type: "table"`, `visualization: "table"`
- [ ] Base columns: `item_number`, `quantity`, `reorder_point`, `created_date`
- [ ] Hidden columns: `warehouse_id`, `is_active`
- [ ] Foreign table: `inventory.warehouses` → `name` (aliased as `warehouse_name`)
- [ ] Filters: `is_active eq true`, `item_number ilike dynamic`
- [ ] Sort: `quantity asc`
- [ ] Column order set on all visible columns (1–5)
- [ ] All visible columns have valid `type` in `visual_settings.columns`
- [ ] `created_date` has `format: { "format": "MM/dd/yyyy" }`
- [ ] No validation errors or warnings

---

## Quick Cheat Sheet — Prompt Patterns

| Goal | Prompt template |
|------|-----------------|
| Discover columns | `What columns are available in [schema].[table]?` |
| Create table | `Create a table called "[Name]" using [schema].[table]. Columns: [list]` |
| Add column | `Add the [column] column` |
| Add datetime column | `Add [column] formatted as MM/dd/yyyy` |
| Add hidden column | `Add [column] but keep it hidden` |
| Add join | `Join [schema].[table] on [fk_column] (left join) and show [column(s)]` |
| Add filter | `Filter where [column] [operator] [value]` |
| Add dynamic filter | `Add a dynamic filter on [column] using [operator]. Label it "[label]"` |
| Sort | `Sort by [column] ascending/descending` |
| Remove column | `Remove the [column] column` |
| Remove filter | `Remove the filter on [column]` |
| Remove join | `Remove the join to [table]` |
| Validate | `Validate the current config and tell me if there are any errors` |

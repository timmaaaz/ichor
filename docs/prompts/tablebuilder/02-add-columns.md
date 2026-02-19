# Step 02: Add Columns to an Existing Table

**Goal**: Add one or more columns to an already-saved table config. This exercises `get_table_config` → `apply_column_change` → `preview_table_config`.

**Prerequisite**: Complete Step 01 first and note the `config_id` of the saved config.

---

## Context Setup

Pass the `config_id` from the saved config:

```json
{
  "message": "<prompt>",
  "context_type": "tables",
  "context": {
    "config_id": "<uuid-from-step-01>",
    "state": {
      "baseTable": "inventory.inventory_items",
      "columns": [
        { "source": "inventory_items", "column": "item_number", "alias": "item_number", "data_type": "varchar", "column_type": "string" },
        { "source": "inventory_items", "column": "quantity",    "alias": "quantity",    "data_type": "integer", "column_type": "number" }
      ],
      "joins": [],
      "filters": [],
      "sortBy": []
    }
  }
}
```

> **Tip**: In the UI, the `context.state` is sent automatically from the editor state — you don't construct it manually. These prompts show the shape for reference when testing via curl.

---

## Prompt 2A — Add a single column

```
Add the description column to this table.
```

**Expected agent behavior:**
1. Calls `get_table_config` to get wire-format config
2. Calls `search_database_schema` if description's pg_type is unknown
3. Calls `apply_column_change` with `operation="add"` for `description`
4. Calls `preview_table_config`

**What to verify:**
- `description` appears in `select.columns`
- `description` appears in `visual_settings.columns` with type `string`

---

## Prompt 2B — Add multiple columns at once

```
Add the unit_cost, reorder_point, and last_counted_date columns.
Format last_counted_date as MM/dd/yyyy.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Calls `search_database_schema` to find pg_types for all three
3. Calls `apply_column_change` once with all three columns (or once per column — both are valid)
4. Calls `preview_table_config` once with a summary description

**What to verify:**
- All three columns added to `select.columns`
- `unit_cost` has type `number`
- `last_counted_date` has type `datetime` with format `MM/dd/yyyy`
- `reorder_point` has type `number`

---

## Prompt 2C — Add a UUID column (ID column)

```
Add the warehouse_id column. I want to use it for filtering but not display it to users.
```

**Expected agent behavior:**
1. Adds `warehouse_id` with type `uuid` in visual settings
2. Sets `hidden: true` on the column config (it will be selected but not displayed)

**What to verify:**
- `warehouse_id` in `select.columns`
- `visual_settings.columns.warehouse_id.hidden: true`
- `visual_settings.columns.warehouse_id.type: "uuid"`

---

## Prompt 2D — Add a boolean column

```
Add the is_active column.
```

**What to verify:**
- `is_active` has type `boolean` in visual settings

---

## Prompt 2E — Rename a column with an alias

```
Add the notes column, but display it with the header "Internal Notes".
```

**What to verify:**
- `visual_settings.columns.notes.header: "Internal Notes"`
- Column `name` is still `notes` (the DB column name)

---

## Ordering Constraint

If you want explicit column ordering, all visible columns must have `order` values:

```
Add the sku column and set the display order to: item_number(1), sku(2), quantity(3), reorder_point(4).
```

**What to verify:**
- ALL visible columns in `visual_settings.columns` have non-zero `order` values
- The validation rule: "if any column has an explicit order, ALL visible columns must have an order"

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| `missing column type: new_col` | Agent forgot to add type to visual_settings | Re-run with explicit type request |
| `ordering constraint: all or none must have order` | Some columns have order, some don't | Ask agent to set order on ALL columns |
| `invalid type "text"` | Wrong visual type (should be "string") | Remind agent: pg_type "text" → visual type "string" |

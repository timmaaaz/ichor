# Step 01: Create a Basic Table Config

**Goal**: Start a brand-new table config from scratch using the inventory items table. This exercises `search_database_schema`, `apply_column_change`, and `preview_table_config`.

---

## Context Setup

For a new table, omit `config_id` (the agent will create a new config, not update an existing one):

```json
{
  "message": "<prompt>",
  "context_type": "tables",
  "context": {
    "state": {
      "baseTable": "",
      "columns": [],
      "joins": [],
      "filters": [],
      "sortBy": []
    }
  }
}
```

---

## Prompt 1A — Simple table, specific columns

Use this when you know exactly what you want:

```
Create a new table called "Inventory Items" showing the inventory.inventory_items table.
Include these columns: item_number, quantity, reorder_point, and created_date.
Format created_date as yyyy-MM-dd.
```

**Expected agent behavior:**
1. Calls `search_database_schema` on `inventory.inventory_items` to find pg_types
2. Calls `apply_column_change` with `operation="add"` for each column
3. Calls `preview_table_config` with a description of the added columns
4. You accept the preview → table config is created

**What to verify:**
- Preview shows 4 columns in visual settings
- `created_date` has type `datetime` with format `yyyy-MM-dd`
- `quantity` and `reorder_point` have type `number`
- `item_number` has type `string`

---

## Prompt 1B — Let the agent discover columns

Use this when you want the agent to suggest useful columns:

```
I want to create a new inventory items table using inventory.inventory_items.
What columns are available? Show me the most useful ones for a stock management view.
```

**Expected agent behavior:**
1. Calls `search_database_schema` on `inventory.inventory_items`
2. Reports back available columns with their types
3. Suggests a set of useful columns
4. Waits for your approval before building

**Follow-up after the agent responds:**

```
Yes, add those columns. Also include created_date formatted as MM/dd/yyyy.
```

---

## Prompt 1C — Minimal table (just an ID column to start)

Use this to test the minimum valid config:

```
Create a minimal table config called "Test Inventory" for inventory.inventory_items.
Just add the item_number column for now.
```

**What to verify:**
- Config is valid with a single column
- `item_number` has type `string` in visual settings
- Title is "Test Inventory"

---

## Validation Checkpoints

After accepting the preview, the config should have:

```json
{
  "title": "Inventory Items",
  "widget_type": "table",
  "visualization": "table",
  "data_source": [{
    "type": "query",
    "source": "inventory_items",
    "schema": "inventory",
    "select": {
      "columns": [
        { "name": "item_number" },
        { "name": "quantity" },
        { "name": "reorder_point" },
        { "name": "created_date" }
      ]
    }
  }],
  "visual_settings": {
    "columns": {
      "item_number":    { "name": "item_number",    "header": "Item Number",    "type": "string" },
      "quantity":       { "name": "quantity",       "header": "Quantity",       "type": "number" },
      "reorder_point":  { "name": "reorder_point",  "header": "Reorder Point",  "type": "number" },
      "created_date":   { "name": "created_date",   "header": "Created Date",   "type": "datetime", "format": { "type": "date", "format": "yyyy-MM-dd" } }
    }
  }
}
```

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| `missing column type: created_date` | datetime column missing format config | Add `format: { type: "date", format: "yyyy-MM-dd" }` |
| `invalid type "timestamp"` | pg_type passed as visual type | Agent should map pg_type → visual type via typemapper |
| Preview not received | Agent called `update_table_config` directly | Agent should only use `preview_table_config` |

---

## Notes

- The agent auto-fills `config_id` into tool calls from the context — for a new table, there is no ID yet, so `create_table_config` is what the frontend calls after accepting the preview.
- Column `type` in `visual_settings` must be one of: `string`, `number`, `datetime`, `boolean`, `uuid`, `status`, `computed`, `lookup`.
- `datetime` columns **must** have a `format` config using date-fns tokens (not Go format `2006-01-02`).

# Step 03: Add a Join (Foreign Table)

**Goal**: Join a related table and pull columns from it. This exercises `search_database_schema` → `apply_join_change` → `preview_table_config`.

**Prerequisite**: Complete Step 01 (have a table config with at least an ID column from the base table).

---

## Context Setup

```json
{
  "message": "<prompt>",
  "context_type": "tables",
  "context": {
    "config_id": "<uuid-from-step-01>",
    "state": {
      "baseTable": "inventory.inventory_items",
      "columns": [...],
      "joins": [],
      "filters": [],
      "sortBy": []
    }
  }
}
```

---

## Prompt 3A — Simple left join

```
Join the inventory.warehouses table to show the warehouse name alongside each inventory item.
Use a left join on warehouse_id.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Calls `search_database_schema` on `inventory.warehouses` to discover columns and relationships
3. Calls `apply_join_change` with:
   - `operation: "add"`
   - `join.table: "warehouses"`, `join.schema: "inventory"`
   - `join.join_type: "left"`
   - `join.relationship_from: "warehouse_id"` (from inventory_items)
   - `join.relationship_to: "id"` (on warehouses)
   - `columns_to_add: ["name"]` (the warehouse name column)
4. Calls `preview_table_config`

**What to verify in the config:**
```json
"select": {
  "foreign_tables": [{
    "table": "warehouses",
    "schema": "inventory",
    "join_type": "left",
    "relationship_from": "warehouse_id",
    "relationship_to": "id",
    "columns": [
      { "name": "name", "alias": "warehouse_name" }
    ]
  }]
}
```

And in `visual_settings.columns`:
```json
"warehouse_name": { "name": "warehouse_name", "header": "Warehouse", "type": "string" }
```

---

## Prompt 3B — Join and bring in multiple columns

```
Join the products table (products schema) to show the product name and SKU.
The inventory_items table has a product_id foreign key that links to products.id.
Use a left join.
```

**What to verify:**
- `foreign_tables` has one entry for `products.products`
- Both `name` and `sku` columns appear in the join's `columns` array
- Both appear in `visual_settings.columns` with type `string`

---

## Prompt 3C — Join with column aliasing

```
Join inventory.warehouses on warehouse_id. I want the warehouse name column,
but call it "Storage Location" in the header.
```

**What to verify:**
- `columns: [{ "name": "name", "alias": "storage_location" }]` in the foreign table
- `visual_settings.columns.storage_location.header: "Storage Location"`

---

## Prompt 3D — Nested join (join from a joined table)

```
Join inventory.warehouses on warehouse_id to get the warehouse name.
Then also join geography.cities through the warehouse's city_id to show the city name.
```

**Expected structure:**
```json
"foreign_tables": [{
  "table": "warehouses",
  "schema": "inventory",
  "join_type": "left",
  "relationship_from": "warehouse_id",
  "relationship_to": "id",
  "columns": [{ "name": "name", "alias": "warehouse_name" }],
  "foreign_tables": [{
    "table": "cities",
    "schema": "geography",
    "join_type": "left",
    "relationship_from": "city_id",
    "relationship_to": "id",
    "columns": [{ "name": "name", "alias": "city_name" }]
  }]
}]
```

---

## Prompt 3E — Multiple joins to the same table (aliased)

```
This table has both a created_by and updated_by column, both pointing to core.users.
Add both joins so I can show the creator's and updater's names.
Use aliases "creator" and "updater" for the two joins.
```

**What to verify:**
- Two entries in `foreign_tables` with distinct `alias` values
- `alias: "creator"` on the first join, `alias: "updater"` on the second
- Columns aliased to avoid collision: e.g., `creator_name`, `updater_name`

---

## Removing a Join

```
Remove the join to the warehouses table.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Calls `apply_join_change` with `operation: "remove"` and the join identifier
3. Calls `preview_table_config`

**What to verify:**
- `foreign_tables` no longer contains the warehouses join
- Any columns that came from warehouses are removed from `select.columns` and `visual_settings.columns`

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| `missing column type: warehouse_name` | Joined column not added to visual_settings | Agent must add all joined columns to visual_settings |
| Join added but column not visible | Agent added to foreign_tables but forgot visual_settings | Ask agent to confirm visual_settings was updated |
| Wrong join direction | relationship_from/to reversed | Correct: `relationship_from` is on the SOURCE table, `relationship_to` is on the TARGET table |

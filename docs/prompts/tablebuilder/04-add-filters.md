# Step 04: Add Filters

**Goal**: Add static and dynamic filters to a table config. This exercises `apply_filter_change` → `preview_table_config`.

**Prerequisite**: Have a saved table config with at least some columns.

---

## Filter Operators Reference

| Operator | Meaning | Example value |
|----------|---------|---------------|
| `eq` | Equal | `"active"`, `true`, `42` |
| `neq` | Not equal | `"inactive"` |
| `gt` | Greater than | `0` |
| `gte` | Greater than or equal | `10` |
| `lt` | Less than | `100` |
| `lte` | Less than or equal | `50` |
| `in` | In array | `["a", "b", "c"]` |
| `like` | SQL LIKE | `"%search%"` |
| `ilike` | Case-insensitive LIKE | `"%search%"` |
| `is_null` | IS NULL | _(no value needed)_ |
| `is_not_null` | IS NOT NULL | _(no value needed)_ |

---

## Prompt 4A — Static equality filter

```
Add a filter to only show active inventory items. Filter on the is_active column where it equals true.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Checks if `is_active` is already in the selected columns
3. Calls `apply_filter_change` with `operation="add"`, `filter={column: "is_active", operator: "eq", value: true}`
   - If `is_active` is not already selected, the tool auto-adds it as a hidden column
4. Calls `preview_table_config`

**What to verify in the config:**
```json
"filters": [
  { "column": "is_active", "operator": "eq", "value": true }
]
```

---

## Prompt 4B — Numeric threshold filter

```
Add a filter to only show items where quantity is greater than 0.
```

**What to verify:**
```json
{ "column": "quantity", "operator": "gt", "value": 0 }
```

---

## Prompt 4C — Multiple filters (AND logic)

```
Filter to show only active items (is_active = true) where quantity is less than reorder_point.
```

> Note: Cross-column comparisons (column vs column) may not be supported by the operator set. If the agent says this isn't supported, the filter should use a static threshold instead. Ask:
> ```
> Then filter where quantity is less than 10 as a proxy for low stock.
> ```

**What to verify:** Two filter entries in `filters[]`.

---

## Prompt 4D — Text search filter (ILIKE)

```
Add a case-insensitive filter on the item_number column that matches anything containing "INV".
```

**What to verify:**
```json
{ "column": "item_number", "operator": "ilike", "value": "%INV%" }
```

---

## Prompt 4E — Null check filter

```
Add a filter to show only items that have a description (description is not null).
```

**What to verify:**
```json
{ "column": "description", "operator": "is_not_null" }
```

---

## Prompt 4F — Dynamic filter (runtime value)

```
Add a dynamic filter on warehouse_id so users can filter by warehouse at runtime.
Label it "Warehouse".
```

**What to verify:**
```json
{ "column": "warehouse_id", "operator": "eq", "dynamic": true, "label": "Warehouse" }
```

---

## Prompt 4G — Filter on a joined column

After completing Step 03 (warehouse join):

```
Add a filter to only show items in active warehouses.
Filter on the warehouse's is_active column where it equals true.
```

**Note**: Filters on joined columns typically use the format `table.column` or the alias. Ask the agent to use the correct column reference.

---

## Removing Filters

```
Remove the filter on is_active.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Calls `apply_filter_change` with `operation="remove"`, `filter={column: "is_active"}`
3. Calls `preview_table_config`

**What to verify:**
- `filters` array no longer contains the `is_active` entry
- If `is_active` was auto-added as hidden (not explicitly selected), it should also be removed from `select.columns`

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| `invalid operator: "contains"` | Not a valid operator | Use `ilike` with `%value%` pattern |
| Column auto-added as visible | Agent added filtered column as visible instead of hidden | Ask agent to add the column as `hidden: true` |
| Filter value type mismatch | Boolean `"true"` (string) vs `true` (bool) | The agent should use the correct JSON type |

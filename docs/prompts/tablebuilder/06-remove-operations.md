# Step 06: Remove Columns, Filters, and Joins

**Goal**: Test removal operations. Each remove uses the same `apply_*_change` tool but with `operation="remove"`.

---

## Remove a Column

### Prompt 6A — Remove a single column

```
Remove the description column from this table.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Calls `apply_column_change` with `operation="remove"`, `columns=["description"]`
3. Calls `preview_table_config`

**What to verify:**
- `description` removed from `select.columns`
- `description` removed from `visual_settings.columns`
- If `description` was being used in a filter, the agent should warn you

---

### Prompt 6B — Remove multiple columns

```
Remove item_number and notes from this table.
```

**What to verify:**
- Both columns gone from `select.columns`
- Both gone from `visual_settings.columns`

---

### Prompt 6C — Remove a column that's used in a filter

```
Remove the is_active column.
```

**Expected behavior:** The agent should warn you that `is_active` is used in an active filter and ask how to proceed. You can respond:

```
Yes, remove the filter too and then remove the column.
```

---

## Remove a Filter

### Prompt 6D — Remove a specific filter

```
Remove the filter on quantity.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Calls `apply_filter_change` with `operation="remove"`, `filter={column: "quantity"}`
3. Calls `preview_table_config`

---

### Prompt 6E — Remove all filters

```
Remove all filters from this table.
```

**What to verify:** `filters: []`

---

## Remove a Join

### Prompt 6F — Remove a joined table

```
Remove the join to the warehouses table.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Calls `apply_join_change` with `operation="remove"` targeting the warehouses join
3. Calls `preview_table_config`

**What to verify:**
- `foreign_tables` no longer contains the warehouses entry
- Any columns that came from warehouses (e.g., `warehouse_name`) are removed from `visual_settings.columns`

---

## Edge Cases to Test

### Prompt 6G — Remove a hidden column (auto-added by filter)

After adding a filter on `is_active` (which auto-added it as hidden):

```
Remove the filter on is_active.
```

**Expected behavior:**
- Filter removed from `filters[]`
- Since `is_active` was added implicitly (hidden, not user-selected), it should also be removed from `select.columns`

---

### Prompt 6H — Try to remove a column that doesn't exist

```
Remove the foobar_column from this table.
```

**Expected behavior:** Agent reports the column doesn't exist and does nothing (no preview sent).

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Column removed from select but still in visual_settings | Partial removal | Config will fail validation; ask agent to also remove from visual_settings |
| Filter removed but hidden column stays selected | Tool didn't clean up auto-added columns | Ask agent to also remove the hidden column |
| Remove affects wrong column | Ambiguous column name in join | Specify table prefix: "remove the warehouses.name column" |

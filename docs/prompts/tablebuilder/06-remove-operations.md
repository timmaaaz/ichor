# Step 06: Remove Columns, Filters, and Joins

**What this tests**: Removing things you've already added — columns, filters, and joined tables.

---

## Remove a Column

### Prompt 6A — Remove a single column

```
Remove the description column from this table.
```

**What the AI will do:**
1. Load the current config
2. Remove `description` from the column list
3. Show you a preview

**What you should see:**
- "Description" column is gone from the preview
- No validation errors

> **Important**: If the column you're removing is used in a filter, the AI should warn you (see Prompt 6C below).

---

### Prompt 6B — Remove multiple columns

```
Remove item_number and notes from this table.
```

**What you should see:**
- Both columns gone from the preview
- Remaining columns are unaffected

---

### Prompt 6C — Remove a column that's used in a filter

```
Remove the is_active column.
```

**Expected behavior**: The AI should warn you that `is_active` is currently used in an active filter and ask how to proceed. When it does, respond with:

```
Yes, remove the filter too and then remove the column.
```

**What you should see:**
- Both the filter and the column are removed in the preview
- No orphaned filter left behind

---

## Remove a Filter

### Prompt 6D — Remove a specific filter

```
Remove the filter on quantity.
```

**What you should see:**
- The quantity filter is gone
- Other filters (if any) are unaffected

---

### Prompt 6E — Remove all filters

```
Remove all filters from this table.
```

**What you should see:**
- No filters in the preview (the "Filters" section is empty or absent)

---

## Remove a Join

### Prompt 6F — Remove a joined table

```
Remove the join to the warehouses table.
```

**What you should see:**
- Warehouse-related columns (e.g., "Warehouse Name") are gone from the preview
- No validation errors from orphaned columns

---

## Edge Cases to Test

### Prompt 6G — Remove a filter that auto-added a hidden column

After adding a filter on `is_active` (which caused `is_active` to be added as a hidden column):

```
Remove the filter on is_active.
```

**Expected behavior:**
- The filter is removed
- Since `is_active` was only added for filtering (hidden, not user-selected), it should also be removed from the columns automatically

**What you should see:** The AI removes both the filter and the hidden column.

---

### Prompt 6H — Try to remove a column that doesn't exist

```
Remove the foobar_column from this table.
```

**Expected behavior:** The AI should tell you the column doesn't exist and not make any changes. No preview is sent.

---

## Common Errors

| What you see | What caused it | What to try |
|---|---|---|
| Validation error after removing a column | Column was removed from the select list but its display settings weren't cleaned up | Ask: "Check for and fix any validation errors in the current config" |
| Hidden filter column wasn't removed after removing a filter | The AI didn't clean up the auto-added column | Ask: "Also remove the hidden is_active column since we removed the filter that needed it" |
| Wrong column removed when join table has a column with the same name | Ambiguous column name | Specify which table: "Remove the warehouses.name column" |

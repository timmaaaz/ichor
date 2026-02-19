# Step 05: Configure Sort Order

**Goal**: Set, append, or replace the sort configuration. This exercises `apply_sort_change` → `preview_table_config`.

**Prerequisite**: Have a saved table config with some columns.

---

## Sort Operations

| Operation | Behavior |
|-----------|----------|
| `set` | Replaces the ENTIRE sort with the provided columns |
| `add` | Appends new sort columns to existing sort |
| `remove` | Removes specific columns from the sort |

---

## Prompt 5A — Set a single sort column

```
Sort the table by item_number ascending.
```

**Expected agent behavior:**
1. Calls `get_table_config`
2. Calls `apply_sort_change` with `operation="set"`, `sort=[{column: "item_number", direction: "asc"}]`
3. Calls `preview_table_config`

**What to verify:**
```json
"sort": [
  { "column": "item_number", "direction": "asc", "priority": 1 }
]
```

---

## Prompt 5B — Sort by multiple columns with priority

```
Sort by quantity ascending first, then by item_number ascending as a secondary sort.
```

**What to verify:**
```json
"sort": [
  { "column": "quantity",    "direction": "asc", "priority": 1 },
  { "column": "item_number", "direction": "asc", "priority": 2 }
]
```

---

## Prompt 5C — Sort descending (newest first)

```
Sort by created_date descending so the newest items appear first.
```

**What to verify:**
```json
"sort": [
  { "column": "created_date", "direction": "desc", "priority": 1 }
]
```

---

## Prompt 5D — Add a sort column to existing sort

```
Add a secondary sort on item_number ascending.
```

**Expected agent behavior:**
- Uses `operation="add"` (not `set`) so it appends rather than replaces
- Sets a higher priority number for the secondary sort

---

## Prompt 5E — Replace entire sort

```
Replace the current sort. Sort only by reorder_point descending — highest priority items first.
```

**Expected agent behavior:**
- Uses `operation="set"` to replace whatever sort was previously configured

---

## Prompt 5F — Remove a sort column

```
Remove the sort on item_number. Keep the sort on quantity.
```

**Expected agent behavior:**
- Uses `operation="remove"` targeting only `item_number`
- The `quantity` sort remains unchanged

---

## Prompt 5G — Clear all sorting

```
Remove all sorting from this table.
```

**What to verify:**
- `sort: []` (empty array)

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Sort replaced unexpectedly | Agent used `set` instead of `add` | Clarify: "add a secondary sort, don't replace existing" |
| Sort not cleared | Agent used `remove` on each column individually | Ask for "clear all sorting" or "remove all sorts" |
| Priority numbers not set | Agent left out `priority` field | Ask agent to assign priorities: 1 for primary, 2 for secondary, etc. |

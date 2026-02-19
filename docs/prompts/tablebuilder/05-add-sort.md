# Step 05: Configure Sort Order

**What this tests**: Setting, adding to, or replacing the sort order for the table rows.

**Prerequisite**: Have a saved table with some columns.

---

## Sort Operations at a Glance

| Operation | What It Does |
|---|---|
| Set | Replaces the entire sort with new columns |
| Add | Adds a sort column without removing existing ones |
| Remove | Removes one column from the sort, keeps others |

---

## Prompt 5A — Sort by a single column

```
Sort the table by item_number ascending.
```

**What the AI will do:**
1. Load the current config
2. Set the sort to `item_number ascending`
3. Show you a preview

**What you should see:**
- Sort is configured: item_number, A→Z (ascending)
- The AI confirms this is now the primary sort

---

## Prompt 5B — Sort by multiple columns

Use this when you want a primary sort and a tiebreaker:

```
Sort by quantity ascending first, then by item_number ascending as a secondary sort.
```

**What you should see:**
- Primary sort: quantity (lowest first)
- Secondary sort: item_number (A→Z when quantities are equal)

---

## Prompt 5C — Newest first (descending date sort)

```
Sort by created_date descending so the newest items appear first.
```

**What you should see:**
- Sort: created_date, newest first (descending)

---

## Prompt 5D — Add to existing sort without replacing it

```
Add a secondary sort on item_number ascending.
```

**What the AI should do:**
- Add `item_number` as a secondary sort
- **Not** replace the existing primary sort

**What you should see:**
- Your existing primary sort is still there
- `item_number` is added as a secondary sort after it

---

## Prompt 5E — Replace all sorting

```
Replace the current sort. Sort only by reorder_point descending — highest priority items first.
```

**What you should see:**
- Previous sort is gone
- Only `reorder_point descending` remains

---

## Prompt 5F — Remove one sort column, keep others

```
Remove the sort on item_number. Keep the sort on quantity.
```

**What you should see:**
- `item_number` no longer in the sort
- `quantity` sort remains unchanged

---

## Prompt 5G — Clear all sorting

```
Remove all sorting from this table.
```

**What you should see:**
- No sort columns configured
- The AI confirms sort is cleared

---

## Common Errors

| What you see | What caused it | What to try |
|---|---|---|
| Existing sort was replaced when you wanted to add | AI used "set" instead of "add" | Clarify: "Add as a secondary sort without removing the existing sort" |
| Sort wasn't cleared | AI only removed one column | Ask: "Remove ALL sort columns, not just one" |
| Wrong column sorted | Column name was ambiguous (same name in multiple joined tables) | Specify which table: "Sort by the inventory_items.quantity column" |

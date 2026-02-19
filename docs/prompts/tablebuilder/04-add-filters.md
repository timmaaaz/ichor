# Step 04: Add Filters

**What this tests**: Adding rules that limit which rows are shown. Filters can be fixed (always applied) or dynamic (the user can change the value at runtime).

**Prerequisite**: Have a saved table with some columns.

---

## Filter Types at a Glance

| Filter Type | What It Does | Example |
|---|---|---|
| Static | Always filters to a fixed value | Always show only active items |
| Dynamic | Shows a search box — users enter the value | Let users filter by warehouse at runtime |

## Filter Operators Reference

| Operator | Plain English | Example |
|---|---|---|
| `eq` | Equals | `is_active = true` |
| `neq` | Does not equal | `status ≠ cancelled` |
| `gt` | Greater than | `quantity > 0` |
| `gte` | Greater than or equal to | `quantity ≥ 10` |
| `lt` | Less than | `quantity < 100` |
| `lte` | Less than or equal to | `quantity ≤ 50` |
| `in` | Is one of a list | `status in [pending, active]` |
| `ilike` | Contains (case-insensitive) | `item_number contains "INV"` |
| `is_null` | Has no value | `description is empty` |
| `is_not_null` | Has a value | `description is not empty` |

---

## Prompt 4A — Filter by a yes/no column

```
Add a filter to only show active inventory items. Filter on the is_active column where it equals true.
```

**What the AI will do:**
1. Load the current config
2. Add a filter: `is_active equals true`
3. If `is_active` isn't already a selected column, it'll be added as a hidden column automatically
4. Show you a preview

**What you should see:**
- The preview config now includes a filter rule for active items only
- The AI's response should confirm: "Added filter: is_active = true"

---

## Prompt 4B — Filter by a numeric threshold

```
Add a filter to only show items where quantity is greater than 0.
```

**What you should see:**
- A filter rule: quantity > 0
- The AI confirms the filter was added

---

## Prompt 4C — Multiple filters (both must match)

```
Filter to show only active items (is_active = true) where quantity is less than 10.
```

**What you should see:**
- Two filter rules applied — both must be true for a row to show
- Preview confirms both filters

---

## Prompt 4D — Text search (contains)

```
Add a case-insensitive filter on the item_number column that matches anything containing "INV".
```

**What you should see:**
- A filter that matches item numbers like `INV-001`, `INV-042`, etc.
- The AI uses the "ilike" operator (case-insensitive contains)

---

## Prompt 4E — Filter for rows with no value

```
Add a filter to show only items that have a description (description is not null).
```

**What you should see:**
- A filter that excludes items with no description

---

## Prompt 4F — Dynamic filter (user-controlled at runtime)

```
Add a dynamic filter on warehouse_id so users can filter by warehouse at runtime.
Label it "Warehouse".
```

**What you should see:**
- The preview includes a dynamic filter entry
- The AI confirms a "Warehouse" filter control will appear for users

> A **dynamic filter** doesn't filter the table immediately — it creates a filter control (like a dropdown or text box) that users interact with when they're using the table.

---

## Prompt 4G — Filter on a column from a joined table

After completing Step 03 (with a warehouse join):

```
Add a filter to only show items in active warehouses.
Filter on the warehouse's is_active column where it equals true.
```

**What you should see:**
- The filter references the warehouse's active status, not the item's
- Preview confirms the filter was added correctly

---

## Removing a Filter

```
Remove the filter on is_active.
```

**What you should see:**
- The filter is removed in the preview
- If `is_active` was only added for filtering (it was hidden), the AI may also remove it from the columns — that's correct behavior

---

## Common Errors

| What you see | What caused it | What to try |
|---|---|---|
| AI says "contains is not a valid operator" | "contains" isn't a valid operator name | Use `ilike` with `%value%` — the AI usually handles this automatically |
| Filter added but values are wrong type | AI used `"true"` (text) instead of `true` (boolean) | Ask: "The filter value for is_active should be a boolean true, not the text 'true'" |
| Dynamic filter doesn't appear as a filter control | The filter was added as static instead of dynamic | Ask: "Make the [column] filter dynamic with label '[label]'" |

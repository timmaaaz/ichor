# Step 02: Add Columns to an Existing Table

**What this tests**: Adding one or more columns to a table that's already been saved.

**Prerequisite**: Complete Step 01 first so you have a saved table to modify.

---

## Prompt 2A — Add a single column

```
Add the description column to this table.
```

**What the AI will do:**
1. Load the current table configuration
2. Look up the `description` column's type if needed
3. Add it to the table
4. Show you a preview

**What you should see in the preview:**
- A new "Description" column added to the table
- The column shows as text (string type)

---

## Prompt 2B — Add multiple columns at once

```
Add the unit_cost, reorder_point, and last_counted_date columns.
Format last_counted_date as MM/dd/yyyy.
```

**What you should see in the preview:**
- Three new columns added
- `unit_cost` shows as a number
- `last_counted_date` shows formatted dates like `01/15/2024`
- `reorder_point` shows as a number

---

## Prompt 2C — Add a hidden ID column

Sometimes you need a column for filtering purposes but don't want to display it to users:

```
Add the warehouse_id column. I want to use it for filtering but not display it to users.
```

**What you should see in the preview:**
- The AI confirms `warehouse_id` was added as a hidden column
- You won't see it as a visible column in the preview table — that's intentional

<details>
<summary>Network tab verification (optional)</summary>

Look for `visual_settings.columns.warehouse_id.hidden: true` in the preview JSON.

</details>

---

## Prompt 2D — Add a yes/no column

```
Add the is_active column.
```

**What you should see in the preview:**
- `is_active` shows as a boolean (true/false) column

---

## Prompt 2E — Rename a column header

```
Add the notes column, but display it with the header "Internal Notes".
```

**What you should see in the preview:**
- Column header reads "Internal Notes" in the table
- The underlying column name is still `notes`

---

## Prompt 2F — Set display order

Use this if you want to control the left-to-right column order in the table:

```
Add the sku column and set the display order to: item_number(1), sku(2), quantity(3), reorder_point(4).
```

**What you should see in the preview:**
- Columns appear in the specified order in the preview table

> **Note**: If you set an order for one column, you need to set it for all visible columns. The AI handles this automatically when you specify the full order.

---

## Common Errors

| What you see | What caused it | What to try |
|---|---|---|
| Validation error about "ordering constraint" | Some columns have an order set but others don't | Ask: "Set the display order for all columns" and list them numbered |
| New column appears with the wrong type | The AI mapped the database type incorrectly | Ask: "Update the [column] column to be type [string/number/date/boolean]" |
| Column added but not visible in preview | Column may have been added as hidden | Ask: "Make [column] visible" |

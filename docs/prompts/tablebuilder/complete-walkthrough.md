# Complete Table Builder Walkthrough

This is the primary test script. Run through it from top to bottom to validate the full table builder flow — from creating a brand-new table to removing columns and adding dynamic filters.

**What you'll build**: An inventory stock table using `inventory.inventory_items`, with joins, filters, sorting, and column ordering.

---

## Before You Start

1. Make sure the app is running and you're logged in
2. Navigate to the table builder and open the AI chat panel
3. You'll be sending 11 messages total — each message builds on the last

---

## Message 1: Discover the Table

**Goal**: Ask the AI to look up what's available before building anything.

**Send this:**
```
I want to build a table for inventory.inventory_items.
What columns are available and what types are they?
```

**What should happen:**
- The AI looks up the table schema
- It lists the available columns with their types
- **No preview is sent yet** — this is information gathering only

**Check**: The AI describes columns like `item_number`, `quantity`, `reorder_point`, `created_date`, etc. with their types (text, number, date, true/false, UUID).

---

## Message 2: Create the Base Table

**Goal**: Create the initial table with your core columns.

**Send this:**
```
Great. Create a new table called "Inventory Stock View" with these columns:
- item_number
- quantity
- reorder_point
- unit_cost
- created_date (format as MM/dd/yyyy)
```

**What should happen:**
1. The AI adds all 5 columns to the config
2. A preview appears in the UI

**Check in the preview:**
- Title: "Inventory Stock View"
- 5 columns visible: Item Number, Quantity, Reorder Point, Unit Cost, Created Date
- Created Date shows formatted dates (e.g., `01/15/2024`) not raw timestamps
- No validation error shown

**Accept the preview.** ✓

---

## Message 3: Add a Hidden ID Column

**Goal**: Add `warehouse_id` as a hidden column (you'll need it for the join in Message 4, but don't want to display it).

**Send this:**
```
Add the warehouse_id column but keep it hidden — I need it for filtering but don't want to display it.
```

**What should happen:**
1. The AI adds `warehouse_id` to the configuration
2. A preview appears

**Check in the preview:**
- The AI confirms `warehouse_id` was added as a hidden column
- You do NOT see it as a visible column in the table — that's correct

**Accept the preview.** ✓

---

## Message 4: Add a Join

**Goal**: Link the warehouses table so you can show the warehouse name.

**Send this:**
```
Join the inventory.warehouses table on warehouse_id (left join).
Show the warehouse name. Call the column "Warehouse" in the header.
```

**What should happen:**
1. The AI looks up the warehouses table
2. Adds the join and the warehouse name column
3. A preview appears

**Check in the preview:**
- A "Warehouse" column now appears in the table
- It shows warehouse names (like "Main Warehouse") not raw IDs
- No validation errors

**Accept the preview.** ✓

---

## Message 5: Add Filters

**Goal**: Limit the table to only show meaningful rows — active items with stock on hand.

**Send this:**
```
Add two filters:
1. Only show items where is_active is true
2. Only show items where quantity is greater than 0
```

**What should happen:**
1. The AI adds both filters
2. A single preview appears with both filters applied

**Check in the preview:**
- The AI confirms 2 filters were added:
  - is_active = true (active items only)
  - quantity > 0 (items with stock only)

**Accept the preview.** ✓

---

## Message 6: Configure Sort

**Goal**: Show the lowest-stock items first (they're the most urgent to reorder).

**Send this:**
```
Sort this table by quantity ascending so the lowest-stock items appear first.
```

**What should happen:**
1. The AI sets the sort
2. A preview appears

**Check in the preview:**
- Sort: quantity ascending (lowest first)

**Accept the preview.** ✓

---

## Message 7: Set Column Display Order

**Goal**: Control the left-to-right order of columns in the table.

**Send this:**
```
Set the display order for the columns:
1. item_number
2. warehouse (warehouse name)
3. quantity
4. reorder_point
5. unit_cost
6. created_date
```

**What should happen:**
- The AI sets the column order in the configuration
- A preview appears with columns in the specified order

**Check in the preview:**
- Columns appear in the order you specified
- Hidden columns (`warehouse_id`, `is_active`) don't need to be listed — they have no display order

**Accept the preview.** ✓

---

## Message 8: Remove a Column

**Goal**: Remove `unit_cost` — we decided we don't need it.

**Send this:**
```
Remove the unit_cost column.
```

**What should happen:**
1. The AI removes the column
2. A preview appears without it

**Check in the preview:**
- "Unit Cost" is gone from the table
- Remaining columns are still there and correctly ordered
- No validation errors

**Accept the preview.** ✓

---

## Message 9: Remove a Filter

**Goal**: Remove the quantity > 0 filter — we'll show zero-stock items too.

**Send this:**
```
Remove the filter on quantity.
```

**What should happen:**
1. The AI removes the quantity filter
2. A preview appears

**Check:**
- Only one filter remains: is_active = true
- The quantity filter is gone

**Accept the preview.** ✓

---

## Message 10: Add a Dynamic Filter

**Goal**: Add a search box so users can filter by item number at runtime.

**Send this:**
```
Add a dynamic filter on item_number so users can search by item number at runtime.
Use ilike operator. Label it "Search by Item #".
```

**What should happen:**
1. The AI adds a dynamic filter
2. A preview appears

**Check:**
- The AI confirms a dynamic filter was added
- The label is "Search by Item #"

> A dynamic filter creates a user-facing search control — users type in the table UI to filter rows. It doesn't filter the preview (which is just showing the configuration).

**Accept the preview.** ✓

---

## Message 11: Final Validation

**Goal**: Confirm everything is correct and there are no errors.

**Send this:**
```
Can you validate the current table config and confirm everything looks correct?
```

**What should happen:**
- The AI reviews the config and reports its findings
- **No validation errors** should be reported

**Check:**
- AI confirms the config is valid
- AI summarizes: roughly 5 visible columns + 1 hidden + 1 joined, 2 filters, 1 sort

---

## Final Checklist

After completing all 11 messages, your table should have:

- [ ] Title: "Inventory Stock View"
- [ ] Visible columns: item_number, warehouse name, quantity, reorder_point, created_date (in that order)
- [ ] Hidden columns: warehouse_id, is_active
- [ ] Join to inventory.warehouses showing the warehouse name with header "Warehouse"
- [ ] Filter: is_active = true (only active items)
- [ ] Dynamic filter: item_number ilike, labeled "Search by Item #"
- [ ] Sort: quantity ascending (lowest first)
- [ ] No validation errors

---

## Quick Reference — Prompt Patterns

Keep these handy for follow-up testing:

| Goal | What to say |
|------|-------------|
| Discover columns | `What columns are available in [schema].[table]?` |
| Create table | `Create a table called "[Name]" using [schema].[table]. Columns: [list]` |
| Add a column | `Add the [column] column` |
| Add a date column | `Add [column] formatted as MM/dd/yyyy` |
| Add a hidden column | `Add [column] but keep it hidden` |
| Add a join | `Join [schema].[table] on [fk_column] (left join) and show [column(s)]` |
| Add a static filter | `Filter where [column] [equals/is greater than/contains] [value]` |
| Add a dynamic filter | `Add a dynamic filter on [column] using [operator]. Label it "[label]"` |
| Set sort | `Sort by [column] ascending/descending` |
| Add secondary sort | `Add a secondary sort on [column] ascending, without replacing the existing sort` |
| Remove a column | `Remove the [column] column` |
| Remove a filter | `Remove the filter on [column]` |
| Remove a join | `Remove the join to the [table] table` |
| Validate | `Validate the current config and tell me if there are any errors` |

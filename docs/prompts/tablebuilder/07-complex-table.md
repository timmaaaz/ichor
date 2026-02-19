# Step 07: Complex Table in One Request

**What this tests**: The AI's ability to handle a multi-part request and build a complete table in a single go. This is the most realistic "real user" scenario.

---

## Prompt 7A — Full inventory dashboard in one message

```
Create a new inventory management table called "Inventory Dashboard" using inventory.inventory_items.

I want:
- Columns: item_number, quantity, reorder_point, unit_cost, created_date (formatted as MM/dd/yyyy)
- Join inventory.warehouses on warehouse_id (left join) and show the warehouse name
- Filter to only show active items (is_active = true) and items with quantity > 0
- Sort by quantity ascending (low stock first)
```

**What the AI will do** (you don't need to check each step — but here's what it should do internally):
1. Look up the `inventory_items` schema to check column types
2. Look up the `warehouses` schema to find columns
3. Add all 5 base columns
4. Add the warehouse join with the name column
5. Add both filters
6. Set the sort
7. Send a single preview at the end

**What you should see in the preview:**
- Table title: "Inventory Dashboard"
- 6 visible columns: Item Number, Quantity, Reorder Point, Unit Cost, Created Date, Warehouse (name)
- Created Date shows formatted dates (e.g., `01/15/2024`)
- The AI's description mentions: 2 filters, 1 join, 1 sort rule

**Validation checklist** (tick off as you verify):
- [ ] Title is "Inventory Dashboard"
- [ ] 5 base columns are present
- [ ] Warehouse Name column appears (from the join)
- [ ] The AI mentions filtering for active items with quantity > 0
- [ ] The AI mentions sorting by quantity ascending
- [ ] No validation errors shown

---

## Prompt 7B — Sales orders with customer info

```
Build a table called "Sales Orders" using sales.orders.

Include:
- Columns: order_number, status, total_amount, created_date (format: yyyy-MM-dd HH:mm:ss)
- Join core.users on customer_id (left join) — show first_name as "Customer Name"
- Filter: only show orders where status is not "cancelled"
- Sort: created_date descending (newest first)
```

**What you should see:**
- 5 columns (4 base + customer name from join)
- Dates include time (e.g., `2024-01-15 14:30:00`)
- AI mentions filter excludes cancelled orders
- Sort is newest orders at the top

---

## Prompt 7C — Product catalog with brand and category

```
Create a "Product Catalog" table from products.products.

Columns: name, sku, price, is_active
Join products.brands on brand_id (left join) — show brand name
Join products.categories on category_id (left join) — show category name
Filter: only active products (is_active = true)
Sort: name ascending
```

**What you should see:**
- 6 columns (4 base + brand name + category name, both from joins)
- Two joined tables appear in the config
- Filter for active only
- Products sorted A→Z

---

## Prompt 7D — Let the AI take the lead

This tests whether the AI asks good clarifying questions when the request is open-ended:

```
Build a user management table from core.users. Show me the most important columns and suggest some useful filters.
```

**Expected behavior:**
1. The AI looks up the schema and proposes a column set
2. The AI asks for your approval before building
3. You respond: "Yes, use those columns plus add a filter for active users"
4. The AI builds the config and sends a preview

**What you should check:**
- Does the AI ask before building? (Good behavior — it shouldn't just guess)
- Does it suggest reasonable columns? (Name, email, status, created date are typical)
- Does it apply your follow-up instruction correctly?

---

## Prompt 7E — Fix a broken config

Use this to test error recovery. If you have a config with a validation error, send:

```
The current table config has a validation error. Can you look at it and fix whatever is wrong?
```

**Expected behavior:**
1. The AI inspects the config
2. Identifies the problem (e.g., a column with no display type)
3. Fixes it
4. Sends a corrected preview

**What you should see:** A preview with no validation errors.

---

## Full Complexity Checklist (for Prompt 7A)

After accepting the preview, verify the table shows:

- [ ] Title "Inventory Dashboard" in the table header
- [ ] Item Number column (text)
- [ ] Quantity column (number)
- [ ] Reorder Point column (number)
- [ ] Unit Cost column (number)
- [ ] Created Date column showing formatted dates like `01/15/2024`
- [ ] Warehouse column showing warehouse names (not IDs)
- [ ] The table only shows active items with quantity > 0
- [ ] Rows are sorted with lowest quantity at the top

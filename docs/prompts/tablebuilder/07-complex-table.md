# Step 07: Complex Table in One Request

**Goal**: Test the agent's ability to decompose a complex multi-operation request and execute it correctly. The agent should chain operation tool calls and send a single preview.

---

## Prompt 7A — Full inventory dashboard in one shot

```
Create a new inventory management table called "Inventory Dashboard" using inventory.inventory_items.

I want:
- Columns: item_number, quantity, reorder_point, unit_cost, created_date (formatted as MM/dd/yyyy)
- Join inventory.warehouses on warehouse_id (left join) and show the warehouse name
- Filter to only show active items (is_active = true) and items with quantity > 0
- Sort by quantity ascending (low stock first)
```

**Expected agent behavior (in order):**
1. `search_database_schema` on `inventory.inventory_items` to find pg_types
2. `search_database_schema` on `inventory.warehouses` to find columns/relationships
3. `apply_column_change` with all 5 base columns
4. `apply_join_change` for warehouses + warehouse_name column
5. `apply_filter_change` for `is_active = true`
6. `apply_filter_change` for `quantity > 0`
7. `apply_sort_change` for `quantity asc`
8. `preview_table_config` with a summary description — ONE preview at the end

**What to verify:**
- All 5 base columns in `select.columns`
- `foreign_tables` has warehouses join with `name` column aliased as `warehouse_name`
- Two filters: `is_active eq true`, `quantity gt 0`
- Sort: `quantity asc, priority 1`
- All columns have types in `visual_settings.columns`
- `created_date` has `type: "datetime"` with format config

---

## Prompt 7B — Orders table with customer info

```
Build a table called "Sales Orders" using sales.orders.

Include:
- Columns: order_number, status, total_amount, created_date (format: yyyy-MM-dd HH:mm:ss)
- Join core.users on customer_id (left join) — show first_name and last_name as "Customer Name" (you'll need to combine them, or just show first_name)
- Filter: only show orders where status is not "cancelled" (neq operator)
- Sort: created_date descending (newest first)
```

---

## Prompt 7C — Products with brand and category

```
Create a "Product Catalog" table from products.products.

Columns: name, sku, price, is_active
Join products.brands on brand_id (left join) — show brand name
Join products.categories on category_id (left join) — show category name
Filter: only active products (is_active = true)
Sort: name ascending
```

**What to verify:**
- Two foreign table joins
- Both join columns appear in `visual_settings.columns`
- Single filter on `is_active`

---

## Prompt 7D — Multi-step with agent questions

This tests the agent asking clarifying questions when the request is ambiguous:

```
Build a user management table from core.users. Show me the most important columns and suggest some useful filters.
```

**Expected behavior:**
1. Agent calls `search_database_schema` on `core.users`
2. Agent proposes columns and asks for approval
3. You respond: "Yes, use those columns plus add a filter for active users"
4. Agent builds the config and sends a preview

---

## Prompt 7E — Fix a broken config

Use this to test error recovery. First create a config missing a column type, then:

```
The current table config has a validation error. Can you look at it and fix whatever is wrong?
```

**Expected behavior:**
1. Agent calls `get_table_config`
2. Agent calls `validate_table_config` to check errors
3. Agent identifies the issue (e.g., missing column type) and fixes it via `apply_column_change` or similar
4. Sends a preview of the corrected config

---

## Complexity Checklist

After completing Prompt 7A, verify the full config has all of these:

- [ ] `title` set
- [ ] `widget_type: "table"` and `visualization: "table"`
- [ ] `data_source[0].type: "query"`
- [ ] `data_source[0].source` and `schema` set correctly
- [ ] `select.columns` has all base columns
- [ ] `select.foreign_tables` has the warehouse join with columns
- [ ] `filters` has both filters
- [ ] `sort` has the sort entry
- [ ] `visual_settings.columns` has an entry for EVERY column (including joined columns)
- [ ] Every visible column in `visual_settings.columns` has a valid `type`
- [ ] `datetime` columns have a `format` config with a date-fns token
- [ ] No column uses Go date format (`2006-01-02` — wrong; use `yyyy-MM-dd` — correct)

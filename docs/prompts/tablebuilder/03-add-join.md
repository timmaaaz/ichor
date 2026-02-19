# Step 03: Add a Join (Pull In Data from a Related Table)

**What this tests**: Adding data from a related table — for example, showing the warehouse name next to each inventory item, when the items table only stores a `warehouse_id`.

**Prerequisite**: Complete Step 01 first (have a saved table with at least a few columns).

> **What is a join?** A join links two tables together using a shared ID. For example, `inventory_items` has a `warehouse_id` column that points to the `warehouses` table. A join lets you show the warehouse's name (from the warehouses table) alongside each inventory item.

---

## Prompt 3A — Simple join (show a related name)

```
Join the inventory.warehouses table to show the warehouse name alongside each inventory item.
Use a left join on warehouse_id.
```

**What the AI will do:**
1. Load the current config
2. Look up the `inventory.warehouses` table to find its columns
3. Add the join configuration
4. Add the warehouse `name` column to the table
5. Show you a preview

**What you should see in the preview:**
- A new "Warehouse Name" (or similar) column in the table
- Each row shows the warehouse name instead of just a raw ID

<details>
<summary>Network tab verification (optional)</summary>

Look for a `foreign_tables` entry with:
- `table: "warehouses"`, `schema: "inventory"`, `join_type: "left"`
- `relationship_from: "warehouse_id"`, `relationship_to: "id"`
- A `name` column inside that join's columns array

</details>

---

## Prompt 3B — Join and show multiple related columns

```
Join the products table (products schema) to show the product name and SKU.
The inventory_items table has a product_id foreign key that links to products.id.
Use a left join.
```

**What you should see in the preview:**
- Two new columns from the products table: Name and SKU
- Both appear as text columns

---

## Prompt 3C — Join with a custom column header

```
Join inventory.warehouses on warehouse_id. I want the warehouse name column,
but call it "Storage Location" in the header.
```

**What you should see in the preview:**
- A column with the header "Storage Location" (not "Warehouse Name")
- It displays the warehouse name from the related table

---

## Prompt 3D — Nested join (join from a joined table)

This is an advanced scenario: joining a table that's already joined to another table.

```
Join inventory.warehouses on warehouse_id to get the warehouse name.
Then also join geography.cities through the warehouse's city_id to show the city name.
```

**What you should see in the preview:**
- A "Warehouse Name" column
- A "City Name" column (coming from the city linked to each warehouse)

---

## Prompt 3E — Two joins to the same table (different purposes)

```
This table has both a created_by and updated_by column, both pointing to core.users.
Add both joins so I can show the creator's and updater's names.
Use aliases "creator" and "updater" for the two joins.
```

**What you should see in the preview:**
- Two user name columns: one for the creator, one for the updater
- They're named distinctly to avoid confusion (e.g., "Creator Name", "Updater Name")

---

## Removing a Join

```
Remove the join to the warehouses table.
```

**What you should see:**
- The warehouse name column disappears from the preview
- No validation errors remain

---

## Common Errors

| What you see | What caused it | What to try |
|---|---|---|
| Validation error about "missing column type" for the joined column | The AI added the join but forgot to configure the column's display type | Ask: "Make sure the warehouse_name column has a display type set" |
| Joined column doesn't appear in the preview | The AI set up the join but didn't add the column to display | Ask: "Also show the [column] column from the [table] join" |
| Wrong table linked (join direction reversed) | The AI linked the wrong column | Clarify: "The [source table] has a [fk_column] that links to [target table].[id]" |

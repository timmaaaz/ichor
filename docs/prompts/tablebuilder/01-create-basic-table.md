# Step 01: Create a Basic Table

**What this tests**: Starting a brand-new table configuration from scratch. The AI will look up the available columns in the database and then build the config based on your request.

**Prerequisite**: No existing table needed — this creates a new one.

---

## Prompt 1A — Specific columns you already know

Use this when you know exactly what columns you want:

```
Create a new table called "Inventory Items" showing the inventory.inventory_items table.
Include these columns: item_number, quantity, reorder_point, and created_date.
Format created_date as yyyy-MM-dd.
```

**What the AI will do:**
1. Look up the `inventory.inventory_items` table in the database to check column types
2. Add each column to the configuration
3. Show you a preview

**What you should see in the preview:**
- Table title: "Inventory Items"
- 4 columns: Item Number, Quantity, Reorder Point, Created Date
- `created_date` shows formatted dates like `2024-01-15` (not a raw timestamp)
- `quantity` and `reorder_point` appear as numbers
- `item_number` appears as text

<details>
<summary>Network tab verification (optional)</summary>

In the `table_config_preview` event, look for:
- `visual_settings.columns.item_number.type: "string"`
- `visual_settings.columns.quantity.type: "number"`
- `visual_settings.columns.created_date.type: "datetime"` with `format.format: "yyyy-MM-dd"`

</details>

---

## Prompt 1B — Let the AI suggest columns

Use this when you're not sure what columns are available:

```
I want to create a new inventory items table using inventory.inventory_items.
What columns are available? Show me the most useful ones for a stock management view.
```

**What the AI will do:**
1. Look up the table schema and list available columns
2. Suggest a set of useful columns
3. Wait for your approval before building anything

**What you should see:**
- A text response listing columns with descriptions
- No preview yet — the AI is asking for your input first

**Follow-up prompt** (send this after the AI lists its suggestions):

```
Yes, add those columns. Also include created_date formatted as MM/dd/yyyy.
```

---

## Prompt 1C — Minimal table (one column)

Use this to test the smallest valid configuration:

```
Create a minimal table config called "Test Inventory" for inventory.inventory_items.
Just add the item_number column for now.
```

**What you should see in the preview:**
- Table title: "Test Inventory"
- 1 column: Item Number
- Preview is valid (no errors shown)

---

## Common Errors

| What you see | What caused it | What to try |
|---|---|---|
| Preview shows a validation error mentioning "missing column type" | The AI forgot to assign a type to a date column | Re-send the prompt and explicitly ask to "format created_date as a date" |
| Preview not shown at all | The AI may have skipped the preview step | Ask: "Can you send me a preview of the current config?" |
| Date column shows raw timestamps like `2024-01-15T00:00:00Z` | Date format wasn't set correctly | Ask: "Update created_date to display as MM/dd/yyyy" |

---

## Notes

- Column types are automatically inferred from the database — you shouldn't need to specify them unless you want something specific
- Date formatting uses display tokens: `yyyy-MM-dd`, `MM/dd/yyyy`, `MM/dd/yyyy HH:mm:ss` — not technical formats
- After you Accept the preview, the table is saved and assigned a config ID you can use in later steps

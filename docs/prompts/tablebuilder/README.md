# Table Builder Chat Prompts

This directory contains test prompts for the **table builder AI assistant** — the in-app chat tool that lets you build and modify data tables by describing what you want in plain English.

## What is the Table Builder?

The table builder lets you configure data tables (like a filtered, sorted spreadsheet view of your data) by chatting with an AI assistant. You describe the columns, filters, and sorting you want, and the assistant builds it for you. Before anything is saved, it shows you a **preview** — you must click Accept before any changes are made.

---

## How to Run These Tests

### Step 1: Open the Table Builder Chat

Navigate to the table builder section of the app and open the AI chat panel. Make sure you're logged in.

### Step 2: Copy a Prompt

Pick a prompt from any of the test files below. Copy the text inside the code block (the gray box) and paste it into the chat.

### Step 3: Check the Results

After sending your message:

1. **Watch the AI's response** — it should describe what it's doing and then show you a preview.
2. **Review the preview** — the table configuration is displayed in the UI before being saved.
3. **Accept or reject** — click Accept if the preview looks correct. The test files describe what you should expect to see.

### Step 4: Verify Success

Each prompt file has a **"What you should see"** section. Use that to check whether the AI did the right thing.

---

## Two Ways to Verify Results

### Option A: Browser UI (Recommended)

The preview pane shows you the table configuration visually. Each test file describes what columns, filters, and sorting should appear in the preview.

### Option B: Browser DevTools — Network Tab (Deeper Verification)

If you want to verify the exact configuration the AI produced (especially for edge cases), you can inspect it in the browser's developer tools. Here's how:

1. Open DevTools: Press **F12** (Windows/Linux) or **Cmd+Option+I** (Mac)
2. Click the **Network** tab at the top
3. Send your chat message
4. Look for a request named **`chat`** in the list (type: `fetch` or `xhr`)
5. Click it, then click **Preview** or **Response** on the right
6. Look for the `table_config_preview` event — the config JSON is inside it

> **Tip**: Filter the network requests by typing "chat" in the filter box to find it quickly.

Each test file includes an optional **"Network tab verification"** section showing exactly what to look for in the JSON.

---

## Prompt Files

| File | What It Tests |
|------|---------------|
| [01-create-basic-table.md](01-create-basic-table.md) | Create a brand-new table with base columns |
| [02-add-columns.md](02-add-columns.md) | Add more columns to an existing table |
| [03-add-join.md](03-add-join.md) | Pull in columns from a related table (e.g., warehouse name) |
| [04-add-filters.md](04-add-filters.md) | Limit which rows are shown (e.g., only active items) |
| [05-add-sort.md](05-add-sort.md) | Control how rows are ordered |
| [06-remove-operations.md](06-remove-operations.md) | Remove columns, filters, or joins |
| [07-complex-table.md](07-complex-table.md) | Build a complete table with everything in one request |
| [complete-walkthrough.md](complete-walkthrough.md) | Full step-by-step test from start to finish |

**Start here if you're new**: Run [complete-walkthrough.md](complete-walkthrough.md) from top to bottom. It covers all the major operations in a logical order.

---

## Suggested Test Tables

These tables are good candidates for testing because they have a variety of column types and relationships:

| Table | What It Contains |
|-------|-----------------|
| `inventory.inventory_items` | Items with quantity, reorder points, and links to warehouses — good for joins and filters |
| `core.users` | User accounts — well-known, has dates, UUIDs, and booleans |
| `sales.orders` | Orders with links to customers and line items |
| `products.products` | Products with brand and category relationships |

---

## Glossary

These terms come up throughout the test files:

| Term | What It Means |
|------|---------------|
| **Base table** | The main table you're querying data from (e.g., `inventory.inventory_items`) |
| **Column** | A field from the table shown as a column in the UI (e.g., `quantity`, `item_number`) |
| **Join** | Pulling in data from a related table (e.g., showing the warehouse name next to each inventory item) |
| **Filter** | A rule that limits which rows are shown (e.g., "only show active items") |
| **Static filter** | A filter with a fixed value baked in (e.g., always filter `is_active = true`) |
| **Dynamic filter** | A filter where the user can enter a value at runtime (e.g., a search box) |
| **Sort** | The order rows appear in the table (e.g., lowest quantity first) |
| **Preview** | The AI shows you the proposed configuration before saving — you must Accept it for it to take effect |
| **Schema** | A namespace for tables in the database (e.g., `inventory`, `core`, `sales`) — think of it as a folder |
| **UUID** | A unique ID used to identify records — looks like `550e8400-e29b-41d4-a716-446655440000` |
| **Hidden column** | A column included in the query for filtering but not displayed to users |
| **config_id** | The unique ID of a saved table configuration |

---

## Key Behaviors to Understand

- **The AI never saves directly** — it always sends a preview first. Nothing changes until you Accept.
- **A preview can fail validation** — if the AI made a mistake, you'll see a validation error instead of a preview. This is expected behavior and worth testing.
- **Changes build on each other** — each message picks up where the previous one left off. The AI loads the current config before making changes.
- **Order of operations matters** — you generally need to add columns before you can filter or sort by them.

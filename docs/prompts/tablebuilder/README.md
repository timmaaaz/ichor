# Table Builder Chat Prompts

This directory contains reusable prompts for testing the table builder AI assistant via chat.

## How the Table Builder Chat Works

The table builder assistant is accessed via `POST /v1/agent/chat` with `context_type: "tables"`. The agent uses a set of operation tools to build and modify `tablebuilder.Config` objects. It never writes to the database directly — it always sends a **preview** for the user to accept or reject in the UI.

### API Request Format

```json
{
  "message": "<your prompt here>",
  "context_type": "tables",
  "context": {
    "config_id": "<uuid>",
    "state": {
      "baseTable": "schema.table",
      "columns": [],
      "joins": [],
      "filters": [],
      "sortBy": []
    }
  }
}
```

- For a **new table** (no existing config): omit `config_id` or set it to an empty string.
- For an **existing table**: include the `config_id` UUID and pass the current `state` from the editor.

### Agent Tool Flow (for every operation)

1. `get_table_config` — fetch current config (auto-filled from `config_id`)
2. `apply_*_change` — apply the operation in memory
3. `preview_table_config` — validate and send SSE preview to frontend
4. User accepts or rejects in the UI

### SSE Events Emitted

| Event | Meaning |
|-------|---------|
| `table_config_preview` | Config is valid, preview sent to user |
| `table_config_validation_error` | Config failed validation |
| `message_complete` | Agent done responding |

---

## Prompt Files

| File | Purpose |
|------|---------|
| [01-create-basic-table.md](01-create-basic-table.md) | Create a new table config with base columns |
| [02-add-columns.md](02-add-columns.md) | Add more columns to an existing config |
| [03-add-join.md](03-add-join.md) | Join a related table and pull columns from it |
| [04-add-filters.md](04-add-filters.md) | Add static and dynamic filters |
| [05-add-sort.md](05-add-sort.md) | Configure sort order |
| [06-remove-operations.md](06-remove-operations.md) | Remove columns and filters |
| [07-complex-table.md](07-complex-table.md) | Full complex table in one request |
| [complete-walkthrough.md](complete-walkthrough.md) | End-to-end walkthrough of ALL steps |

---

## Test Tables

These are good candidate tables to use for testing:

| Table | Schema | Notes |
|-------|--------|-------|
| `inventory_items` | `inventory` | Items with qty, reorder point, FK to warehouses |
| `users` | `core` | Users table, well-known |
| `orders` | `sales` | Orders with FK to customers, line items |
| `products` | `products` | Products with brand, category FKs |

Use `search_database_schema` in the agent to discover exact column names.

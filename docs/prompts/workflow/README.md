# Workflow Chat Prompts

This directory contains reusable prompts for testing the workflow automation AI assistant via chat.

## How the Workflow Chat Works

The workflow assistant is accessed via `POST /v1/agent/chat` with `context_type: "workflow"`. The agent uses a set of tools to discover, read, create, and modify workflow automation rules. It **never writes to the database directly** — it always sends a **preview** for the user to accept or reject in the UI.

### API Request Format

```json
{
  "message": "<your prompt here>",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<uuid>",
    "rule_name": "Optional Rule Name",
    "is_new": false
  }
}
```

- For a **new workflow** (blank canvas): omit `context` or set `is_new: true`.
- For an **existing workflow**: include `workflow_id` UUID and optionally `rule_name`.
- The `context` can also include `nodes` and `edges` arrays from the frontend canvas state.

### Two Creation Modes

| Mode | When | Tools Used |
|------|------|-----------|
| **Draft Builder** (guided, incremental) | New workflows | `start_draft` → `add_draft_action` → `preview_draft` |
| **Full Payload** (direct) | Updates to existing workflows | `preview_workflow` with full workflow JSON |

### Agent Tool Flow

**Creating a new workflow (Draft Builder):**
1. `discover` (category: `action_types`) — learn config schemas and output ports
2. `start_draft` — create in-memory draft with name, entity, trigger
3. `add_draft_action` — add actions one at a time, chained with `after`
4. `preview_draft` — validate and send SSE preview to frontend
5. User accepts → frontend persists the workflow

**Modifying an existing workflow:**
1. `get_workflow_rule` — fetch current workflow summary
2. (optionally) `explain_workflow_node` — drill into specific action config
3. `preview_workflow` — send updated full workflow payload for user approval
4. User accepts → frontend persists the changes

**Reading workflows:**
- `list_workflow_rules` — list all automation rules
- `get_workflow_rule` — get summary of one rule (flow outline, action types)
- `explain_workflow_node` — deep detail on a single action (config, recipients, ports)

### SSE Events Emitted

| Event | Meaning |
|-------|---------|
| `message_start` | Start of an LLM turn |
| `content_chunk` | Streamed text from LLM |
| `tool_call_start` | LLM is invoking a tool |
| `tool_call_result` | Tool call completed |
| `workflow_preview` | Workflow validated, preview sent to user |
| `message_complete` | Agent done responding |
| `error` | Any failure |

---

## Prompt Files

| File | Purpose |
|------|---------|
| [01-discover.md](01-discover.md) | Discover action types, trigger types, and entities |
| [02-list-and-read.md](02-list-and-read.md) | List workflows, read summaries, explain nodes |
| [03-create-simple.md](03-create-simple.md) | Create a new workflow with a single action |
| [04-create-branching.md](04-create-branching.md) | Create a workflow with conditional branching |
| [05-modify-existing.md](05-modify-existing.md) | Modify an existing workflow |
| [06-alerts.md](06-alerts.md) | Work with workflow alerts (inbox, details, rule history) |
| [07-complex-workflow.md](07-complex-workflow.md) | Complex multi-step real-world workflows |
| [complete-walkthrough.md](complete-walkthrough.md) | End-to-end walkthrough covering all operations |

---

## Seed Workflows Available for Testing

After running `make seed` or in test environments, these workflows exist:

| Workflow Name | Trigger | Entity | Description |
|---------------|---------|--------|-------------|
| `Simple Test Workflow` | on_create | (varies by seed) | 1 action: `create_alert` with start edge |
| `Sequence Test Workflow` | on_update | (varies by seed) | 3 sequential `create_alert` actions chained |
| `Branching Test Workflow` | on_create | (varies by seed) | `evaluate_condition` → high value alert OR normal value alert |

Use `list_workflow_rules` in the chat to discover all available rules in your environment.

---

## Available Entities (Use `discover` to get current list)

Common entities workflows can trigger on:

| Entity | Schema.Table | Good for |
|--------|-------------|---------|
| Orders | `sales.orders` | Order processing automation |
| Inventory Items | `inventory.inventory_items` | Stock level alerts |
| Users | `core.users` | User lifecycle events |
| Purchase Orders | `procurement.purchase_orders` | Procurement workflows |

---

## Available Action Types (Use `discover` to get schemas)

| Action Type | Purpose |
|------------|---------|
| `create_alert` | Send an in-app alert to users/roles |
| `send_email` | Send an email via Resend API |
| `send_notification` | Push notification |
| `evaluate_condition` | Branch workflow on true/false condition |
| `check_inventory` | Check current inventory level |
| `check_reorder_point` | Check if stock is below reorder threshold |
| `allocate_inventory` | Reserve inventory |
| `reserve_inventory` | Reserve inventory items |
| `commit_allocation` | Commit a reservation |
| `release_reservation` | Release a previously reserved allocation |
| `seek_approval` | Pause workflow and request human approval |
| `delay` | Wait for a time period |
| `lookup_entity` | Fetch related entity data |
| `create_entity` | Create a new entity record |
| `update_field` | Update a field on an entity |
| `transition_status` | Change entity status |
| `log_audit_entry` | Write an audit log entry |

---

## Available Trigger Types

| Trigger | When It Fires |
|---------|--------------|
| `on_create` | When a new entity record is created |
| `on_update` | When an existing entity record is updated |
| `on_delete` | When an entity record is deleted |
| `scheduled` | On a time-based schedule |

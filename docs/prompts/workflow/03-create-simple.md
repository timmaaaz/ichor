# Step 03: Create a Simple Workflow (Draft Builder)

**Goal**: Create a new workflow with a single action using the draft builder. This exercises `discover`, `start_draft`, `add_draft_action`, and `preview_draft`.

---

## Context Setup

For a new workflow (blank canvas), send `is_new: true` or omit context:

```json
{
  "message": "<prompt>",
  "context_type": "workflow",
  "context": {
    "is_new": true
  }
}
```

---

## Prompt 3A — Create a Single-Action Alert Workflow

Use this to test the simplest possible new workflow:

```
Create a new workflow called "Low Stock Alert" that triggers when inventory items are updated.
When it fires, create an alert with severity "high", title "Low Stock Warning",
and message "An inventory item has been updated — check stock levels."
Send the alert to admin users.
```

**Expected agent behavior:**
1. Calls `discover` with `category: "action_types"` to learn the `create_alert` config schema
2. Calls `start_draft` with `name: "Low Stock Alert"`, `entity: "inventory.inventory_items"`, `trigger_type: "on_update"`
3. Calls `add_draft_action` with:
   - `name: "Alert - Low Stock"`
   - `action_type: "create_alert"`
   - `action_config` with severity, title, message, recipients
   - No `after` field (first action = start node)
4. Calls `preview_draft` → server emits `workflow_preview` SSE event
5. Tells user "Preview is ready for your review"

**What to verify in the preview:**
- Rule name: "Low Stock Alert"
- Entity: `inventory.inventory_items`
- Trigger: `on_update`
- 1 action of type `create_alert`
- 1 start edge: nil → action
- Alert severity: `high`
- Alert title: "Low Stock Warning"

---

## Prompt 3B — Agent-Guided Creation (Conversational)

Use this to test the agent's guided creation conversation:

```
I want to set up an automation for when new orders come in.
```

**Expected agent behavior (multi-turn):**
1. Asks clarifying questions: what entity? what should happen?
2. Suggests a plan: "Trigger on `sales.orders` on_create, send a notification?"
3. After user confirms, calls `discover` → `start_draft` → `add_draft_action` → `preview_draft`

**Follow-up prompt (after agent responds with a plan):**

```
Yes, that sounds right. When a new order is created, send an email to the orders team.
```

**What to verify:**
- Agent does NOT immediately call tools on first message — it asks first
- Agent proposes entity (`sales.orders`) and trigger (`on_create`) based on context
- Final preview has `send_email` action type with reasonable config
- Agent describes what was built after preview is sent

---

## Prompt 3C — Email Notification on Create

Use this to test `send_email` action type:

```
Build a workflow called "New Order Email" on sales.orders triggered when an order is created.
Add a send_email action that sends to "orders@company.com" with subject "New Order Received"
and body "A new order has been placed in the system."
```

**Expected agent behavior:**
1. Calls `discover` with `category: "action_types"` for the `send_email` schema
2. Calls `start_draft` for `sales.orders` / `on_create`
3. Calls `add_draft_action` with:
   - `action_type: "send_email"`
   - `action_config` with recipients array, subject, body
4. Calls `preview_draft`

**What to verify in the preview:**
- Entity: `sales.orders`
- Trigger: `on_create`
- Action type: `send_email`
- Recipients include "orders@company.com"
- Subject and body match the prompt

---

## Prompt 3D — Workflow with a Delay Action

Use this to test the `delay` action type:

```
Create a workflow "Order Follow-Up" on sales.orders on_create.
First add a delay of 24 hours, then send an alert with title "Follow-Up Needed"
and message "This order was placed 24 hours ago and may need follow-up."
```

**Expected agent behavior:**
1. `discover` → `start_draft`
2. `add_draft_action` for delay (no `after` — first action)
3. `add_draft_action` for create_alert with `after: "Delay 24h"` (or whatever the delay was named)
4. `preview_draft`

**What to verify in the preview:**
- 2 actions: delay → create_alert
- 2 edges: start edge to delay, sequence edge from delay to alert
- `after` correctly chains the actions

---

## Validation Checkpoints

After accepting the preview, the workflow should have this structure:

```json
{
  "name": "Low Stock Alert",
  "is_active": true,
  "entity": "inventory.inventory_items",
  "trigger_type": "on_update",
  "actions": [
    {
      "name": "Alert - Low Stock",
      "action_type": "create_alert",
      "action_config": {
        "alert_type": "low_stock",
        "severity": "high",
        "title": "Low Stock Warning",
        "message": "An inventory item has been updated — check stock levels.",
        "recipients": { "users": [...], "roles": [] }
      },
      "is_active": true
    }
  ],
  "edges": [
    {
      "edge_type": "start",
      "target_action_id": "<action-id>"
    }
  ]
}
```

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Agent skips `discover` and invents config | Missing discovery step | Agent should always `discover action_types` before creating actions |
| `start_draft` fails with entity not found | Wrong entity name format | Entity must be `schema.table` format (e.g., `inventory.inventory_items`) |
| No `after` field on first action | Correct — first action should omit `after` | This is correct behavior; start edge is auto-generated |
| Agent calls `preview_workflow` instead of `preview_draft` | Using wrong preview tool | For draft-based creation, use `preview_draft` |
| `action_config` missing required fields | Wrong config for action type | Use `discover action_types` result to get the correct schema |

---

## Notes

- **Draft TTL is 10 minutes.** If the conversation takes too long, the draft may expire. Start over with `start_draft`.
- The `after` field uses the **action name** (e.g., `"after": "Evaluate Stock"`), not an ID. Names must match exactly.
- Entity names resolve automatically — you can use `"inventory.inventory_items"` or the UUID. Names are preferred for readability.
- After `preview_draft` fires, the agent should tell the user "the preview is ready" — it should NOT call any other tools.

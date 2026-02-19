# 03 — Create a Simple Workflow

**Goal**: Build a new workflow using the chatbot. This covers the guided creation flow: the chatbot asks questions or follows your instructions, builds the workflow step by step, and sends a preview for you to review.

---

## Setup

For a new workflow, tell the chatbot it's a blank canvas:

```json
{
  "message": "your prompt here",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

---

## Prompts to Try

### 3A — Single alert on inventory update

```
Create a new workflow called "Low Stock Alert" that triggers when inventory items are updated.
When it fires, create an alert with severity "high", title "Low Stock Warning",
and message "An inventory item has been updated — check stock levels."
```

**What to check:**
- Chatbot calls a few tools behind the scenes (you may see "checking action types", "starting draft", etc.)
- Chatbot summarizes what it's building before the preview
- A **workflow preview card** appears in the UI
- The preview shows: name "Low Stock Alert", triggers on inventory item updates, 1 action

---

### 3B — Let the chatbot guide you (conversational)

```
I want to set up an automation for when new orders come in.
```

**What to check:**
- Chatbot should **ask clarifying questions** rather than immediately building something
- It should ask or suggest: what entity? what should happen when an order is created?
- After you answer, it builds step by step

**Follow-up prompt:**
```
Yes — when a new order is created, send an email to the orders team at orders@company.com
with subject "New Order Received".
```

**What to check after follow-up:**
- Chatbot builds the workflow and sends a preview
- Preview shows: sales.orders, on_create trigger, send_email action

---

### 3C — Email notification on new record

```
Build a workflow called "New Order Email" that fires when an order is created in sales.orders.
Send an email to orders@company.com with subject "New Order Received"
and body "A new order has been placed in the system."
```

**What to check:**
- Preview appears with the right trigger (on_create on orders)
- Action type is send_email
- Email goes to the right address

---

### 3D — Workflow with a delay

```
Create a workflow "Order Follow-Up" on sales.orders triggered on_create.
Wait 24 hours, then create an alert with title "Follow-Up Needed"
and message "This order was placed 24 hours ago and may need follow-up."
```

**What to check:**
- Preview shows 2 steps: a delay action, then a create_alert action
- Steps are chained in order (delay fires first, alert fires after)

---

## What a Good Response Looks Like

After you submit a prompt, the chatbot should:

1. Briefly explain what it's going to build (or ask if requirements are unclear)
2. Make a few tool calls in the background
3. Send a preview card to the UI
4. Say something like "The preview is ready — please accept or reject it"

It should **not**:
- Immediately save anything without showing a preview first
- Ask a dozen clarifying questions for a clearly-specified request
- Dump raw JSON as a response

---

## Common Issues

| Problem | What to look for |
|---------|-----------------|
| No preview card appears | Chatbot may have hit a validation error — check if it explains the issue |
| Chatbot asks too many questions | For detailed prompts like 3A and 3C, it should just build without interrogating |
| Preview shows wrong entity or trigger | Chatbot may have misunderstood — try being more explicit (e.g., "inventory.inventory_items") |
| Draft expires mid-conversation | Chatbot will say "draft not found" — start fresh with a new message |

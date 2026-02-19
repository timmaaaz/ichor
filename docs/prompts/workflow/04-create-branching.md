# 04 — Create a Branching Workflow

**Goal**: Build a workflow that takes different paths depending on a condition — "if X is true, do this; otherwise do that."

---

## Setup

```json
{
  "message": "your prompt here",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

---

## Prompts to Try

### 4A — Basic true/false branch

```
Create a workflow called "Order Value Check" on sales.orders triggered on_create.
Check if the order amount is greater than 1000.
If it is (high value), create an alert with severity "high" titled "High Value Order".
If it isn't (normal value), create an alert with severity "low" titled "Standard Order".
```

**What to check:**
- Preview shows 3 steps: a condition check, then two separate alert actions
- The condition evaluates amount > 1000
- Each path leads to a different alert
- Both paths are visible in the preview graph

---

### 4B — Inventory reorder check

```
Build a workflow "Reorder Check" on inventory.inventory_items triggered on_update.
Check if the item is below the reorder point.
If it is below the threshold, send an alert with title "Reorder Required" with severity "high".
If it's fine, log an audit entry noting the stock is OK.
```

**What to check:**
- Preview shows: check action → two branches (alert or log)
- Chatbot correctly identifies the check action type and both output paths
- Both branches are present in the preview

---

### 4C — Three-step branch

```
Create "Procurement Approval" on procurement.purchase_orders on_create.
First, check if the order total is above 5000.
If yes: request approval from the procurement team, then send an email confirming it was approved.
If no: send a confirmation email directly saying it was auto-approved.
```

**What to check:**
- Preview shows 4 steps total
- High-value path: condition → seek approval → send email
- Low-value path: condition → send email (directly)
- Both paths end with a send_email step

---

## What Branching Looks Like in the Preview

The preview graph shows:
- A diamond or node for the condition step
- Two edges leaving it (one for each outcome)
- Each edge leads to its own next step

If the graph only shows one path, the chatbot may have missed a branch — let it know:

```
The preview only shows the high-value path. Can you also add the standard alert for when the amount is 1000 or less?
```

---

## Common Issues

| Problem | What to look for |
|---------|-----------------|
| Only one branch appears in the preview | Chatbot may have only added one `after` step — follow up to add the other |
| Chatbot asks which port to use | It should know automatically from the action type; if unsure, tell it "output-true for high value, output-false for normal" |
| Condition check seems backwards | Clarify: "check_reorder_point succeeds when stock IS below the threshold, meaning a reorder is needed" |
| Preview shows all actions but no branching connections | Edges may be missing — ask "can you make sure the two alert actions are connected to the correct outputs of the condition?" |

# 07 — Complex Workflows

**Goal**: Test the chatbot with multi-step, real-world automation scenarios. These combine branching, sequential steps, and multiple action types.

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

## Scenarios

### 7A — Inventory reorder pipeline

```
Build a workflow called "Inventory Reorder Pipeline" on inventory.inventory_items
triggered when items are updated.

Steps:
1. Check if the item is below the reorder point
2. If it is below threshold: reserve the inventory, then create an alert "Reorder Triggered" with severity "high"
3. If it's above threshold: log an audit entry "Stock level acceptable"
```

**What to check:**
- Preview shows 4 actions
- Two distinct branches from the check step
- Low-stock path has 2 steps after the check (reserve → alert)
- Fine-stock path has 1 step (log)

---

### 7B — Purchase order approval

```
Create "Purchase Order Approval" on procurement.purchase_orders triggered on_create.

Steps:
1. Check if the total is over 10,000
2. If yes: request approval from the procurement manager, then send an email "PO Approved" once approved
3. If no: send a confirmation email "PO Auto-Approved — under threshold"
```

**What to check:**
- Preview shows 4 actions
- High-value path: condition → seek approval → send email (3 steps)
- Low-value path: condition → send email (2 steps)
- Both `send_email` actions have different subjects per the prompt

---

### 7C — New user onboarding sequence

```
Build "New User Onboarding" on core.users triggered on_create.

Sequence:
1. Send a welcome email immediately
2. Wait 24 hours
3. Send a follow-up email with onboarding resources
4. Create an alert for HR: "New user onboarded — verify profile"
```

**What to check:**
- Preview shows 4 actions in a straight line (no branching)
- Order is correct: email → delay → email → alert
- No actions are skipped or reordered

---

### 7D — One-shot complex request

Test whether the chatbot handles a fully-specified request without needing back-and-forth:

```
Build a complete workflow called "Sales Order Fulfillment" on sales.orders triggered on_update.

Steps:
1. Check if status changed to "fulfillment_ready"
2. If yes: allocate inventory, then send email "Fulfillment Started" to the warehouse team
3. If allocation fails: create a critical alert "Inventory Shortage" for operations
4. If status not ready: log an audit entry "Order status update received"
```

**What to check:**
- Chatbot does NOT ask clarifying questions (the prompt is fully specified)
- Builds everything in one go and sends a preview
- Preview shows all 5+ actions with correct connections
- Chatbot summarizes what was built before or after the preview appears

---

## What to Watch For

**Good signs:**
- Chatbot says what it's going to build before calling tools
- After adding several steps, it summarizes: "I've added 5 actions: [list]. Does this match what you wanted?"
- Preview card appears with a meaningful description
- Complex branching is clearly visible in the preview graph

**Warning signs:**
- Chatbot gets stuck asking questions on a fully-specified request
- Preview only shows some of the actions (steps may be missing)
- Chatbot confuses which path is "success" vs "failure" for inventory actions — it's OK to clarify

**If something's off**, try:
```
The preview is missing the allocation failure path. Can you add a critical alert for when the inventory allocation fails?
```

---

## Common Issues

| Problem | What to look for |
|---------|-----------------|
| Chatbot breaks a complex request into too many back-and-forth turns | For fully specified prompts, it should build in one shot |
| Preview is missing a branch | Follow up asking for the missing path specifically |
| Actions appear but aren't connected properly | Ask: "Can you make sure all the steps are connected in the right order?" |
| Draft expires mid-build (10 min timeout) | Chatbot will report "draft not found" — start fresh |

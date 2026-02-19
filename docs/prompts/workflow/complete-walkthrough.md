# Complete Workflow Chat Walkthrough

End-to-end test covering all major workflow chat features in order. Run these phases in sequence — each builds on context from the previous steps.

**Time**: ~30 minutes
**Requires**: Dev environment running with seed data loaded (`make seed`)

---

## Phase 1 — Discover what's available

**Request:**
```json
{
  "message": "What workflow action types are available? Give me a quick overview.",
  "context_type": "workflow"
}
```

**Pass when:**
- [ ] Response lists action types (create_alert, send_email, evaluate_condition, delay, seek_approval, etc.)
- [ ] Each action is described in plain language
- [ ] No raw JSON dumped as the response

---

## Phase 2 — List existing workflows

**Request:**
```json
{
  "message": "Show me all the workflow rules in the system.",
  "context_type": "workflow"
}
```

**Pass when:**
- [ ] Response includes all 3 seeded workflows: Simple Test Workflow, Sequence Test Workflow, Branching Test Workflow
- [ ] Each rule shows its trigger type and active status
- [ ] **Save the UUID of "Branching Test Workflow"** — you'll need it in phases 4 and 5

---

## Phase 3 — Understand a workflow

**Request:**
```json
{
  "message": "Explain what the Branching Test Workflow does. Walk me through its flow.",
  "context_type": "workflow"
}
```

**Pass when:**
- [ ] Response describes: what triggers it, what it evaluates, and what happens on each path
- [ ] Explains the branching logic in plain language (not just "there are 3 nodes")
- [ ] Does not make up information

---

## Phase 4 — Drill into an action

Use the UUID from Phase 2.

**Request:**
```json
{
  "message": "In the Branching Test Workflow, what exactly does the Evaluate Amount action check?",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<UUID-of-Branching-Test-Workflow>",
    "rule_name": "Branching Test Workflow"
  }
}
```

**Pass when:**
- [ ] Chatbot explains the condition: checks if amount is greater than 1000
- [ ] Describes both outcomes: high-value path vs. normal-value path
- [ ] Does not re-fetch the whole workflow (it's already in context)

---

## Phase 5 — Check alert recipients

Using the same workflow context as Phase 4.

**Request:**
```json
{
  "message": "Who is set up to receive the alert on the high-value path?",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<UUID-of-Branching-Test-Workflow>",
    "rule_name": "Branching Test Workflow"
  }
}
```

**Pass when:**
- [ ] Response shows recipient names or emails — not UUID strings
- [ ] These are the **configured** recipients, not a history of who has been notified
- [ ] Chatbot does not go off and list alerts from your inbox

---

## Phase 6 — Check your alert inbox

**Request:**
```json
{
  "message": "Do I have any active alerts in my inbox?",
  "context_type": "workflow"
}
```

**Pass when:**
- [ ] Response clearly states what's in your inbox (or says it's empty)
- [ ] Alerts show title and severity
- [ ] Response is about YOUR inbox only (not all system alerts)

---

## Phase 7 — Create a simple workflow

**Request:**
```json
{
  "message": "Create a new workflow called 'Low Stock Alert' on inventory.inventory_items triggered when items are updated. Add a single alert action with severity high, title 'Low Stock Warning', message 'An item has been updated — check stock levels.'",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

**Pass when:**
- [ ] Chatbot makes tool calls in the background (you see activity before the response)
- [ ] **A preview card appears in the UI**
- [ ] Preview shows: "Low Stock Alert", inventory items, on_update trigger, 1 action
- [ ] Chatbot says something like "Preview is ready for your review"
- [ ] Chatbot does NOT say it has saved the workflow (it hasn't — you need to accept first)

---

## Phase 8 — Create a branching workflow

**Request:**
```json
{
  "message": "Build a new workflow called 'Order Value Check' on sales.orders on_create. Check if amount > 1000. If true, create a high-severity alert 'High Value Order'. If false, create a low-severity alert 'Standard Order'.",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

**Pass when:**
- [ ] **Preview card appears**
- [ ] Preview shows 3 steps: condition check + two alert actions
- [ ] The two alert actions are on separate branches (visible in the preview graph)
- [ ] Both alerts have different titles and severities as specified

---

## Phase 9 — Modify an existing workflow

Use the UUID of "Simple Test Workflow" from Phase 2.

**Request:**
```json
{
  "message": "Change the alert severity in the 'Create Alert' action to 'critical'.",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<UUID-of-Simple-Test-Workflow>",
    "rule_name": "Simple Test Workflow"
  }
}
```

**Pass when:**
- [ ] **Preview card appears**
- [ ] Preview description mentions the severity change
- [ ] The workflow still has its original name, trigger, and entity
- [ ] Only the severity changed — nothing else was altered

---

## Phase 10 — Remove a draft action mid-build

This is a two-turn test.

**Turn 1:**
```json
{
  "message": "Start building a new workflow called 'Test Draft' on core.users on_create. Add a delay action first, then a create_alert action.",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

Wait for the chatbot to confirm it added both actions.

**Turn 2** (continue the same conversation):
```
Actually, remove the delay action. Just do the alert directly without the wait.
```

**Pass when:**
- [ ] Chatbot confirms the delay was removed
- [ ] Preview shows only 1 action (the alert)
- [ ] The alert is now the first (and only) step

---

## Phase 11 — Check if a workflow has fired

**Request:**
```json
{
  "message": "Has the Simple Test Workflow triggered any alerts? How many?",
  "context_type": "workflow"
}
```

**Pass when:**
- [ ] Chatbot searches rule alert history (not your personal inbox)
- [ ] Reports a count — even if zero, says "No alerts have been fired by this rule"
- [ ] Does not confuse rule alert history with your personal inbox

---

## Overall Pass Checklist

### Behavior
- [ ] Chatbot never claims to save something — always sends a preview first
- [ ] Chatbot explains what it's doing before making tool calls (or summarizes after)
- [ ] Responses are in plain language — no raw JSON walls
- [ ] When workflow context is provided, chatbot doesn't re-fetch it unnecessarily

### Previews
- [ ] Preview card appeared for every create/update operation (phases 7, 8, 9, 10)
- [ ] Each preview had a meaningful description of what changed
- [ ] Preview content matched what was asked for

### Accuracy
- [ ] Names, trigger types, and entities matched what was specified in prompts
- [ ] Recipients shown as names/emails, never raw UUIDs
- [ ] Alert inbox (yours) vs. rule alert history (everyone's) correctly distinguished

---

## Quick Reference

| What you want | Prompt |
|--------------|--------|
| See available building blocks | "What action types are available?" |
| List all rules | "Show me all workflow rules" |
| Understand a rule | "Explain the [Name] workflow" |
| See action details | "What does the [Action Name] action do?" |
| See configured recipients | "Who receives the alert from [Action Name]?" |
| Your inbox | "Do I have any active alerts?" |
| Rule fire history | "Has [Rule Name] sent any alerts?" |
| New workflow | "Create a workflow called X on Y triggered when Z" |
| Add branching | "If [condition], do A. Otherwise do B." |
| Edit existing | "Change [thing] in the [Action Name] action" |
| Pause a rule | "Deactivate the [Rule Name] workflow" |

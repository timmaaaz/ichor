# 02 — List and Read Existing Workflows

**Goal**: Ask the chatbot about workflows that already exist. Nothing gets changed — these are all read-only.

---

## Setup

No context needed for listing. For questions about a specific open workflow, include its ID:

```json
{
  "message": "your prompt here",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<uuid>",
    "rule_name": "Branching Test Workflow"
  }
}
```

---

## Prompts to Try

### 2A — List everything

```
Show me all the workflow rules in the system.
```

**What to check:**
- Response includes the 3 seeded workflows: Simple Test Workflow, Sequence Test Workflow, Branching Test Workflow
- Each shows: name, what entity it watches, what triggers it, whether it's active
- Response is readable — not a raw data dump

---

### 2B — Get a workflow summary

```
Explain what the Branching Test Workflow does. Walk me through its flow.
```

**What to check:**
- Chatbot describes the trigger (on_create) and what entity it watches
- Explains the structure: there's a condition check, then two paths — one for high values, one for normal values
- Doesn't just say "there are 3 actions and 3 edges" — actually explains the logic in words

---

### 2C — Drill into a specific action

```
In the Branching Test Workflow, what exactly does the "Evaluate Amount" action check?
```

**What to check:**
- Chatbot explains: checks if the amount field is greater than 1000
- Explains in plain language ("if the amount is over 1000, it takes the high-value path")
- If there are two paths, names both of them

---

### 2D — Ask about alert recipients

```
In the Branching Test Workflow, who receives the alert when it goes down the high-value path?
```

**What to check:**
- Chatbot shows recipient names and/or emails — not raw IDs
- These are the **configured** recipients (who the workflow is set up to notify), not a history of who has been notified
- If there are role-based recipients, the role name is shown

---

### 2E — Compare two workflows

```
How is the Simple Test Workflow different from the Sequence Test Workflow?
```

**What to check:**
- Chatbot correctly describes both:
  - Simple: 1 action
  - Sequence: 3 actions that run one after another
- Explains the difference conversationally

---

### 2F — Find workflows for a domain

```
Are there any workflow rules that watch inventory items?
```

**What to check:**
- Chatbot lists any inventory-related rules (or says there are none if the seeded data doesn't include any)
- Doesn't confuse entity names — inventory items are in the `inventory` schema

---

## Common Issues

| Problem | What to look for |
|---------|-----------------|
| Chatbot re-fetches a workflow that's already open | If workflow_id is in context, it shouldn't make an extra API call to re-fetch it |
| Recipients shown as long UUID strings | Should resolve to names/emails |
| Chatbot answers "who receives alerts" by showing fired alert history | It should explain **configured** recipients, not past alert events |

# 05 — Modify an Existing Workflow

**Goal**: Make changes to a workflow that already exists. The chatbot reads the current state, makes your requested change, and sends a preview — nothing is saved until you accept.

---

## Setup

Open the workflow you want to edit by passing its ID and name in context:

```json
{
  "message": "your prompt here",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<uuid-of-the-workflow>",
    "rule_name": "Branching Test Workflow"
  }
}
```

Get the UUID from the list you saw in step 02, or from `list_workflow_rules`.

---

## Prompts to Try

### 5A — Change a config value

```
Change the alert severity in the "High Value Alert" action from "high" to "critical".
```

**What to check:**
- A preview card appears
- Preview description says something like "Changed alert severity to critical"
- The rest of the workflow (other actions, trigger, entity) is unchanged

---

### 5B — Add a new action to the end

```
Add a log_audit_entry action at the end of the "Sequence Test Workflow" after the last alert.
The message should say "Sequence workflow completed successfully."
```

**What to check:**
- Preview shows 4 actions (original 3 + new log action)
- New action is chained after the last existing action
- The 3 original actions are still there, unchanged

---

### 5C — Remove an action

```
Remove the "Normal Value Alert" branch from the "Branching Test Workflow".
The workflow should only alert on high-value orders now.
```

**What to check:**
- Preview shows 2 actions (condition + high value alert only)
- The false-branch path and alert are gone
- The high-value path still works

---

### 5D — Pause a workflow

```
Deactivate the "Simple Test Workflow" — I want to pause it without deleting it.
```

**What to check:**
- Preview shows the workflow with active status set to off/false
- All actions and structure are unchanged
- Chatbot explains that the workflow won't fire while deactivated

---

### 5E — Rename a workflow

```
Rename the "Sequence Test Workflow" to "Order Processing Pipeline".
```

**What to check:**
- Preview shows the new name
- Everything else (actions, trigger, entity) is unchanged

---

### 5F — Update who receives an alert

```
In the "Branching Test Workflow", add the "managers" role to the recipients of the "High Value Alert".
```

**What to check:**
- Chatbot may ask for clarification on which role (if there's ambiguity)
- Preview shows the updated alert action with the managers role added
- Other recipients are not removed

---

## What a Good Modify Flow Looks Like

1. You send the change request with workflow context
2. Chatbot reads the existing workflow from context (should **not** re-fetch it separately)
3. For detailed config changes, chatbot may look up the specific action first
4. Preview card appears with a clear description of what changed
5. You accept → change is saved

---

## Common Issues

| Problem | What to look for |
|---------|-----------------|
| Preview wipes out existing actions | Chatbot must include all existing actions in the update, not just the changed one |
| Preview shows wrong workflow | Check that `workflow_id` in context matches the workflow you're editing |
| Chatbot tries to re-create instead of update | Should be using `preview_workflow` (update path), not the draft builder tools |
| Chatbot makes the change but loses the edges | Edges between actions should be preserved when only one action is modified |

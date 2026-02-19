# Step 02: List and Read Existing Workflows

**Goal**: List all workflows in the system, get a summary of a specific rule, and drill into action details. These are read-only operations — no changes are made.

---

## Context Setup

For listing all workflows, no context is needed:

```json
{
  "message": "<prompt>",
  "context_type": "workflow"
}
```

For reading a specific workflow already open in the UI:

```json
{
  "message": "<prompt>",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<uuid>",
    "rule_name": "Simple Test Workflow"
  }
}
```

---

## Prompt 2A — List All Workflows

Use this to see everything in the system:

```
Show me all the workflow rules. How many are there and what do they do?
```

**Expected agent behavior:**
1. Calls `list_workflow_rules`
2. Returns a list of rules with name, entity, trigger type, and active status
3. Summarizes the list in plain language

**What to verify:**
- Agent uses `list_workflow_rules` (not `get_workflow_rule`)
- Response includes rule names, entity, trigger types, and active status
- Response mentions total count
- Does NOT dump raw JSON — presents it clearly

---

## Prompt 2B — Get a Specific Workflow Summary

Use this to understand one workflow's structure:

```
Tell me about the "Branching Test Workflow". What does it do and how is it structured?
```

**Expected agent behavior:**
1. Calls `get_workflow_rule` with `workflow_id: "Branching Test Workflow"` (name resolves to UUID automatically)
2. Returns: trigger, entity, action count, edge count, flow outline
3. Presents a human-readable summary of the flow

**What to verify:**
- Agent calls `get_workflow_rule` (not `explain_workflow_node` or `list_workflow_rules`)
- Response includes: trigger type (on_create), entity, number of actions
- Response shows the flow outline: Start → Evaluate Amount → High Value Alert OR Normal Value Alert
- Response does NOT return raw action configs (that's what `explain_workflow_node` is for)

---

## Prompt 2C — Explain a Specific Action Node

Use this to drill into one action's configuration:

```
In the "Branching Test Workflow", what does the "Evaluate Amount" action check? What's the condition?
```

**Expected agent behavior:**
1. Calls `explain_workflow_node` with `node_name: "Evaluate Amount"` (workflow_id auto-filled from context or resolved by name)
2. Returns the full action config including conditions array
3. Explains the condition in plain language: "checks if amount > 1000"

**What to verify:**
- Agent uses `explain_workflow_node` (not `get_workflow_rule`)
- Response shows the condition: `field: "amount"`, `operator: "greater_than"`, `value: 1000`
- Agent explains it in plain language rather than dumping raw JSON

---

## Prompt 2D — Who Receives Alerts?

Use this to find configured alert recipients:

```
In the "Simple Test Workflow", who receives the alert when it fires?
```

**Expected agent behavior:**
1. Calls `explain_workflow_node` with `node_name: "Create Alert"` (the alert action)
2. Returns recipients with resolved names and emails (not raw UUIDs)
3. Lists recipients by name

**What to verify:**
- Agent uses `explain_workflow_node`, NOT `list_my_alerts` or `list_alerts_for_rule`
- Recipients are shown as names/emails, not UUIDs
- If only one node in the workflow, agent may infer the node name from context

---

## Prompt 2E — Compare Two Workflows

Use this to compare structure across rules:

```
How does the "Simple Test Workflow" differ from the "Sequence Test Workflow"? Which one is more complex?
```

**Expected agent behavior:**
1. Calls `get_workflow_rule` for "Simple Test Workflow"
2. Calls `get_workflow_rule` for "Sequence Test Workflow"
3. Compares them: action count, trigger types, flow structure
4. Explains the difference clearly

**What to verify:**
- Agent makes TWO `get_workflow_rule` calls (one per workflow)
- Simple: 1 action, Sequence: 3 sequential actions
- Agent correctly identifies which is more complex

---

## Prompt 2F — Find Workflows for a Specific Entity

Use this when you want to know all rules watching a given entity:

```
Are there any workflow rules that trigger on inventory items? What do they do when inventory changes?
```

**Expected agent behavior:**
1. Calls `list_workflow_rules`
2. Filters the results for inventory-related entities
3. Describes what each rule does when inventory items change

**What to verify:**
- Agent uses `list_workflow_rules` to find rules
- Response correctly filters for inventory-related entities
- If no inventory rules exist in the seed data, agent says so clearly

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Agent re-fetches workflow already in context | Ignoring `context.workflow_id` | Context should prevent redundant `get_workflow_rule` calls when workflow is already provided |
| Agent calls `list_my_alerts` for recipient question | Confusing alert inbox with alert config | `explain_workflow_node` is for configured recipients; `list_my_alerts` is for fired alerts |
| Agent shows raw UUIDs for recipients | Skipped `explain_workflow_node` enrichment | Recipients should always be resolved to names/emails |
| Agent dumps raw JSON instead of explaining | Bad response formatting | Agent should extract key facts and explain in plain language |

---

## Notes

- `get_workflow_rule` returns a **summary** with flow outline — it does NOT include raw action configs.
- `explain_workflow_node` returns **full config** for one action — use this when you need to see the actual settings.
- The `workflow_id` parameter accepts either a UUID or a rule name — names are resolved automatically.
- When a `workflow_id` is set in the context, the agent should NOT call `get_workflow_rule` to re-fetch it — the context already contains the full workflow state.

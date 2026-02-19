# Complete Workflow Chat Walkthrough

This is a sequenced end-to-end test that covers all workflow chat operations in order. Follow the phases in sequence — each builds on the previous.

**Estimated completion**: ~30 minutes
**Tests**: 11 phases covering all tools and patterns

---

## Setup

Use the Ichor development environment with seed data loaded (`make seed`).
All prompts use `POST /v1/agent/chat` with the listed request body.

---

## Phase 1: Discovery

**Goal**: Verify the agent can discover all building blocks.

**Request:**
```json
{
  "message": "What workflow action types are available? Give me a quick overview.",
  "context_type": "workflow"
}
```

**Pass criteria:**
- [ ] Agent calls `discover` with `category: "action_types"`
- [ ] Response lists all 17+ action types
- [ ] Agent explains them in plain language (no raw JSON dump)
- [ ] SSE events include `tool_call_start` and `tool_call_result` for `discover`

---

## Phase 2: List Existing Workflows

**Goal**: Verify the agent can enumerate the seeded workflow rules.

**Request:**
```json
{
  "message": "Show me all the workflow rules in the system.",
  "context_type": "workflow"
}
```

**Pass criteria:**
- [ ] Agent calls `list_workflow_rules`
- [ ] Response includes the 3 seeded workflows: "Simple Test Workflow", "Sequence Test Workflow", "Branching Test Workflow"
- [ ] Each rule shows: name, entity, trigger type, active status
- [ ] Total count mentioned

---

## Phase 3: Read a Workflow Summary

**Goal**: Verify the agent can explain a specific workflow's structure.

**Request:**
```json
{
  "message": "Explain what the Branching Test Workflow does. Walk me through its flow.",
  "context_type": "workflow"
}
```

**Pass criteria:**
- [ ] Agent calls `get_workflow_rule` with "Branching Test Workflow" (name resolves to UUID)
- [ ] Response describes: trigger (on_create), entity, 3 actions
- [ ] Flow outline shows the branching structure: Evaluate Amount → High Value OR Normal Value
- [ ] Agent does NOT dump raw JSON

---

## Phase 4: Explain a Node

**Goal**: Verify the agent can drill into action-level config.

**Request:**
```json
{
  "message": "In the Branching Test Workflow, what condition does 'Evaluate Amount' check?",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<UUID-of-Branching-Test-Workflow>"
  }
}
```

*(Get the UUID from the Phase 2 list response)*

**Pass criteria:**
- [ ] Agent calls `explain_workflow_node` with `node_name: "Evaluate Amount"` (workflow_id auto-injected)
- [ ] Response shows the condition: field=amount, operator=greater_than, value=1000
- [ ] Agent does NOT re-call `get_workflow_rule` (already in context)
- [ ] Recipients shown as names (not UUIDs)

---

## Phase 5: Check Alert Recipients

**Goal**: Verify the agent correctly distinguishes configured recipients from alert history.

**Request:**
```json
{
  "message": "Who is set up to receive alerts from the High Value Alert action?",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<UUID-of-Branching-Test-Workflow>"
  }
}
```

**Pass criteria:**
- [ ] Agent calls `explain_workflow_node` with `node_name: "High Value Alert"`
- [ ] Response shows configured recipients (users/roles) — NOT fired alert history
- [ ] Agent does NOT call `list_my_alerts` or `list_alerts_for_rule`
- [ ] Recipients resolved to names/emails

---

## Phase 6: Check Alert Inbox

**Goal**: Verify the agent can access the current user's alert inbox.

**Request:**
```json
{
  "message": "Do I have any active alerts in my inbox?",
  "context_type": "workflow"
}
```

**Pass criteria:**
- [ ] Agent calls `list_my_alerts` with `status: "active"` (or no status, which defaults to active)
- [ ] Response clearly indicates what's in the inbox
- [ ] If empty: says "You have no active alerts"
- [ ] If not empty: shows count, titles, severities

---

## Phase 7: Create a Simple Workflow

**Goal**: Test the full draft builder flow for a single-action workflow.

**Request:**
```json
{
  "message": "Create a new workflow called 'Low Stock Alert' on inventory.inventory_items triggered when items are updated. Add a single create_alert action with severity high, title 'Low Stock Warning', message 'An item has been updated — check stock levels.'",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

**Pass criteria:**
- [ ] Agent calls `discover` with `action_types`
- [ ] Agent calls `start_draft` with entity `inventory.inventory_items`, trigger `on_update`
- [ ] Agent calls `add_draft_action` with correct `create_alert` config
- [ ] Agent calls `preview_draft` with a description
- [ ] SSE emits `workflow_preview` event
- [ ] Preview shows: name, entity, trigger, 1 action, 1 start edge
- [ ] Agent announces "Preview is ready"

---

## Phase 8: Create a Branching Workflow

**Goal**: Test the draft builder with evaluate_condition branching.

**Request:**
```json
{
  "message": "Build a new workflow called 'Order Value Check' on sales.orders on_create. Check if amount > 1000. If true, create a high-severity alert 'High Value Order'. If false, create a low-severity alert 'Standard Order'.",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

**Pass criteria:**
- [ ] Agent calls `discover` (learns `output-true`/`output-false` ports)
- [ ] Agent calls `start_draft`
- [ ] Agent adds 3 actions: evaluate_condition, high alert, normal alert
- [ ] `after` fields use correct port syntax: `"Condition:output-true"` and `"Condition:output-false"`
- [ ] Preview shows 3 actions + 3 edges (start + 2 sequence)
- [ ] Both branch edges have `source_output` set correctly

---

## Phase 9: Modify an Existing Workflow

**Goal**: Test the preview_workflow update path.

First, use the "Simple Test Workflow" UUID from Phase 2.

**Request:**
```json
{
  "message": "Change the alert severity in the 'Create Alert' action of the Simple Test Workflow to 'critical'.",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<UUID-of-Simple-Test-Workflow>",
    "rule_name": "Simple Test Workflow"
  }
}
```

**Pass criteria:**
- [ ] Agent does NOT call `get_workflow_rule` (already in context)
- [ ] Agent optionally calls `explain_workflow_node` to confirm current config
- [ ] Agent calls `preview_workflow` with the full updated payload and `workflow_id`
- [ ] Preview shows `severity: "critical"` on the alert action
- [ ] All other fields are unchanged (name, entity, trigger, edges)
- [ ] `description` field describes what changed

---

## Phase 10: Remove a Draft Action

**Goal**: Test `remove_draft_action` mid-conversation.

Start a new draft conversation:

**Turn 1:**
```json
{
  "message": "Start building a new workflow called 'Test Draft' on core.users on_create. Add a delay action first, then a create_alert action.",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

**Turn 2 (same conversation):**
```
Actually, remove the delay action. Just do the alert directly.
```

**Pass criteria:**
- [ ] Turn 1: agent calls `start_draft`, then `add_draft_action` twice (delay + alert)
- [ ] Turn 2: agent calls `remove_draft_action` with the delay action's name
- [ ] Turn 2: agent calls `preview_draft` with just 1 action remaining
- [ ] Preview shows only the alert action (no delay)
- [ ] The start edge correctly points to the alert action now

---

## Phase 11: Final Validation

**Goal**: Verify all previous phases succeeded and nothing was accidentally persisted.

**Request:**
```json
{
  "message": "List all workflow rules again. How many exist now?",
  "context_type": "workflow"
}
```

**Pass criteria:**
- [ ] Agent calls `list_workflow_rules`
- [ ] Count includes original 3 seeded rules + the previewed ones (if accepted by user)
- [ ] No rules accidentally persisted from preview-only operations
- [ ] Agent gives an accurate count

---

## Final Checklist

After completing all phases, verify these overall behaviors:

### Tool Usage
- [ ] `discover` called before any draft creation
- [ ] `list_workflow_rules` used for listing (not `get_workflow_rule`)
- [ ] `get_workflow_rule` used for summaries (not `explain_workflow_node`)
- [ ] `explain_workflow_node` used for action-level details and recipient resolution
- [ ] `list_my_alerts` for personal inbox only
- [ ] `list_alerts_for_rule` for checking if a rule has fired
- [ ] `preview_draft` for new workflows (not `preview_workflow`)
- [ ] `preview_workflow` for existing workflow updates (not `preview_draft`)

### Response Quality
- [ ] Agent never dumps raw JSON as the final response
- [ ] Agent uses plain language to explain tool results
- [ ] Agent announces "Preview is ready" after emitting workflow_preview
- [ ] Agent does NOT try to persist changes itself

### Context Handling
- [ ] When `workflow_id` is in context, agent doesn't re-fetch the workflow
- [ ] `workflow_id` is correctly auto-injected into tools that need it

### Error Handling
- [ ] If `discover` returns unexpected output, agent asks for clarification
- [ ] If a draft expires, agent explains and suggests starting over

---

## Quick Prompt Cheat Sheet

| What you want | Prompt pattern |
|--------------|----------------|
| Discover action types | `"What action types are available?"` |
| Discover triggers | `"What trigger types can I use?"` |
| Discover entities | `"What entities can workflows trigger on?"` |
| List all rules | `"Show me all workflow rules"` |
| Get one rule summary | `"Explain the [Rule Name] workflow"` |
| Explain one action | `"What does the [Action Name] action do?"` |
| Who receives alerts | `"Who gets alerts from [Action Name]?"` |
| My alert inbox | `"Do I have any alerts?"` |
| Check if rule fired | `"Has [Rule Name] triggered any alerts?"` |
| Create new workflow | `"Create a new workflow called... on [entity] triggered on_create..."` |
| Add a branching step | `"If [condition], do X. If not, do Y."` |
| Modify existing | `"Change [field] in the [Action Name] action to [value]"` |
| Deactivate rule | `"Pause the [Rule Name] workflow"` |

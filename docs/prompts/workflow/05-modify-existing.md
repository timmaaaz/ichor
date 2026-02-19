# Step 05: Modify an Existing Workflow

**Goal**: Update an existing workflow by reading its current state, making targeted changes, and submitting a `preview_workflow` for user approval. This exercises the full-payload update path.

---

## Context Setup

For modifying a workflow, pass the workflow ID in context. The agent reads the current state from context and uses `preview_workflow` with the full updated payload:

```json
{
  "message": "<prompt>",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<uuid-of-workflow-to-modify>",
    "rule_name": "Branching Test Workflow"
  }
}
```

**Note:** The `workflow_id` is auto-injected into `preview_workflow` and `explain_workflow_node` tool calls — the agent doesn't need to look it up.

---

## Prompt 5A — Change Alert Severity

Use this to test a simple config change on an existing action:

```
Change the alert severity in the "High Value Alert" action from "high" to "critical".
```

**Expected agent behavior:**
1. Reads the workflow from context (does NOT re-call `get_workflow_rule`)
2. Optionally calls `explain_workflow_node` to confirm current config of "High Value Alert"
3. Constructs the updated workflow payload with `severity: "critical"` on that action
4. Calls `preview_workflow` with the full updated workflow + `workflow_id` + description
5. Server emits `workflow_preview` SSE event

**What to verify:**
- Agent does NOT call `get_workflow_rule` if workflow is already in context
- The updated `action_config.severity` is `"critical"` in the preview
- All other actions and edges remain unchanged
- `description` field in `preview_workflow` describes what changed

---

## Prompt 5B — Add a New Action to an Existing Workflow

Use this to test appending an action to a linear workflow:

```
Add a log_audit_entry action at the end of the "Sequence Test Workflow" after the last alert.
The audit log message should be "Sequence workflow completed successfully."
```

**Expected agent behavior:**
1. Reads workflow from context (3 existing actions + 3 edges)
2. Builds the updated payload with a 4th action (`log_audit_entry`)
3. Adds a sequence edge from `Sequence Action 3` to the new log action
4. Calls `preview_workflow` with the full 4-action workflow
5. Emits `workflow_preview`

**What to verify in the preview:**
- 4 actions (3 original + 1 new log_audit_entry)
- 4 edges (start + 2 sequence + 1 new sequence to log)
- Existing action IDs are preserved (use the UUID, not temp:N for existing actions)
- New action uses no ID (will be assigned on save)

---

## Prompt 5C — Remove an Action from a Workflow

Use this to test removing an action while preserving the chain:

```
Remove the "Normal Value Alert" branch from the "Branching Test Workflow".
The workflow should only alert on high-value orders now.
```

**Expected agent behavior:**
1. Reads workflow from context
2. Constructs updated payload WITHOUT the "Normal Value Alert" action
3. Removes the `output-false` edge that connected to it
4. Keeps the `output-true` branch intact
5. Calls `preview_workflow`

**What to verify in the preview:**
- Only 2 actions remain: condition + high value alert
- Only 2 edges: start + the true-branch sequence edge
- The false-branch edge and action are gone

---

## Prompt 5D — Deactivate a Workflow

Use this to test toggling active status:

```
Deactivate the "Simple Test Workflow" — I want to pause it without deleting it.
```

**Expected agent behavior:**
1. Reads workflow from context
2. Constructs updated payload with `is_active: false`
3. Calls `preview_workflow` with `description: "Deactivating workflow"`
4. Emits preview

**What to verify:**
- `is_active: false` in the preview payload
- All actions and edges are unchanged
- Agent explains what "deactivate" means (won't trigger anymore)

---

## Prompt 5E — Rename a Workflow

Use this to test name changes:

```
Rename the "Sequence Test Workflow" to "Order Processing Pipeline".
```

**Expected agent behavior:**
1. Reads workflow from context
2. Constructs updated payload with `name: "Order Processing Pipeline"`
3. Calls `preview_workflow` with the new name
4. Emits preview

**What to verify:**
- `name` is "Order Processing Pipeline" in the preview
- Everything else (actions, edges, trigger, entity) is unchanged

---

## Prompt 5F — Update Alert Recipients

Use this to test recipient changes, which require the full config:

```
In the "Branching Test Workflow", add a role to the "High Value Alert" recipients.
I want the "managers" role to also receive this alert.
```

**Expected agent behavior:**
1. Calls `explain_workflow_node` with `node_name: "High Value Alert"` to see current recipients
2. Constructs updated payload with the managers role UUID added to `recipients.roles`
3. Calls `preview_workflow` with the full updated payload

**What to verify:**
- `explain_workflow_node` is called first (to get current config)
- The updated action_config.recipients.roles includes the new role
- Agent explains it resolved the role name to a UUID

---

## Understanding the Full-Payload Update Pattern

When modifying an existing workflow, `preview_workflow` requires the complete workflow object:

```json
{
  "workflow_id": "<uuid>",
  "workflow": {
    "name": "Workflow Name",
    "is_active": true,
    "entity_id": "<entity-uuid>",
    "trigger_type_id": "<trigger-type-uuid>",
    "actions": [
      {
        "id": "<existing-action-uuid>",
        "name": "Action Name",
        "action_type": "create_alert",
        "action_config": { ... },
        "is_active": true
      }
    ],
    "edges": [
      {
        "target_action_id": "<action-uuid>",
        "edge_type": "start"
      }
    ]
  },
  "description": "Changed alert severity from high to critical"
}
```

Key rules:
- **Existing actions MUST include their `id` UUID** — this preserves the action on update
- **New actions omit `id`** — they get a new UUID on save
- **Edges reference action UUIDs directly** (not `temp:N` when updating)
- The `entity` and `trigger_type` shorthand names also work (resolved to UUIDs automatically)

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Agent re-fetches workflow already in context | Not reading context | Context already has full workflow state — no need to call `get_workflow_rule` |
| Updated payload drops existing action IDs | Agent forgot to include `id` fields | All existing actions must have their UUID `id` in the update payload |
| Agent uses `temp:N` IDs in update | Using draft builder syntax for updates | `temp:N` only works for NEW actions in new workflows — updates must use real UUIDs |
| `preview_workflow` fails with "entity not found" | Entity/trigger names not resolving | Use `entity_id` UUID directly for updates, or verify entity name format |
| Agent calls `add_draft_action` on existing workflow | Using draft builder for updates | Draft builder is for NEW workflows only; use `preview_workflow` for updates |

---

## Notes

- `preview_workflow` is the only tool for updating existing workflows — do NOT use `add_draft_action` on a saved workflow.
- The `workflow_id` in context is auto-injected into `preview_workflow` — agent doesn't need to explicitly pass it.
- For complex updates, the agent may call `explain_workflow_node` first to confirm current config before making changes.
- The `description` field in `preview_workflow` is shown to the user in the preview card — make it meaningful (e.g., "Added audit log at end of sequence").

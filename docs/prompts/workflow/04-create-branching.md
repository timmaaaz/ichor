# Step 04: Create a Branching Workflow (Conditional Logic)

**Goal**: Create a workflow with `evaluate_condition` branching. This exercises multi-action draft builds with `after: "Action:port"` syntax to express conditional flow.

---

## Context Setup

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

## Prompt 4A — Simple True/False Branch

Use this to test the canonical branching pattern:

```
Create a workflow called "Order Value Check" on sales.orders triggered on_create.
Check if the order amount is greater than 1000.
If true (high value), create an alert with severity "high" titled "High Value Order".
If false (normal value), create an alert with severity "low" titled "Standard Order".
```

**Expected agent behavior:**
1. `discover` (action_types) — learns evaluate_condition has `output-true` and `output-false` ports
2. `start_draft` with `name: "Order Value Check"`, entity `sales.orders`, trigger `on_create`
3. `add_draft_action` — condition action (no `after`):
   - `name: "Check Order Value"`
   - `action_type: "evaluate_condition"`
   - `action_config: { "conditions": [{ "field": "amount", "operator": "greater_than", "value": 1000 }] }`
4. `add_draft_action` — true branch:
   - `name: "High Value Alert"`
   - `action_type: "create_alert"`
   - `after: "Check Order Value:output-true"`
5. `add_draft_action` — false branch:
   - `name: "Standard Alert"`
   - `action_type: "create_alert"`
   - `after: "Check Order Value:output-false"`
6. `preview_draft`

**What to verify in the preview:**
- 3 actions: evaluate_condition, and two create_alert actions
- 3 edges: start → condition, condition:output-true → high alert, condition:output-false → standard alert
- `source_output` on the branch edges matches `"output-true"` and `"output-false"`
- Both branch actions are valid `create_alert` configs

---

## Prompt 4B — Inventory Reorder Check

Use this to test a real inventory automation scenario:

```
Build a workflow "Reorder Check" on inventory.inventory_items triggered on_update.
Check if the item is below the reorder point. Use check_reorder_point action type.
If the check succeeds (below threshold), send an alert with title "Reorder Required"
with severity "high". If it fails (above threshold), log an audit entry.
```

**Expected agent behavior:**
1. `discover` (action_types) — learns `check_reorder_point` output ports (`success`/`failure`)
2. `start_draft` with `inventory.inventory_items` / `on_update`
3. `add_draft_action` — check_reorder_point (first action, no `after`)
4. `add_draft_action` — create_alert with `after: "Check Reorder:success"`
5. `add_draft_action` — log_audit_entry with `after: "Check Reorder:failure"`
6. `preview_draft`

**What to verify:**
- `check_reorder_point` is the first/start action
- Both branches have the correct `source_output` values (`success` and `failure`)
- Agent discovers the correct output ports rather than guessing

---

## Prompt 4C — Three-Step Branch with Convergence

Use this to test a longer branching chain:

```
Create "Procurement Approval" on procurement.purchase_orders on_create.
First, evaluate if the order total is above 5000.
If yes: seek approval from the procurement team, then send an email when approved.
If no: immediately send a confirmation email.
```

**Expected agent behavior:**
1. `discover`
2. `start_draft`
3. Add `evaluate_condition` (start node)
4. Add `seek_approval` with `after: "Evaluate:output-true"`
5. Add `send_email` (confirmation after approval) with `after: "Seek Approval:success"` or similar port
6. Add `send_email` (low-value confirmation) with `after: "Evaluate:output-false"`
7. `preview_draft`

**What to verify:**
- 4 actions with correct `after` chaining
- `seek_approval` only on the high-value branch
- Both branches end with a `send_email` action

---

## Understanding Output Ports

Different action types have different output ports:

| Action Type | Output Ports |
|------------|--------------|
| `evaluate_condition` | `output-true`, `output-false` |
| `check_inventory` | `success`, `failure` |
| `check_reorder_point` | `success`, `failure` |
| `allocate_inventory` | `success`, `failure` |
| `reserve_inventory` | `success`, `failure` |
| `seek_approval` | `success`, `failure` (or `approved`, `rejected`) |
| `send_email` | default (single output) |
| `create_alert` | default (single output) |

Use `discover` with `action_types` to get the authoritative list — do not rely on this table as a source of truth.

---

## Validation Checkpoints

For a 3-action branching workflow, the edges should look like:

```json
"edges": [
  {
    "edge_type": "start",
    "source_action_id": null,
    "target_action_id": "<condition-action-id>"
  },
  {
    "edge_type": "sequence",
    "source_action_id": "<condition-action-id>",
    "source_output": "output-true",
    "target_action_id": "<high-value-alert-id>"
  },
  {
    "edge_type": "sequence",
    "source_action_id": "<condition-action-id>",
    "source_output": "output-false",
    "target_action_id": "<standard-alert-id>"
  }
]
```

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Agent uses `"after": "Condition"` without port | Missing port for branching action | For branching actions, `after` MUST include port: `"Condition:output-true"` |
| Agent guesses wrong port names | Skipped `discover` | Always call `discover action_types` to get correct port names |
| Only one branch action created | Agent didn't understand bidirectional branching | Both `output-true` and `output-false` need separate `add_draft_action` calls |
| `source_output` missing from edges | Draft builder failed to parse `after` syntax | Check that `after` value includes the colon-separated port |
| Condition action uses wrong operator | Config schema mismatch | Use `discover` to get the correct operator names for `evaluate_condition` |

---

## Notes

- The `after: "ActionName:port"` shorthand is the primary way to express edges in the draft builder — explicit `edges` array is optional.
- If `after` uses a port that doesn't exist for that action type, the draft validation will fail. Use `discover` to get the correct port names.
- For converging branches (multiple paths leading to a single next step), each branch adds its own `add_draft_action` call with an `after` pointing to the convergence step.

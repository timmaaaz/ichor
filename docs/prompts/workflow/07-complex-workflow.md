# Step 07: Complex Real-World Workflows

**Goal**: Test multi-step, multi-branch workflows that represent realistic business automation scenarios. These exercises combine discovery, draft building, branching, and multi-action chains.

---

## Context Setup

All complex workflows start fresh:

```json
{
  "message": "<prompt>",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

---

## Scenario 7A — Inventory Reorder Pipeline

A complete inventory automation: check stock, branch on threshold, reserve or alert.

```
Build a workflow called "Inventory Reorder Pipeline" on inventory.inventory_items
triggered when items are updated.

The workflow should:
1. Check if the item is below the reorder point
2. If below threshold: reserve inventory, then create an alert "Reorder Triggered" with severity "high"
3. If above threshold: log an audit entry "Stock level acceptable"
```

**Expected tool sequence:**
1. `discover` (action_types) — learns `check_reorder_point`, `reserve_inventory`, `create_alert`, `log_audit_entry`
2. `start_draft`
3. `add_draft_action` — check_reorder_point (start node)
4. `add_draft_action` — reserve_inventory with `after: "Check Reorder:success"`
5. `add_draft_action` — create_alert with `after: "Reserve Inventory:success"` (or appropriate port)
6. `add_draft_action` — log_audit_entry with `after: "Check Reorder:failure"`
7. `preview_draft`

**What to verify:**
- 4 actions, 4 edges (start + 3 sequences)
- Two distinct branches from the check action
- Success branch: check → reserve → alert (3 steps)
- Failure branch: check → log (2 steps)
- Correct `source_output` values on each branch edge

---

## Scenario 7B — Order Approval Workflow

A procurement workflow with human approval gate:

```
Create "Purchase Order Approval" on procurement.purchase_orders triggered on_create.

Steps:
1. Evaluate if total > 10000 (high-value order)
2. If yes: seek approval from the procurement manager, then send email "PO Approved" on success
3. If no: directly send a confirmation email "PO Auto-Approved — under threshold"
```

**Expected tool sequence:**
1. `discover`
2. `start_draft`
3. `add_draft_action` — evaluate_condition (no after)
4. `add_draft_action` — seek_approval with `after: "Evaluate PO:output-true"`
5. `add_draft_action` — send_email (approved) with `after: "Seek Approval:success"`
6. `add_draft_action` — send_email (auto-approved) with `after: "Evaluate PO:output-false"`
7. `preview_draft`

**What to verify:**
- 4 actions
- High-value path: condition → seek_approval → send_email (3 steps)
- Low-value path: condition → send_email (2 steps)
- Both `send_email` actions have different subjects/bodies per the prompt

---

## Scenario 7C — New User Onboarding Sequence

A linear multi-step workflow with delay:

```
Build "New User Onboarding" on core.users triggered on_create.

Sequence:
1. Send a welcome email immediately
2. Wait 24 hours
3. Send a follow-up email with onboarding resources
4. Create an alert for HR: "New user onboarded — verify profile"
```

**Expected tool sequence:**
1. `discover`
2. `start_draft` (entity: `core.users`, trigger: `on_create`)
3. `add_draft_action` — send_email "welcome" (no after)
4. `add_draft_action` — delay with `after: "Welcome Email"`
5. `add_draft_action` — send_email "follow-up" with `after: "Wait 24h"` (or whatever delay was named)
6. `add_draft_action` — create_alert with `after: "Follow-Up Email"`
7. `preview_draft`

**What to verify:**
- 4 actions in a linear chain
- 4 edges: start + 3 sequences
- No branching (all `always` or `sequence` edges, not `output-true/false`)
- Delay action correctly named so the next `after` reference resolves

---

## Scenario 7D — Full Stock Management Workflow

Tests the inventory-specific actions together:

```
Create "Stock Management" on inventory.inventory_items on_update.

If stock is low (check_reorder_point fails):
  1. Allocate existing inventory (allocate_inventory)
  2. If allocation succeeds: commit the allocation, then alert "Stock Allocated"
  3. If allocation fails: alert "Allocation Failed" with severity critical

If stock is fine (check passes): log_audit_entry "Stock OK"
```

**Expected tool sequence:**
1. `discover` (learn allocate_inventory, commit_allocation output ports)
2. `start_draft`
3. Check reorder (start node)
4. Allocate with `after: "Check:success"`  ← Note: "failure" = below threshold = needs reorder
   (Agent may need to interpret "low stock = check failure" semantics)
5. Commit allocation with `after: "Allocate:success"`
6. Alert "Stock Allocated" with `after: "Commit:success"`
7. Alert "Allocation Failed" with `after: "Allocate:failure"`
8. Log audit with `after: "Check:failure"`  ← "success" = above threshold = stock OK
9. `preview_draft`

**What to verify:**
- Complex multi-level branching
- 6 actions total
- Correct port interpretation (agent may need to clarify the "success vs failure" semantics with user)
- If agent is unsure about port semantics, it should ask the user

---

## Scenario 7E — One-Shot Complex Request

Tests whether the agent handles all steps in a single message (no interactive back-and-forth):

```
Build a complete workflow called "Sales Order Fulfillment" on sales.orders triggered on_update.
1. Check if status changed to "fulfillment_ready"
2. If yes: allocate inventory, then send email "Fulfillment Started" to the warehouse team
3. If allocation fails: alert "Inventory Shortage" severity critical to operations
4. If status not ready: log audit "Order status update received"
```

**Expected agent behavior:**
- Does NOT ask clarifying questions (requirements are fully specified)
- Completes the full draft in one conversation turn
- Calls `discover`, `start_draft`, multiple `add_draft_action`, then `preview_draft`
- Presents a summary of the built workflow before the preview

**What to verify:**
- Agent doesn't get stuck asking questions when enough info is provided
- Preview covers all 4+ actions with correct branching
- No missed steps

---

## Complexity Checklist

For a well-formed complex workflow preview, verify:

- [ ] **Actions**: All expected action types present
- [ ] **Edges**: 1 start edge + N-1 sequence edges (at minimum)
- [ ] **Branching**: `source_output` set on all branch edges
- [ ] **Configs**: All action_config fields present (no empty objects)
- [ ] **Active**: `is_active: true` on all actions and the rule
- [ ] **Names**: Action names match references in `after` fields
- [ ] **Ports**: `source_output` values match the actual output ports from `discover`

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Agent adds all actions in one message without confirming | Spec ambiguous | For complex workflows, agent should confirm plan before building |
| Branch from wrong port | Wrong success/failure semantics | Agent should discover ports and confirm with user if unclear |
| Missing `after` references break chain | Typo in action name | Action names in `after` must match exactly what was passed to `add_draft_action` |
| Draft expires mid-conversation | 10-minute TTL | Start a new draft if agent reports "draft not found" |
| Agent uses `preview_workflow` instead of `preview_draft` | Wrong preview for new workflow | New drafts → `preview_draft`; existing workflow updates → `preview_workflow` |

---

## Notes

- Complex workflows benefit from the agent summarizing what it built before the preview: "I've added 5 actions: [list]. Does this match what you wanted before I preview?"
- The agent should call `discover` once at the start and use the results for all action config schemas in that conversation — not once per action.
- For workflows where port semantics are ambiguous (e.g., does `check_reorder_point` "succeed" when stock IS low or is NOT low?), the agent should ask for clarification.

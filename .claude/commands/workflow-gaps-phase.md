# Workflow Action Gap Remediation Phase Jump Command

Jump to and execute a specific phase in the Workflow Action Gap Remediation implementation plan.

## Your Task

### 1. Parse Parameters

This command takes a phase number parameter: `/workflow-gaps-phase N`

If no parameter provided, ask the user which phase they want to jump to.

### 2. Validate Phase Number

1. Read `.claude/plans/WORKFLOW_GAPS_PLAN/PROGRESS.yaml`
2. Verify phase number is valid (between 1 and 9)
3. Check if phase documentation exists (`.claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_N_*.md`)

### 3. Check Prerequisites

Before jumping to the phase:

1. Check if previous phases are completed
   - If jumping ahead (skipping uncompleted phases), warn the user:
     ```
     ⚠️  Warning: You are jumping to Phase N, but these earlier phases are not completed:
     - Phase X: [name] (status: pending/in_progress)
     - Phase Y: [name] (status: pending/in_progress)

     This may cause issues if Phase N depends on earlier work.
     Do you want to proceed anyway? (yes/no)
     ```
   - Wait for user confirmation

2. Check phase dependencies:
   - Read `dependencies.internal` from the phase in PROGRESS.yaml
   - Verify all dependency phases are `completed`
   - If not, warn the user and ask for confirmation

3. Check external dependencies:
   - Read `dependencies.external` from PROGRESS.yaml
   - Verify all external plans are completed
   - If not, warn the user

### 4. Update Current Phase

If user confirms (or no warnings):
- Update `project.current_phase` to the selected phase number
- Update `project.status` to `in_progress`
- Update `context.current_focus` to describe this phase
- Save PROGRESS.yaml

### 5. Execute the Phase

Follow the exact same workflow as the `/workflow-gaps-next` command:

1. Load phase documentation
2. Set phase status to `in_progress`
3. Create TodoWrite list from phase tasks
4. Execute tasks sequentially with incremental PROGRESS.yaml updates
5. Run phase validation
6. Mark phase as completed if validation passes
7. Provide completion summary

Refer to the `workflow-gaps-next` command for detailed execution steps.

## Important Notes

### Jumping Ahead

- Strongly discourage jumping ahead unless user has a specific reason
- Make warnings very visible
- Explain potential consequences
- Always get explicit user confirmation

### Jumping Back

- Jumping back to a completed phase is safer
- Ask user if they want to:
  - Re-execute the phase (reset status to `in_progress`)
  - Review the phase without re-executing
  - Modify the phase (expert users only)

### Error Handling

- If phase documentation doesn't exist, suggest running `/workflow-gaps-build-phase`
- If jumping to a phase that's currently `blocked`, explain the blocker first
- If phase has status `in_progress`, ask if they want to continue or restart

## Phase Reference

| Phase | Name | Category |
|-------|------|----------|
| 1 | Add Missing Tables to Whitelist | backend |
| 2 | Fix FieldChanges Propagation | backend |
| 3 | Implement send_notification | backend |
| 4 | Implement send_email | backend |
| 5 | Implement seek_approval | backend+database |
| 6 | Add create_purchase_order | backend |
| 7 | Add receive_inventory | backend |
| 8 | Add call_webhook | backend |
| 9 | Add Template Arithmetic | backend |

## Example Usage

```
/workflow-gaps-phase 2    # Jump to Phase 2 (FieldChanges fix)
/workflow-gaps-phase 5    # Jump to Phase 5 (seek_approval - complex)
/workflow-gaps-phase 9    # Jump to Phase 9 (template arithmetic)
```

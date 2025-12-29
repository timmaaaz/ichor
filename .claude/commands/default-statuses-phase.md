# Default Status Management Phase Jump Command

Jump to and execute a specific phase in the Default Status Management implementation plan.

## Your Task

### 1. Parse Parameters

This command takes a phase number parameter: `/default-statuses-phase N`

If no parameter provided, ask the user which phase they want to jump to.

### 2. Validate Phase Number

1. Read `.claude/plans/DEFAULT_STATUSES_PLAN/PROGRESS.yaml`
2. Verify phase number is valid (between 1 and 3)
3. Check if phase documentation exists (`.claude/plans/DEFAULT_STATUSES_PLAN/phases/PHASE_N_*.md`)

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
   - Phase 2 depends on Phase 1
   - Phase 3 depends on Phase 2
   - If not, warn the user and ask for confirmation

### 4. Update Current Phase

If user confirms (or no warnings):
- Update `project.current_phase` to the selected phase number
- Update `project.status` to `in_progress`
- Update `context.current_focus` to describe this phase
- Save PROGRESS.yaml

### 5. Execute the Phase

Follow the exact same workflow as the `/default-statuses-next` command:

1. Load phase documentation
2. Set phase status to `in_progress`
3. Create TodoWrite list from phase tasks
4. Execute tasks sequentially with incremental PROGRESS.yaml updates
5. Run phase validation
6. Mark phase as completed if validation passes
7. Provide completion summary

Refer to the `default-statuses-next.md` for detailed execution steps.

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

- If phase documentation doesn't exist, suggest running `/default-statuses-build-phase`
- If jumping to a phase that's currently `blocked`, explain the blocker first
- If phase has status `in_progress`, ask if they want to continue or restart

## Phase Reference

| Phase | Name | Dependencies |
|-------|------|--------------|
| 1 | Form Configuration FK Default Resolution | None |
| 2 | Workflow Integration for Status Transitions | Phase 1 |
| 3 | Alert System Enhancement | Phase 2 |

## Example Usage

```bash
# Jump to Phase 2
/default-statuses-phase 2

# If Phase 1 not complete, you'll see:
# ⚠️  Warning: Phase 1 (Form Configuration FK Default Resolution) is not completed.
# Phase 2 depends on FK default resolution being implemented.
# Do you want to proceed anyway? (yes/no)
```

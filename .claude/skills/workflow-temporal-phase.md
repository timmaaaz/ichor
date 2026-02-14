# Workflow Temporal Implementation Phase Jump Command

Jump to and execute a specific phase in the Workflow Temporal Implementation plan.

## Your Task

### 1. Parse Parameters

This command takes a phase number parameter: `/workflow-temporal-phase N`

If no parameter provided, ask the user which phase they want to jump to.

### 2. Validate Phase Number

1. Read `.claude/plans/WORKFLOW_TEMPORAL_PLAN/PROGRESS.yaml`
2. Verify phase number is valid (between 1 and 13)
3. Check if phase documentation exists (`.claude/plans/WORKFLOW_TEMPORAL_PLAN/phases/PHASE_N_*.md`)

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

Follow the exact same workflow as the `/workflow-temporal-next` command:

1. Load phase documentation
2. Set phase status to `in_progress`
3. Create TodoWrite list from phase tasks
4. Execute tasks sequentially with incremental PROGRESS.yaml updates
5. Run phase validation
6. Mark phase as completed if validation passes
7. Provide completion summary

Refer to the `next.md.template` for detailed execution steps.

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

- If phase documentation doesn't exist, suggest running `/workflow-temporal-build-phase`
- If jumping to a phase that's currently `blocked`, explain the blocker first
- If phase has status `in_progress`, ask if they want to continue or restart

## Example Usage

### Scenario 1: Jumping Ahead

```
User: /workflow-temporal-phase 5
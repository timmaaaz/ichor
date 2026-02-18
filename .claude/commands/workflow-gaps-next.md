# Workflow Action Gap Remediation Next Phase Command

Execute the next pending phase in the Workflow Action Gap Remediation implementation plan, or continue the current in-progress phase.

## Your Task

### 1. Determine Which Phase to Execute

1. Read `.claude/plans/WORKFLOW_GAPS_PLAN/PROGRESS.yaml`
2. Check the `current_phase` field and phase statuses:
   - If a phase has status `in_progress`, continue that phase
   - Otherwise, find the first phase with status `pending`
   - If all phases are `completed`, congratulate the user and exit
3. Verify phase documentation exists (e.g., `.claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_1_*.md`)

### 2. Check Dependencies

For the selected phase:
- Check `dependencies.internal` - verify all prerequisite phases are `completed`
- Check `dependencies.external` - verify all external plans are completed
- If dependencies not met, warn user and ask if they want to proceed anyway

### 3. Load Phase Instructions

Load the corresponding phase file from `.claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_N_*.md`

If phase documentation doesn't exist, suggest running `/workflow-gaps-build-phase` first.

### 4. Check Prerequisites

Before starting, verify:
- All previous phases are marked `completed` in PROGRESS.yaml
- If prerequisites are not met, inform the user and ask if they want to continue anyway
- For backend phases, ensure Go backend access is available
- For database phases, ensure database access is available

### 5. Update PROGRESS.yaml - Phase Start

Update the phase entry in PROGRESS.yaml:
- Set `status: in_progress`
- Update `project.current_phase` to this phase number
- Update `project.status` to `in_progress`
- Update `project.summary.phases_in_progress` count
- Update `project.summary.phases_pending` count
- Save the file

### 6. Create Todo List

Based on the phase tasks in PROGRESS.yaml, create a todo list using the TodoWrite tool:
- One todo item for each task in the phase
- Use descriptive names matching the task names from PROGRESS.yaml
- All start as `pending` status
- Use proper activeForm (present continuous, e.g., "Creating component" not "Create component")

### 7. Execute Tasks Systematically

For each task in order:

1. **Mark task as in_progress** in your todo list
2. **Update PROGRESS.yaml**: Set the task `status: in_progress`
3. **Execute the task** by following the phase instructions:
   - Read the detailed implementation steps from the phase markdown
   - Create or modify files as specified
   - Follow the exact specifications and code examples provided
   - For each file affected:
     - Update the file's `status: in_progress` in PROGRESS.yaml
     - After completing the file, set `status: completed`
4. **Validate the task**: Run any validation checks specified in the phase file
5. **Mark task as completed** in your todo list
6. **Update PROGRESS.yaml**: Set the task `status: completed`
7. **Move to next task**

### 8. Progress Updates - Task Level

After each task completion, update PROGRESS.yaml:
- Set task `status: completed`
- Update any file statuses within that task
- Update `context.next_task` to the next pending task
- Add accomplishment to `context.recent_changes`
- Increment `project.summary.files_created` or `files_modified` as appropriate
- Save the file

**Important**: Update PROGRESS.yaml incrementally after each task, not all at once at the end.

### 9. Phase Validation

After all tasks are completed:

1. Run validation checks specified in the phase's `validation` section
2. Update each validation check's `status: completed` in PROGRESS.yaml
3. If any validation fails:
   - Mark the phase as `blocked`
   - Add the issue to `blockers`
   - Inform the user
   - Do NOT mark phase as completed

### 10. Update PROGRESS.yaml - Phase Complete

If all validations pass:
- Set phase `status: completed`
- Update `project.summary.phases_completed` count
- Update `project.summary.phases_in_progress` count
- Update `context.current_focus` to next phase
- Update `context.next_task` to describe next phase
- Add key decisions to `context.decisions` if any were made
- Save the file

### 11. Summary Report

Provide a comprehensive summary:
- What was accomplished
- Files created/modified
- Validation results
- Next phase preview
- Any notes or recommendations
- Suggested next commands

## Important Guidelines

### Incremental Progress Updates

- Update PROGRESS.yaml after EACH task, not just at the end
- This allows resuming if interrupted
- Use Edit tool to update specific sections

### Error Handling

- If a task fails, mark it as `blocked` and document the blocker
- Ask the user how they want to proceed
- Don't continue to next phase if current phase is incomplete

### Code Quality

- Follow all patterns and conventions in CLAUDE.md
- Run `go build ./...` and `go test ./...` for Go changes
- Ensure all code compiles before marking tasks complete

### Testing as You Go

- Test each file after creating it
- Don't wait until the end to discover issues
- If something doesn't work, fix it before proceeding

### Communication

- Be verbose about what you're doing
- Explain each step clearly
- Show progress frequently
- Ask for confirmation before major changes

## Category-Specific Notes

### Backend Phases (Go)

- All backend work in Go codebase
- Run `go build ./...` after each significant change
- Run `go test ./...` after completing a phase
- Follow Ardan Labs patterns strictly
- Keep layers pure: no business logic in API, no HTTP in business

### Database Phases

- Use migrations for schema changes (`business/sdk/migrate/sql/migrate.sql`)
- Always add new version, never edit existing
- Test with `make migrate`

## Example Usage

User runs: `/workflow-gaps-next`

You respond:

1. "Reading PROGRESS.yaml to determine next phase..."
2. "Phase 2 ([Phase Name]) is next. Prerequisites: Phase 1 ✅"
3. "Loading phase documentation..."
4. "Starting Phase 2..."
5. [Update PROGRESS.yaml]
6. [Create todo list with TodoWrite]
7. "Working on Task 1: [task name]..."
8. [Execute task, update PROGRESS.yaml]
9. "Task 1 complete! Moving to Task 2..."
10. [Continue through all tasks]
11. "Running phase validation checks..."
12. [Validate phase completion]
13. "Phase 2 complete! Summary: [details]"
14. "Ready for Phase 3. Run `/workflow-gaps-next` again to continue."

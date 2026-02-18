# Workflow Action Gap Remediation Build Phase Command

Generate documentation for the next phase using the phase template.

Note: All 9 phase documents for this plan have already been created in `.claude/plans/WORKFLOW_GAPS_PLAN/phases/`. This command is provided for completeness and in case additional phases need to be documented in the future.

## Your Task

### 1. Determine Next Phase to Document

1. Read `.claude/plans/WORKFLOW_GAPS_PLAN/PROGRESS.yaml`
2. Check `planning_status.next_phase_to_document`
3. If `null`, all phases are documented - inform user and exit
4. Verify the phase exists in the `phases` array

### 2. Gather Phase Information from User

Ask the user the following questions about the phase:

#### Question 1: Phase Name
```
What is the name of Phase N?
(Short, descriptive name like "Backend API Development" or "User Authentication UI")
```

#### Question 2: Phase Category
```
What category is this phase?
Options:
- backend: Go backend work (API, business logic, database access)
- frontend: Vue 3 frontend work (components, composables, views)
- database: Database schema, migrations, type generation
- fullstack: Both backend and frontend work
- testing: Test creation and validation
- documentation: Documentation updates and guides

Category:
```

#### Question 3: Phase Description
```
Provide a 1-2 sentence description of what this phase accomplishes:
```

#### Question 4: Primary Goals
```
What are the 2-3 primary goals of this phase?
(One goal per line)
```

#### Question 5: Key Deliverables
```
What are the main deliverables? (files/components to create)
(One per line, with relative paths if possible)
```

#### Question 6: Dependencies
```
Does this phase depend on any previous phases?
(Enter phase numbers separated by commas, or 'none')
```

### 3. Generate Phase Slug

Create a URL-friendly slug from the phase name:
- Convert to uppercase
- Replace spaces with underscores
- Remove special characters
- Example: "Call Webhook" → "CALL_WEBHOOK"

### 4. Write Phase Documentation

Write the generated phase document to:
`.claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_N_[PHASE_SLUG].md`

### 5. Update PROGRESS.yaml

Update the planning status in PROGRESS.yaml:

```yaml
planning_status:
  phase_N_doc_created: true
  next_phase_to_document: N+1 or null if all phases documented
```

Also update the phase entry in the `phases` array:
```yaml
- phase: N
  name: "[PHASE_NAME]"
  description: "[PHASE_DESCRIPTION]"
  status: "pending"
  category: "[PHASE_CATEGORY]"

  tasks:
    - task: "TODO: Define task 1"
      status: "pending"
      notes: []
      files: []

  validation:
    - check: "go build ./... passes"
      status: "pending"
    - check: "go test ./... passes"
      status: "pending"

  deliverables: []
  blockers: []
```

### 6. Provide Summary

Show the user what was created:

```
✅ Phase N Documentation Created

════════════════════════════════════════════════════════

PHASE DETAILS
Name: [PHASE_NAME]
Category: [PHASE_CATEGORY]
File: .claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_N_[PHASE_SLUG].md
Status: Pending (ready for execution)

NEXT STEPS

1. Review the generated phase doc
2. When ready to execute, run: /workflow-gaps-next

════════════════════════════════════════════════════════

PLANNING STATUS
Phases Documented: [DOCUMENTED_COUNT] / 9
Next Phase to Document: Phase [NEXT_PHASE] or "All phases documented!"
```

## Current Status

All 9 phases are already documented:
- Phase 1: PHASE_1_MISSING_TABLES.md ✅
- Phase 2: PHASE_2_FIELD_CHANGES.md ✅
- Phase 3: PHASE_3_SEND_NOTIFICATION.md ✅
- Phase 4: PHASE_4_SEND_EMAIL.md ✅
- Phase 5: PHASE_5_SEEK_APPROVAL.md ✅
- Phase 6: PHASE_6_CREATE_PO.md ✅
- Phase 7: PHASE_7_RECEIVE_INVENTORY.md ✅
- Phase 8: PHASE_8_CALL_WEBHOOK.md ✅
- Phase 9: PHASE_9_TEMPLATE_ARITHMETIC.md ✅

If the user runs this command, inform them all phases are already documented and suggest starting with `/workflow-gaps-plan-review 1` to review the first plan before implementing.

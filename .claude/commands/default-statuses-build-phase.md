# Default Status Management Build Phase Command

Generate documentation for the next phase using the phase template.

## Your Task

### 1. Determine Next Phase to Document

1. Read `.claude/plans/DEFAULT_STATUSES_PLAN/PROGRESS.yaml`
2. Check `planning_status.next_phase_to_document`
3. If `null`, all phases are documented - inform user and exit
4. Verify the phase exists in the `phases` array

### 2. Load Existing Phase Information

For this plan, the phases are already well-defined in the README.md. Read:
- `.claude/plans/DEFAULT_STATUSES_PLAN/README.md`
- `.claude/plans/default-statuses/plan.md` (original detailed plan)

Extract the detailed information for the phase being documented.

### 3. Generate Phase Documentation

Create a detailed phase document at:
`.claude/plans/DEFAULT_STATUSES_PLAN/phases/PHASE_N_NAME.md`

**Phase 1 File**: `PHASE_1_FORM_FK_DEFAULTS.md`
**Phase 2 File**: `PHASE_2_WORKFLOW_INTEGRATION.md`
**Phase 3 File**: `PHASE_3_ALERT_SYSTEM.md`

### Phase Document Structure

```markdown
# Phase N: [Phase Name]

**Status**: Pending
**Category**: [backend/frontend/fullstack]
**Dependencies**: [Previous phases]

---

## Overview

[Description of what this phase accomplishes]

## Goals

1. [Primary goal 1]
2. [Primary goal 2]
3. [Primary goal 3]

---

## Tasks

### Task 1: [Task Name]

**Files to Modify:**
- `path/to/file.go`

**Implementation Steps:**
1. Step 1
2. Step 2
3. Step 3

**Code Examples:**
```go
// Example code
```

### Task 2: [Task Name]

[Continue for all tasks...]

---

## Validation Criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

---

## Testing Strategy

[How to test this phase]

---

## Deliverables

- Deliverable 1
- Deliverable 2

---

## Notes & Gotchas

- Important consideration 1
- Important consideration 2
```

### 4. Update PROGRESS.yaml

Update the planning status:

```yaml
planning_status:
  phase_N_doc_created: true
  next_phase_to_document: N+1 or null
```

### 5. Provide Summary

```
✅ Phase N Documentation Created

════════════════════════════════════════════════════════════

PHASE DETAILS
Name: [Phase Name]
Category: [Category]
File: .claude/plans/DEFAULT_STATUSES_PLAN/phases/PHASE_N_NAME.md
Status: Pending (ready for execution)

PRIMARY GOALS
1. [Goal 1]
2. [Goal 2]
3. [Goal 3]

KEY DELIVERABLES
- [Deliverable 1]
- [Deliverable 2]

════════════════════════════════════════════════════════════

NEXT STEPS

1. Review the generated phase doc and refine if needed
2. Run /default-statuses-next to begin execution
3. Run /default-statuses-build-phase to document next phase

════════════════════════════════════════════════════════════

PLANNING STATUS
Phases Documented: N / 3
Next Phase to Document: Phase N+1 or "All phases documented!"
```

## Phase-Specific Content

### Phase 1: Form Configuration FK Default Resolution

- Category: backend
- Key files: formdataapp.go, forms.go, tableforms.go
- Focus: FK name-to-UUID resolution in formdata package

### Phase 2: Workflow Integration for Status Transitions

- Category: backend
- Key files: allocate.go, seedFrontend.go, updatefield.go
- Focus: Automation rules, workflow event firing

### Phase 3: Alert System Enhancement

- Category: fullstack
- Key files: migrate.sql, alert.go, frontend components
- Focus: Alert tables, routing, UI

## Tips

- Use the detailed content from the original plan.md
- Include specific file paths and line numbers where helpful
- Add code examples where the original plan provides them
- Make the phase documentation actionable and executable

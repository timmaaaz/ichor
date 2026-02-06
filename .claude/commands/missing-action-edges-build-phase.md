# Universal Action Edge Enforcement Build Phase Command

Generate documentation for the next phase using the phase template.

## Your Task

### 1. Determine Next Phase to Document

1. Read `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
2. Check `planning_status.next_phase_to_document`
3. If `null`, all phases are documented - inform user and exit
4. Verify the phase exists in the `phases` array

### 2. Load Phase Information

For the Universal Action Edge Enforcement plan, phase information is already defined in PROGRESS.yaml. Extract:

- Phase number
- Phase name
- Phase description
- Phase category
- Tasks already defined
- Deliverables
- Validation checks

### 3. Read the Original Plan Document

Read `.claude/plans/missing-action-edges.md` for detailed context about what each phase should accomplish. This contains:

- Specific files to modify with line numbers
- Code examples and patterns to follow
- Detailed task breakdowns

### 4. Generate Phase Slug

Create a URL-friendly slug from the phase name:
- Convert to uppercase
- Replace spaces with underscores
- Remove special characters
- Example: "Validation Layer Changes" → "VALIDATION_LAYER"

### 5. Generate Phase Documentation

Create a comprehensive phase document with:

```markdown
# Phase {{N}}: {{PHASE_NAME}}

**Category**: {{CATEGORY}}
**Status**: Pending
**Dependencies**: {{DEPENDENCIES or "None"}}

---

## Overview

{{PHASE_DESCRIPTION}}

### Goals

1. {{PRIMARY_GOAL_1}}
2. {{PRIMARY_GOAL_2}}
3. {{PRIMARY_GOAL_3}}

### Why This Phase Matters

{{EXPLAIN IMPORTANCE}}

---

## Prerequisites

Before starting this phase, ensure:

- [ ] Previous phases are completed
- [ ] Go development environment is ready
- [ ] Database access is available (if needed)

---

## Task Breakdown

### Task 1: {{TASK_NAME}}

**Files**:
- `{{FILE_PATH}}` - {{WHAT TO CHANGE}}

**Implementation**:

```go
// Code example from original plan
```

**Validation**:
- [ ] Code compiles
- [ ] Tests pass

### Task 2: {{TASK_NAME}}

...

---

## Deliverables

- [ ] {{DELIVERABLE_1}}
- [ ] {{DELIVERABLE_2}}
- [ ] {{DELIVERABLE_3}}

---

## Validation Criteria

- [ ] Go compilation passes
- [ ] All tests pass
- [ ] {{PHASE_SPECIFIC_VALIDATION}}

---

## Testing Strategy

### What to Test

- {{TEST_SCENARIO_1}}
- {{TEST_SCENARIO_2}}

### How to Test

```bash
# Commands to run
make test
```

---

## Gotchas and Tips

- {{TIP_1}}
- {{TIP_2}}

---

## Reference

- Original plan: `.claude/plans/missing-action-edges.md`
- Progress tracking: `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
```

### 6. Write Phase Documentation

Write the generated phase document to:
`.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_{{N}}_{{PHASE_SLUG}}.md`

### 7. Update PROGRESS.yaml

Update the planning status in PROGRESS.yaml:

```yaml
planning_status:
  phase_{{N}}_doc_created: true
  next_phase_to_document: {{N+1}} or null if all phases documented
```

### 8. Provide Summary

Show the user what was created:

```
✅ Phase {{N}} Documentation Created

════════════════════════════════════════════════════════

PHASE DETAILS
Name: {{PHASE_NAME}}
Category: {{PHASE_CATEGORY}}
File: .claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_{{N}}_{{PHASE_SLUG}}.md
Status: Pending (ready for execution)

PRIMARY GOALS
1. {{GOAL_1}}
2. {{GOAL_2}}
3. {{GOAL_3}}

KEY DELIVERABLES
- {{DELIVERABLE_1}}
- {{DELIVERABLE_2}}
- {{DELIVERABLE_3}}

DEPENDENCIES
{{DEPENDENCIES or "None"}}

════════════════════════════════════════════════════════

NEXT STEPS

The phase documentation has been created with detailed instructions.

You should now:

1. Review the generated phase doc:
   .claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_{{N}}_{{PHASE_SLUG}}.md

2. Review the plan before implementing:
   /missing-action-edges-plan-review {{N}}

3. When ready to execute, run:
   /missing-action-edges-next

4. To generate next phase doc, run:
   /missing-action-edges-build-phase

════════════════════════════════════════════════════════

PLANNING STATUS
Phases Documented: {{DOCUMENTED_COUNT}} / 6
Next Phase to Document: Phase {{NEXT_PHASE}} or "All phases documented!"
```

## Category-Specific Template Adjustments

### Backend Phases (1, 2, 3, 5)

Add these specific sections:
- Go package structure
- Database changes (if any)
- Business logic organization
- Error handling patterns
- Testing with Go tests

### Testing Phases (4)

Add these specific sections:
- Test file structure
- Test scenarios
- Mock data and fixtures
- Expected vs actual behavior
- How to run tests

### Documentation Phases (6)

Add these specific sections:
- Documentation structure
- Files to update
- Sections to remove/add
- Verification checklist

## Tips

- Pull detailed information from the original `.claude/plans/missing-action-edges.md`
- Include specific file paths and line numbers where available
- Generate useful code examples based on existing patterns
- Make it easy for the executor to follow step-by-step
- Update PROGRESS.yaml immediately

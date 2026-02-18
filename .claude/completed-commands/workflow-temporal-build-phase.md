# Workflow Temporal Implementation Build Phase Command

Generate documentation for the next phase using the phase template.

## Your Task

### 1. Determine Next Phase to Document

1. Read `.claude/plans/WORKFLOW_TEMPORAL_PLAN/PROGRESS.yaml`
2. Check `planning_status.next_phase_to_document`
3. If `null`, all phases are documented - inform user and exit
4. Verify the phase exists in the `phases` array

### 2. Gather Phase Information from User

Ask the user the following questions about the phase:

#### Question 1: Phase Name Confirmation
```
The next phase to document is Phase {{N}}.
Current name from PROGRESS.yaml: "{{PHASE_NAME}}"

Is this name correct, or would you like to change it?
(Press enter to keep, or type a new name)
```

#### Question 2: Category Confirmation
```
Current category from PROGRESS.yaml: "{{CATEGORY}}"

Is this category correct?
Options:
- backend: Go backend work (API, business logic, database access)
- infrastructure: K8s, Docker, Makefile changes
- research: Evaluation and decision-making
- testing: Test creation and validation
- deployment: K8s deployment and production config

Category (press enter to keep current):
```

#### Question 3: Primary Goals
```
What are the 2-3 primary goals of this phase?
(One goal per line, based on tasks in PROGRESS.yaml)

Current tasks listed:
{{LIST TASKS FROM PROGRESS.yaml}}

Primary goals:
```

#### Question 4: Key Deliverables
```
What are the main deliverables? (files/components to create)
(One per line, with relative paths)

Based on PROGRESS.yaml files:
{{LIST FILES FROM PROGRESS.yaml TASKS}}

Deliverables (press enter to use above):
```

#### Question 5: Dependencies Confirmation
```
Does this phase depend on any previous phases?
Current dependencies from PROGRESS.yaml: {{DEPENDENCIES or "None"}}

Dependencies (press enter to keep, or specify phase numbers):
```

### 3. Generate Phase Slug

Create a URL-friendly slug from the phase name:
- Convert to uppercase
- Replace spaces with underscores
- Remove special characters
- Example: "Infrastructure Setup" → "INFRASTRUCTURE_SETUP"

### 4. Generate Phase Documentation

Create comprehensive phase documentation including:

```markdown
# Phase {{N}}: {{PHASE_NAME}}

**Category**: {{CATEGORY}}
**Status**: Pending
**Dependencies**: {{DEPENDENCIES}}

---

## Overview

{{PHASE_DESCRIPTION from PROGRESS.yaml}}

## Goals

1. {{GOAL_1}}
2. {{GOAL_2}}
3. {{GOAL_3}}

## Prerequisites

- {{LIST PREREQUISITES based on dependencies}}

---

## Task Breakdown

### Task 1: {{TASK_NAME}}

**Status**: Pending

**Description**: {{TASK_DESCRIPTION from PROGRESS.yaml}}

**Notes**:
{{LIST NOTES from PROGRESS.yaml}}

**Files**:
{{LIST FILES from PROGRESS.yaml with paths}}

**Implementation Guide**:

```go
// TODO: Add implementation example
```

---

### Task 2: {{TASK_NAME}}

... (repeat for all tasks)

---

## Validation Criteria

{{LIST VALIDATION CHECKS from PROGRESS.yaml}}

- [ ] {{VALIDATION_1}}
- [ ] {{VALIDATION_2}}
- [ ] {{VALIDATION_3}}

---

## Deliverables

{{LIST DELIVERABLES from PROGRESS.yaml}}

---

## Gotchas & Tips

### Common Pitfalls

- TODO: Add common pitfalls for this phase

### Tips

- TODO: Add helpful tips

---

## Testing Strategy

### Unit Tests

- TODO: Describe unit testing approach

### Integration Tests

- TODO: Describe integration testing approach

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate {{N}}

# Review plan before implementing
/workflow-temporal-plan-review {{N}}

# Review code after implementing
/workflow-temporal-review {{N}}
```
```

### 5. Write Phase Documentation

Write the generated phase document to:
`.claude/plans/WORKFLOW_TEMPORAL_PLAN/phases/PHASE_{{N}}_{{PHASE_SLUG}}.md`

### 6. Update PROGRESS.yaml

Update the planning status in PROGRESS.yaml:

```yaml
planning_status:
  phase_{{N}}_doc_created: true
  next_phase_to_document: {{N+1}} or null if all phases documented
```

### 7. Provide Summary

Show the user what was created:

```
✅ Phase {{N}} Documentation Created

════════════════════════════════════════════════════════

PHASE DETAILS
Name: {{PHASE_NAME}}
Category: {{CATEGORY}}
File: .claude/plans/WORKFLOW_TEMPORAL_PLAN/phases/PHASE_{{N}}_{{PHASE_SLUG}}.md
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

The phase documentation has been created with a template structure.

You should now:

1. Review the generated phase doc and fill in TODO sections:
   - Implementation guide code examples
   - Gotchas and tips
   - Testing strategy details

2. Run plan review before implementing:
   /workflow-temporal-plan-review {{N}}

3. When ready to execute, run: /workflow-temporal-next

4. To generate next phase doc, run: /workflow-temporal-build-phase

════════════════════════════════════════════════════════

PLANNING STATUS
Phases Documented: {{DOCUMENTED_COUNT}} / 13
Next Phase to Document: Phase {{NEXT_PHASE}} or "All phases documented!"
```

## Category-Specific Template Adjustments

### Backend Phases

Add these specific sections:
- Go package structure
- Business logic organization
- Testing with Go tests
- Temporal-specific patterns

### Infrastructure Phases

Add these specific sections:
- K8s manifest structure
- Makefile target descriptions
- Docker build steps
- Integration with dev-bounce

### Research Phases

Add these specific sections:
- Evaluation criteria
- Decision matrix
- Recommendation format
- Documentation location

### Testing Phases

Add these specific sections:
- Test file structure
- Test coverage goals
- Mock data and fixtures
- Determinism testing approach

### Deployment Phases

Add these specific sections:
- Deployment verification steps
- Rollback procedures
- Health check configuration
- Monitoring setup

## Tips

- Keep questions concise and focused
- Provide good defaults from PROGRESS.yaml
- Generate useful placeholder text for TODO sections
- Make it easy for user to refine the documentation later
- Update PROGRESS.yaml immediately
- Congratulate user on progress!

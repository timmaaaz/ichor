# Workflow Temporal Implementation Plan Review Command

Review and grade a phase plan document BEFORE implementation.

## Overview

This command reviews phase **plan documentation** (not code) to ensure plans are:
- Complete and well-structured
- Technically sound
- Ready for implementation

**Key Difference from Code Review**:
- `/workflow-temporal-review` reviews **implementation code** after coding
- `/workflow-temporal-plan-review` reviews **plan documents** before coding

## Your Task

### 1. Parse Parameters

This command takes an optional phase number parameter: `/workflow-temporal-plan-review [N]`

- If phase number provided: Review that specific phase's plan document
- If no parameter: Review the current phase plan (from PROGRESS.yaml `current_phase`)

### 2. Check Plan Document Exists

1. Read `.claude/plans/WORKFLOW_TEMPORAL_PLAN/PROGRESS.yaml`
2. Find the specified phase
3. Look for `.claude/plans/WORKFLOW_TEMPORAL_PLAN/phases/PHASE_{{N}}_*.md`

If no plan document exists:
```
No plan document found for Phase {{N}}: {{PHASE_NAME}}

To create the plan document, run:
  /workflow-temporal-build-phase {{N}}

Plan review cannot proceed without a plan document.
```

### 3. Check Previous Plan Review Status

Before proceeding, check if this phase plan has been reviewed before:

**If `plan_reviewed: true`:**
- Display the previous grade: "This plan was previously reviewed with grade: {{plan_review_grade}}"
- Note that this is a re-review
- The new grade will replace the old grade

**Grade Scale Reference:**
| Grade | Meaning | Action Needed |
|-------|---------|---------------|
| A, A- | Excellent plan | Ready to implement |
| B+, B | Good plan | Minor improvements optional |
| B-, C+ | Acceptable | Address concerns before implementing |
| C, C- | Needs Work | Revise plan before implementing |
| D or below | Poor | Significant rework required |

**Re-review Triggers:**
- Previous grade was B- or below
- Significant plan changes since last review
- User explicitly requests re-review
- Critical issues were found but not yet addressed

### 4. Determine Review Agent

Based on the phase `category`, select the appropriate agent:

| Category | Agent(s) to Use |
|----------|----------------|
| `backend` | `go-service-reviewer` |
| `infrastructure` | Manual review (no agent) |
| `research` | Manual review (no agent) |
| `testing` | `go-service-reviewer` (for Go tests) |
| `deployment` | Manual review (no agent) |

### 5. Gather Review Context

Collect information to provide to the review agent:

**Phase Context**:
- Phase number and name
- Phase category
- Phase goals and objectives (from plan document)
- Dependencies listed in the plan

**Plan Document to Review**:
- The entire `.claude/plans/WORKFLOW_TEMPORAL_PLAN/phases/PHASE_{{N}}_*.md` file

### 6. Execute Plan Review

#### For Backend Phases

Spawn `go-service-reviewer` agent with prompt:

```
Please review this PLAN DOCUMENT for Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}.

NOTE: You are reviewing the PLAN, not implementation code. The implementation has not been written yet.

Plan Document:
{{PLAN_DOCUMENT_CONTENT}}

This is a Temporal workflow implementation. Please evaluate this plan for:

1. **Completeness**
   - Are all tasks clearly defined?
   - Are deliverables specified with file paths?
   - Are validation criteria testable?
   - Are success metrics measurable?

2. **Technical Soundness**
   - Does the proposed approach follow Ardan Labs service architecture patterns?
   - Does the plan account for Temporal determinism requirements?
   - Are the proposed file structures appropriate?
   - Are the proposed interfaces well-designed?
   - Do code examples (if any) follow Go best practices?

3. **Temporal-Specific Considerations**
   - Does the plan address determinism requirements?
   - Are activity vs workflow responsibilities clear?
   - Are timeout/retry policies considered?
   - Are payload size limits addressed?

4. **Dependencies**
   - Are prerequisites correctly identified?
   - Are internal phase dependencies listed?
   - Are external dependencies documented?

5. **Risk Management**
   - Are gotchas and common pitfalls documented?
   - Are performance considerations addressed?
   - Are security considerations addressed?

6. **Clarity**
   - Is the scope well-defined?
   - Is the plan unambiguous?
   - Could another developer implement from this plan?

Provide specific, actionable feedback on how to improve the plan BEFORE implementation begins.
```

#### For Research/Infrastructure/Deployment Phases

Provide manual review guidance:

```
Manual Plan Review Checklist for Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}

COMPLETENESS
- [ ] All tasks are clearly defined with specific steps
- [ ] Deliverables are listed with file paths
- [ ] Validation criteria are testable
- [ ] Success metrics are measurable

TECHNICAL SOUNDNESS
- [ ] Proposed approach is appropriate for the problem
- [ ] Architecture decisions are justified
- [ ] Code examples (if any) follow project conventions
- [ ] Edge cases are considered

DEPENDENCIES
- [ ] Prerequisites are correctly identified
- [ ] Internal phase dependencies are listed
- [ ] External dependencies are documented
- [ ] Blocking conditions are clear

RISK MANAGEMENT
- [ ] Gotchas and common pitfalls are documented
- [ ] Performance considerations are addressed
- [ ] Security considerations are addressed (if applicable)
- [ ] Rollback/recovery scenarios considered (if applicable)

CLARITY
- [ ] Scope is well-defined (what IS and IS NOT included)
- [ ] Plan is unambiguous
- [ ] Another developer could implement from this plan
- [ ] Terminology is consistent

Please review manually and provide feedback.
```

### 7. Present Review Results

After agent(s) complete, present results in structured format:

```
Plan Review Results - Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}

══════════════════════════════════════════════════════════════

REVIEW SCOPE
Category: {{CATEGORY}}
Reviewer(s): {{AGENTS_USED}}
Plan Document: .claude/plans/WORKFLOW_TEMPORAL_PLAN/phases/PHASE_{{N}}_{{NAME}}.md
Previous Review: {{PREVIOUS_GRADE or "None"}}

══════════════════════════════════════════════════════════════

GRADE: {{NEW_GRADE}}
{{If re-review: "Previous: {{PREVIOUS_GRADE}} → New: {{NEW_GRADE}} ({{IMPROVED/UNCHANGED/REGRESSED}})" }}

══════════════════════════════════════════════════════════════

FINDINGS

{{AGENT_FINDINGS}}

══════════════════════════════════════════════════════════════

SUMMARY

Strengths:
   {{POSITIVE_FEEDBACK}}

Improvements Needed:
   {{CONSTRUCTIVE_FEEDBACK}}

Critical Issues:
   {{CRITICAL_ISSUES}}

══════════════════════════════════════════════════════════════

RECOMMENDATIONS

{{RECOMMENDATIONS}}

══════════════════════════════════════════════════════════════

NEXT STEPS

{{If grade >= B:}}
1. Address any critical issues identified in the plan
2. Consider improvement suggestions
3. Proceed to implementation: /workflow-temporal-next

{{If grade < B:}}
1. Address critical issues in the plan document (required)
2. Implement improvement suggestions
3. Re-run plan review: /workflow-temporal-plan-review {{PHASE_NUMBER}}
4. Target grade: B+ or higher before implementing

{{Always:}}
- Plan document: .claude/plans/WORKFLOW_TEMPORAL_PLAN/phases/PHASE_{{N}}_*.md
- After implementation, run code review: /workflow-temporal-review {{PHASE_NUMBER}}
```

### 8. Update PROGRESS.yaml (Required)

**Required Updates** - Always update these fields on the reviewed phase:
- Set `plan_reviewed: true` on the phase that was reviewed
- Set `plan_review_grade` to the grade assigned by the reviewer (e.g., "A-", "B+", "C")

**Optional Updates** - Consider adding review notes:
- Add plan review findings to `context.decisions`
- Note significant plan changes in `context.recent_changes`
- Add any new blockers if critical issues found

## Plan Review Timing Guidelines

### When to Review Plans

**Best Times**:
- After creating phase documentation with `/workflow-temporal-build-phase`
- Before starting implementation with `/workflow-temporal-next`
- After making significant changes to the plan document
- When uncertain about architectural approach

**Not Necessary**:
- For trivial phases (simple, well-understood changes)
- For phases following well-established patterns
- When plan was recently reviewed and unchanged

### Recommended Workflow

```
1. Create phase plan: /workflow-temporal-build-phase
          ↓
2. Review plan: /workflow-temporal-plan-review
          ↓ (if grade < B, revise and re-review)
          ↓
3. Implement: /workflow-temporal-next
          ↓
4. Review code: /workflow-temporal-review
          ↓ (if grade < B, fix and re-review)
          ↓
5. Validate: /workflow-temporal-validate
```

## Agent-Specific Notes

### go-service-reviewer (for plans)

When reviewing plans, focuses on:
- Proposed architecture follows Ardan Labs patterns
- Business logic organization makes sense
- Error handling approach is appropriate
- Database access patterns are sound
- API handler structure is correct
- Security considerations are addressed
- **Temporal determinism requirements are addressed**

## Example Usage

```bash
# Review current phase's plan
/workflow-temporal-plan-review

# Review specific phase's plan
/workflow-temporal-plan-review 3

# After plan review passes, implement
/workflow-temporal-next

# After implementation, review code
/workflow-temporal-review 3
```

## Tips

- Review plans BEFORE implementing to catch issues early
- Plans with grade B+ or higher are ready to implement
- Plans with grade below B need revision before implementation
- Document key decisions from plan review in PROGRESS.yaml
- Use plan review as a learning opportunity
- It's much cheaper to fix issues in plans than in code
- A good plan leads to better implementation

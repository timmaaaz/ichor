# Workflow Action Gap Remediation Plan Review Command

Review and grade a phase plan document BEFORE implementation.

## Overview

This command reviews phase **plan documentation** (not code) to ensure plans are:
- Complete and well-structured
- Technically sound
- Ready for implementation

**Key Difference from Code Review**:
- `/workflow-gaps-review` reviews **implementation code** after coding
- `/workflow-gaps-plan-review` reviews **plan documents** before coding

## Your Task

### 1. Parse Parameters

This command takes an optional phase number parameter: `/workflow-gaps-plan-review [N]`

- If phase number provided: Review that specific phase's plan document
- If no parameter: Review the current phase plan (from PROGRESS.yaml `current_phase`)

### 2. Check Plan Document Exists

1. Read `.claude/plans/WORKFLOW_GAPS_PLAN/PROGRESS.yaml`
2. Find the specified phase
3. Look for `.claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_N_*.md`

If no plan document exists:
```
No plan document found for Phase N: [PHASE_NAME]

To create the plan document, run:
  /workflow-gaps-build-phase N

Plan review cannot proceed without a plan document.
```

### 3. Check Previous Plan Review Status

Before proceeding, check if this phase plan has been reviewed before:

**If `plan_reviewed: true`:**
- Display the previous grade: "This plan was previously reviewed with grade: [plan_review_grade]"
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

### 4. Execute Plan Review

All phases in this plan are `backend` category. Spawn `go-service-reviewer` agent with prompt:

```
Please review this PLAN DOCUMENT for Phase [PHASE_NUMBER]: [PHASE_NAME].

NOTE: You are reviewing the PLAN, not implementation code. The implementation has not been written yet.

Plan Document:
[PLAN_DOCUMENT_CONTENT]

Please evaluate this plan for:

1. **Completeness**
   - Are all tasks clearly defined?
   - Are deliverables specified with file paths?
   - Are validation criteria testable?
   - Are success metrics measurable?

2. **Technical Soundness**
   - Does the proposed approach follow Ardan Labs service architecture patterns?
   - Are the proposed file structures appropriate?
   - Are the proposed interfaces well-designed?
   - Do code examples (if any) follow Go best practices?

3. **Dependencies**
   - Are prerequisites correctly identified?
   - Are internal phase dependencies listed?
   - Are external dependencies documented?

4. **Risk Management**
   - Are gotchas and common pitfalls documented?
   - Are performance considerations addressed?
   - Are security considerations addressed?

5. **Clarity**
   - Is the scope well-defined?
   - Is the plan unambiguous?
   - Could another developer implement from this plan?

Provide specific, actionable feedback on how to improve the plan BEFORE implementation begins.
```

### 5. Present Review Results

After agent completes, present results in structured format:

```
Plan Review Results - Phase [PHASE_NUMBER]: [PHASE_NAME]

══════════════════════════════════════════════════════════════

REVIEW SCOPE
Category: backend
Reviewer(s): go-service-reviewer
Plan Document: .claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_N_[NAME].md
Previous Review: [PREVIOUS_GRADE or "None"]

══════════════════════════════════════════════════════════════

GRADE: [NEW_GRADE]
[If re-review: "Previous: [PREVIOUS_GRADE] → New: [NEW_GRADE] (IMPROVED/UNCHANGED/REGRESSED)"]

══════════════════════════════════════════════════════════════

FINDINGS

[AGENT_FINDINGS]

══════════════════════════════════════════════════════════════

SUMMARY

Strengths:
   [POSITIVE_FEEDBACK]

Improvements Needed:
   [CONSTRUCTIVE_FEEDBACK]

Critical Issues:
   [CRITICAL_ISSUES]

══════════════════════════════════════════════════════════════

RECOMMENDATIONS

[RECOMMENDATIONS]

══════════════════════════════════════════════════════════════

NEXT STEPS

[If grade >= B:]
1. Address any critical issues identified in the plan
2. Consider improvement suggestions
3. Proceed to implementation: /workflow-gaps-next

[If grade < B:]
1. Address critical issues in the plan document (required)
2. Implement improvement suggestions
3. Re-run plan review: /workflow-gaps-plan-review [PHASE_NUMBER]
4. Target grade: B+ or higher before implementing

Always:
- Plan document: .claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_N_*.md
- After implementation, run code review: /workflow-gaps-review [PHASE_NUMBER]
```

### 6. Update PROGRESS.yaml (Required)

**Required Updates** - Always update these fields on the reviewed phase:
- Set `plan_reviewed: true` on the phase that was reviewed
- Set `plan_review_grade` to the grade assigned by the reviewer (e.g., "A-", "B+", "C")

**Optional Updates** - Consider adding review notes:
- Add plan review findings to `context.decisions`
- Note significant plan changes in `context.recent_changes`
- Add any new blockers if critical issues found

## Recommended Workflow

```
1. Review plan: /workflow-gaps-plan-review N
          ↓ (if grade < B, revise and re-review)
          ↓
2. Implement: /workflow-gaps-next
          ↓
3. Review code: /workflow-gaps-review N
          ↓ (if grade < B, fix and re-review)
          ↓
4. Validate: /workflow-gaps-validate
```

## Example Usage

```bash
# Review current phase's plan
/workflow-gaps-plan-review

# Review specific phase's plan
/workflow-gaps-plan-review 5

# After plan review passes, implement
/workflow-gaps-next

# After implementation, review code
/workflow-gaps-review 5
```

## Tips

- All 9 phase plans have already been created — you can review them immediately
- Plans with grade B+ or higher are ready to implement
- Plans with grade below B need revision before implementation
- Document key decisions from plan review in PROGRESS.yaml
- It's much cheaper to fix issues in plans than in code

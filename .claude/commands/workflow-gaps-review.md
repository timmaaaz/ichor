# Workflow Action Gap Remediation Code Review Command

Manually trigger code review for a specific phase using appropriate specialized agents.

## Your Task

### 1. Parse Parameters

This command takes an optional phase number parameter: `/workflow-gaps-review [N]`

- If phase number provided: Review that specific phase
- If no parameter: Review the current phase (from PROGRESS.yaml `current_phase`)

### 2. Read Phase Information

1. Read `.claude/plans/WORKFLOW_GAPS_PLAN/PROGRESS.yaml`
2. Find the specified phase
3. Get the phase `category` field
4. Get the phase `name` and `description`
5. Get the phase `reviewed` and `review_grade` fields (if present)
6. Determine which files were created/modified in this phase

### 2.5. Check Previous Review Status

Before proceeding, check if this phase has been reviewed before:

**If `reviewed: true`:**
- Display the previous grade: "This phase was previously reviewed with grade: [review_grade]"
- Note that this is a re-review
- The new grade will replace the old grade

**Grade Scale Reference:**
| Grade | Meaning | Action Needed |
|-------|---------|---------------|
| A, A- | Excellent | No re-review needed unless major changes |
| B+, B | Good | Consider re-review after fixes |
| B-, C+ | Acceptable | Re-review recommended after improvements |
| C, C- | Needs Work | Re-review required after fixes |
| D or below | Poor | Must re-review after significant rework |

### 3. Determine Review Agent

All phases in this plan are `backend` category (Go code):

| Category | Agent(s) to Use |
|----------|----------------|
| `backend` | `go-service-reviewer` |
| `database` | Manual review (no agent) |

### 4. Gather Review Context

Collect information to provide to the review agent:

**Phase Context**:
- Phase number and name
- Phase goals and objectives
- Files created/modified in this phase
- Any architectural decisions made

**Code to Review**:
- List of files from phase tasks in PROGRESS.yaml
- Read the phase documentation for context

### 5. Execute Code Review

#### For Backend Phases (all phases in this plan)

Spawn `go-service-reviewer` agent with prompt:

```
Please review the Go backend code for Phase [PHASE_NUMBER]: [PHASE_NAME].

Phase Context:
- Goal: [PHASE_GOAL]
- Category: backend
- Files to review:
  [FILE_LIST]

Phase Documentation: [Read from .claude/plans/WORKFLOW_GAPS_PLAN/phases/PHASE_N_*.md]

Please review for:
1. Adherence to Ardan Labs service architecture patterns
2. Idiomatic Go code
3. Proper error handling
4. Security considerations
5. Performance implications
6. Code structure and organization
7. Documentation and comments
8. Interface design and abstractions
9. Test coverage adequacy

Provide specific, actionable feedback.
```

#### For Phase 5 (seek_approval - database + backend)

Provide manual review guidance for the migration in addition to go-service-reviewer:

```
Database Migration Manual Review:
- [ ] Migration is reversible (has DROP TABLE in rollback)
- [ ] Schema changes documented
- [ ] Foreign keys and indexes appropriate
- [ ] UUID[] column uses pq.Array() correctly
- [ ] Performance considerations addressed
```

### 6. Present Review Results

After agent(s) complete, present results in structured format:

```
🔍 Code Review Results - Phase [PHASE_NUMBER]: [PHASE_NAME]

══════════════════════════════════════════════════════════════

REVIEW SCOPE
Category: backend
Reviewer(s): go-service-reviewer
Files Reviewed: [FILE_COUNT]
Previous Review: [PREVIOUS_GRADE or "None"]

══════════════════════════════════════════════════════════════

GRADE: [NEW_GRADE]
[If re-review: "Previous: [PREVIOUS_GRADE] → New: [NEW_GRADE] (IMPROVED/UNCHANGED/REGRESSED)"]

══════════════════════════════════════════════════════════════

FINDINGS

[AGENT_FINDINGS]

══════════════════════════════════════════════════════════════

SUMMARY

✅ Strengths:
   [POSITIVE_FEEDBACK]

⚠️  Improvements Needed:
   [CONSTRUCTIVE_FEEDBACK]

🚨 Critical Issues:
   [CRITICAL_ISSUES]

══════════════════════════════════════════════════════════════

RECOMMENDATIONS

[RECOMMENDATIONS]

══════════════════════════════════════════════════════════════

NEXT STEPS

[If grade >= B:]
1. Address any critical issues identified
2. Consider improvement suggestions for future iterations
3. Proceed to next phase: /workflow-gaps-next

[If grade < B:]
1. Address critical issues (required before proceeding)
2. Implement improvement suggestions
3. Re-run review: /workflow-gaps-review [PHASE_NUMBER]
4. Target grade: B+ or higher before proceeding

Always:
- Run /workflow-gaps-validate to verify changes
```

### 7. Update PROGRESS.yaml (Required)

**Required Updates** - Always update these fields on the reviewed phase:
- Set `reviewed: true` on the phase that was reviewed
- Set `review_grade` to the grade assigned by the reviewer (e.g., "A-", "B+", "C")

**Optional Updates** - Consider adding review notes:
- Add review findings to `context.decisions`
- Note significant refactorings in `context.recent_changes`
- Add any new blockers if critical issues found

## Tips

- Reviews are most valuable for complex phases (Phase 5, 6, 7 especially)
- Act on critical issues immediately
- Consider suggestions for future refactoring
- Document key decisions from review in PROGRESS.yaml
- Don't block progress on minor style suggestions
- **Target grade B+ or higher** before moving to the next phase
- Grades below B should trigger a re-review after fixes

## Example Usage

```bash
# Review current phase
/workflow-gaps-review

# Review specific phase
/workflow-gaps-review 5

# After review, address feedback and validate
/workflow-gaps-validate
```

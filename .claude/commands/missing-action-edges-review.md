# Universal Action Edge Enforcement Code Review Command

Manually trigger code review for a specific phase using appropriate specialized agents.

## Your Task

### 1. Parse Parameters

This command takes an optional phase number parameter: `/missing-action-edges-review [N]`

- If phase number provided: Review that specific phase
- If no parameter: Review the current phase (from PROGRESS.yaml `current_phase`)

### 2. Read Phase Information

1. Read `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
2. Find the specified phase
3. Get the phase `category` field
4. Get the phase `name` and `description`
5. Get the phase `reviewed` and `review_grade` fields (if present)
6. Determine which files were created/modified in this phase

### 2.5. Check Previous Review Status

Before proceeding, check if this phase has been reviewed before:

**If `reviewed: true`:**
- Display the previous grade: "This phase was previously reviewed with grade: {{review_grade}}"
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

**Re-review Triggers:**
- Previous grade was B- or below
- Significant code changes since last review
- User explicitly requests re-review
- Critical issues were found but not yet addressed

### 3. Determine Review Agent

Based on the phase `category`, select the appropriate agent:

| Category | Agent(s) to Use |
|----------|----------------|
| `backend` | `go-service-reviewer` |
| `frontend` | `vue3-best-practices` |
| `database` | Manual review (no agent) |
| `fullstack` | Both `go-service-reviewer` AND `vue3-best-practices` |
| `testing` | `go-service-reviewer` (for Go tests) or manual |
| `documentation` | Manual review (no agent) |

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

#### For Backend Phases

Spawn `go-service-reviewer` agent with prompt:

```
Please review the Go backend code for Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}.

Phase Context:
- Goal: {{PHASE_GOAL}}
- Category: backend
- Files to review:
  {{FILE_LIST}}

Phase Documentation: [Read from .claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_{{N}}_*.md]

Please review for:
1. Adherence to Ardan Labs service architecture patterns
2. Idiomatic Go code
3. Proper error handling
4. Security considerations
5. Performance implications
6. Code structure and organization
7. Documentation and comments

Provide specific, actionable feedback.
```

#### For Testing Phases

Spawn `go-service-reviewer` agent with prompt focusing on test quality:

```
Please review the test code for Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}.

Phase Context:
- Goal: {{PHASE_GOAL}}
- Category: testing
- Files to review:
  {{FILE_LIST}}

Please review for:
1. Test coverage completeness
2. Test naming conventions
3. Proper test isolation
4. Edge case coverage
5. Test data management
6. Assertion clarity
7. Following project test patterns

Provide specific, actionable feedback.
```

#### For Documentation Phases

Provide manual review guidance:

```
📋 Manual Review Checklist for Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}

Documentation Review:
- [ ] Documentation is clear and accurate
- [ ] No references to removed features (execution_order, linear execution)
- [ ] Examples are correct and runnable
- [ ] API documentation matches implementation
- [ ] Breaking changes highlighted
- [ ] Migration guides provided if needed

Please review manually and provide feedback.
```

### 6. Present Review Results

After agent(s) complete, present results in structured format:

```
🔍 Code Review Results - Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}

══════════════════════════════════════════════════════════════

REVIEW SCOPE
Category: {{CATEGORY}}
Reviewer(s): {{AGENTS_USED}}
Files Reviewed: {{FILE_COUNT}}
Previous Review: {{PREVIOUS_GRADE or "None"}}

══════════════════════════════════════════════════════════════

GRADE: {{NEW_GRADE}}
{{If re-review: "Previous: {{PREVIOUS_GRADE}} → New: {{NEW_GRADE}} ({{IMPROVED/UNCHANGED/REGRESSED}})" }}

══════════════════════════════════════════════════════════════

FINDINGS

{{AGENT_FINDINGS}}

══════════════════════════════════════════════════════════════

SUMMARY

✅ Strengths:
   {{POSITIVE_FEEDBACK}}

⚠️  Improvements Needed:
   {{CONSTRUCTIVE_FEEDBACK}}

🚨 Critical Issues:
   {{CRITICAL_ISSUES}}

══════════════════════════════════════════════════════════════

RECOMMENDATIONS

{{RECOMMENDATIONS}}

══════════════════════════════════════════════════════════════

NEXT STEPS

{{If grade >= B:}}
1. Address any critical issues identified
2. Consider improvement suggestions for future iterations
3. Proceed to next phase: /missing-action-edges-next

{{If grade < B:}}
1. Address critical issues (required before proceeding)
2. Implement improvement suggestions
3. Re-run review: /missing-action-edges-review {{PHASE_NUMBER}}
4. Target grade: B+ or higher before proceeding

{{Always:}}
- Run /missing-action-edges-validate to verify changes
```

### 7. Update PROGRESS.yaml (Required)

**Required Updates** - Always update these fields on the reviewed phase:
- Set `reviewed: true` on the phase that was reviewed
- Set `review_grade` to the grade assigned by the reviewer (e.g., "A-", "B+", "C")

**Optional Updates** - Consider adding review notes:
- Add review findings to `context.decisions`
- Note significant refactorings in `context.recent_changes`
- Add any new blockers if critical issues found
- Add detailed review notes to `context.code_review_notes` section

## Review Timing Guidelines

### When to Review

**Best Times**:
- After phase completion, before moving to next phase
- After implementing a complex or critical feature
- Before merging to main branch
- When uncertain about architectural decisions

**Not Necessary**:
- For trivial changes (typo fixes, documentation updates)
- For changes following well-established patterns
- When time-constrained and code is straightforward

### Review Scope

**Full Phase Review**: Review all files created/modified in the phase

**Targeted Review**: Review specific files if:
- Only certain files are complex
- Some files follow established patterns (skip those)
- User requests review of specific concerns

## Agent-Specific Notes

### go-service-reviewer

Focuses on:
- Ardan Labs service architecture patterns
- Business logic organization
- Error handling
- Database access patterns
- API handler structure
- Security considerations

## Example Usage

```bash
# Review current phase
/missing-action-edges-review

# Review specific phase
/missing-action-edges-review 3

# After review, address feedback and validate
/missing-action-edges-validate
```

## Tips

- Reviews are most valuable for complex phases
- Act on critical issues immediately
- Consider suggestions for future refactoring
- Document key decisions from review in PROGRESS.yaml
- Don't block progress on minor style suggestions
- Use reviews as learning opportunities
- **Target grade B+ or higher** before moving to the next phase
- Grades below B should trigger a re-review after fixes
- Track grade trends across phases to identify patterns

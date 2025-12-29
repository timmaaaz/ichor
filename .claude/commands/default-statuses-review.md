# Default Status Management Code Review Command

Manually trigger code review for a specific phase using appropriate specialized agents.

## Your Task

### 1. Parse Parameters

This command takes an optional phase number parameter: `/default-statuses-review [N]`

- If phase number provided: Review that specific phase
- If no parameter: Review the current phase (from PROGRESS.yaml `current_phase`)

### 2. Read Phase Information

1. Read `.claude/plans/DEFAULT_STATUSES_PLAN/PROGRESS.yaml`
2. Find the specified phase
3. Get the phase `category` field
4. Get the phase `name` and `description`
5. Determine which files were created/modified in this phase

### 3. Determine Review Agent

Based on the phase `category`, select the appropriate agent:

| Category | Agent(s) to Use |
|----------|----------------|
| `backend` | `go-service-reviewer` |
| `frontend` | `vue3-best-practices` |
| `fullstack` | Both `go-service-reviewer` AND `vue3-best-practices` |

**Phase Categories for This Plan:**
- Phase 1: `backend` → use `go-service-reviewer`
- Phase 2: `backend` → use `go-service-reviewer`
- Phase 3: `fullstack` → use both agents

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

#### For Backend Phases (1 & 2)

Spawn `go-service-reviewer` agent with prompt:

```
Please review the Go backend code for Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}.

Phase Context:
- Goal: {{PHASE_GOAL}}
- Category: backend
- Files to review:
  - app/domain/formdata/formdataapp/formdataapp.go
  - business/sdk/dbtest/seedmodels/forms.go
  - business/sdk/dbtest/seedmodels/tableforms.go

Please review for:
1. Adherence to Ardan Labs service architecture patterns
2. Idiomatic Go code
3. Proper error handling
4. Security considerations
5. Performance implications
6. Code structure and organization

Provide specific, actionable feedback.
```

#### For Fullstack Phases (3)

Run both reviews sequentially:
1. First spawn `go-service-reviewer` for backend files
2. Then spawn `vue3-best-practices` for frontend files (if any)
3. Combine results

### 6. Present Review Results

After agent(s) complete, present results in structured format:

```
🔍 Code Review Results - Phase {{PHASE_NUMBER}}: {{PHASE_NAME}}

══════════════════════════════════════════════════════════════

REVIEW SCOPE
Category: {{CATEGORY}}
Reviewer(s): {{AGENTS_USED}}
Files Reviewed: {{FILE_COUNT}}

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

1. Address critical issues if any
2. Consider improvement suggestions
3. Update code based on feedback
4. Re-run /default-statuses-review {{PHASE_NUMBER}} if significant changes made
5. Run /default-statuses-validate to verify changes
```

### 7. Update PROGRESS.yaml (Optional)

Consider adding review notes to PROGRESS.yaml:
- Add review findings to `context.decisions`
- Note significant refactorings in `context.recent_changes`
- Add any new blockers if critical issues found

## Review Timing Guidelines

### When to Review

**Best Times**:
- After phase completion, before moving to next phase
- After implementing a complex or critical feature
- When uncertain about architectural decisions

**Not Necessary**:
- For trivial changes (typo fixes, documentation updates)
- For changes following well-established patterns

## Example Usage

```bash
# Review current phase
/default-statuses-review

# Review specific phase
/default-statuses-review 1

# After review, address feedback and validate
/default-statuses-validate
```

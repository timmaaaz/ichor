# Quick Status Command

Display a compact overview of all phases showing execution status, plan review grade, and code review grade.

## Your Task

Run this awk command to extract phase status from PROGRESS.yaml:

```bash
awk '
/^  - phase:/ { phase=$3; gsub(/"/, "", phase) }
/^    name:/ { gsub(/^    name: /, ""); gsub(/"/, ""); name=$0 }
/^    status:/ { gsub(/^    status: /, ""); gsub(/"/, ""); status=$1 }
/^    plan_reviewed:/ { plan_reviewed=$2 }
/^    plan_review_grade:/ { gsub(/^    plan_review_grade: /, ""); gsub(/"/, ""); plan_grade=$1 }
/^    reviewed:/ { reviewed=$2 }
/^    review_grade:/ {
  gsub(/^    review_grade: /, ""); gsub(/"/, ""); review_grade=$1
  if (phase != "") {
    printf "%-5s %-35s %-12s %-6s %-4s %-6s %-4s\n", phase, substr(name,1,35), status, plan_reviewed, plan_grade, reviewed, review_grade
    phase=""
  }
}
' .claude/plans/WORKFLOW_TEMPORAL_PLAN/PROGRESS.yaml
```

## Output Format

Present the results as a formatted table with a header:

```
## Workflow Temporal Implementation - Quick Status

| Phase | Name                                | Execution   | Plan Rev | Grade | Code Rev | Grade |
|-------|-------------------------------------|-------------|----------|-------|----------|-------|
| 1     | Infrastructure Setup                | pending     | false    | null  | false    | null  |
| 2     | Temporalgraph Evaluation            | pending     | false    | null  | false    | null  |
| 3     | Core Models & Context               | pending     | false    | null  | false    | null  |
| 4     | Graph Executor                      | pending     | false    | null  | false    | null  |
| 5     | Workflow Implementation             | pending     | false    | null  | false    | null  |
| 6     | Activities & Async                  | pending     | false    | null  | false    | null  |
| 7     | Trigger System                      | pending     | false    | null  | false    | null  |
| 8     | Edge Store Adapter                  | pending     | false    | null  | false    | null  |
| 9     | Worker Service & Wiring             | pending     | false    | null  | false    | null  |
| 10    | Graph Executor Unit Tests           | pending     | false    | null  | false    | null  |
| 11    | Workflow Integration Tests          | pending     | false    | null  | false    | null  |
| 12    | Edge Case & Limit Tests             | pending     | false    | null  | false    | null  |
| 13    | Kubernetes Deployment               | pending     | false    | null  | false    | null  |
```

Then add a brief summary:

```
---
**Summary:**
- Executed: X/13 phases completed
- Plan Reviews: X/13 completed (Z pending)
- Code Reviews: X/13 completed (Z pending)

**Next Actions:**
- Phase 1 needs: [plan review / implementation / code review]
- To continue: /workflow-temporal-next
```

## Legend

- **Execution**: `completed`, `in_progress`, `pending`, `blocked`
- **Plan Rev**: Whether the phase plan has been reviewed (`true`/`false`)
- **Code Rev**: Whether the implemented code has been reviewed (`true`/`false`)
- **Grade**: Review grade (A, A-, B+, B, etc.) or `null` if not reviewed

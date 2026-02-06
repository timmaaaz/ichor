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
' .claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml
```

## Output Format

Present the results as a formatted table with a header:

```
## Universal Action Edge Enforcement - Quick Status

| Phase | Name                                | Execution   | Plan Rev | Grade | Code Rev | Grade |
|-------|-------------------------------------|-------------|----------|-------|----------|-------|
| 1     | Validation Layer Changes            | completed   | true     | A     | true     | A     |
...
```

Then add a brief summary:

```
---
**Summary:**
- Executed: X/Y phases completed
- Plan Reviews: X/Y completed (Z pending)
- Code Reviews: X/Y completed (Z pending)

**Next Actions:**
- [List phases needing code review]
- [List phases needing plan building]
```

## Legend

- **Execution**: `completed`, `in_progress`, `pending`, `blocked`
- **Plan Rev**: Whether the phase plan has been reviewed (`true`/`false`)
- **Code Rev**: Whether the implemented code has been reviewed (`true`/`false`)
- **Grade**: Review grade (A, A-, B+, B, etc.) or `null` if not reviewed

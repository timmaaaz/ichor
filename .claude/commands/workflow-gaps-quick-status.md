# Workflow Action Gap Remediation Quick Status

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
' .claude/plans/WORKFLOW_GAPS_PLAN/PROGRESS.yaml
```

## Output Format

Present the results as a formatted table with a header:

```
## Workflow Action Gap Remediation - Quick Status

| Phase | Name                                | Execution   | Plan Rev | Grade | Code Rev | Grade |
|-------|-------------------------------------|-------------|----------|-------|----------|-------|
| 1     | Add Missing Tables to Whitelist     | pending     | false    | null  | false    | null  |
| 2     | Fix FieldChanges Propagation        | pending     | false    | null  | false    | null  |
| 3     | Implement send_notification         | pending     | false    | null  | false    | null  |
| 4     | Implement send_email                | pending     | false    | null  | false    | null  |
| 5     | Implement seek_approval             | pending     | false    | null  | false    | null  |
| 6     | Add create_purchase_order           | pending     | false    | null  | false    | null  |
| 7     | Add receive_inventory               | pending     | false    | null  | false    | null  |
| 8     | Add call_webhook                    | pending     | false    | null  | false    | null  |
| 9     | Add Template Arithmetic             | pending     | false    | null  | false    | null  |
```

Then add a brief summary:

```
---
**Summary:**
- Executed: X/9 phases completed
- Plan Reviews: X/9 completed (Y pending)
- Code Reviews: X/9 completed (Y pending)

**Next Actions:**
- Review plan before implementing: /workflow-gaps-plan-review 1
- Start first phase: /workflow-gaps-next
- Jump to specific phase: /workflow-gaps-phase N
```

## Legend

- **Execution**: `completed`, `in_progress`, `pending`, `blocked`
- **Plan Rev**: Whether the phase plan has been reviewed (`true`/`false`)
- **Code Rev**: Whether the implemented code has been reviewed (`true`/`false`)
- **Grade**: Review grade (A, A-, B+, B, etc.) or `null` if not reviewed

## Phase Reference

| # | Name | Effort | Category |
|---|------|--------|----------|
| 1 | Add Missing Tables to Whitelist | Quick (10 min) | backend |
| 2 | Fix FieldChanges Propagation | Quick (30 min) | backend |
| 3 | Implement send_notification | Medium (1h) | backend |
| 4 | Implement send_email | Medium (2h) | backend |
| 5 | Implement seek_approval | High (8h) | backend+database |
| 6 | Add create_purchase_order | High (4h) | backend |
| 7 | Add receive_inventory | Medium (3h) | backend |
| 8 | Add call_webhook | Medium (2h) | backend |
| 9 | Add Template Arithmetic | Medium (1h) | backend |

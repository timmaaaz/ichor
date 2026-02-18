# Workflow Temporal Implementation Status Command

Read `.claude/plans/WORKFLOW_TEMPORAL_PLAN/PROGRESS.yaml` and display a comprehensive status report of the Workflow Temporal Implementation progress.

## Your Task

1. Read the PROGRESS.yaml file from `.claude/plans/WORKFLOW_TEMPORAL_PLAN/PROGRESS.yaml`
2. Display a formatted status report including:
   - Overall project status and current phase
   - Summary metrics (phases completed, files created/modified, etc.)
   - Planning status (which phase docs have been created)
   - Status of all phases with:
     - Phase number, name, and status
     - Category (backend/frontend/database/fullstack/testing/documentation)
     - Number of tasks completed/total
     - Any blockers
   - Current focus and next task from context section
   - Any active blockers across all phases
   - Milestones achieved
   - Dependencies (internal and external)

## Output Format

Use a clear, hierarchical format with:
- Emoji status indicators:
  - ✅ `completed`
  - 🔄 `in_progress`
  - ⏳ `pending`
  - 🚫 `blocked`
  - 📝 `planning`
- Progress bars or percentages where applicable
- Highlight the current phase and next actionable task
- List any blockers prominently
- Show planning progress (which phase docs exist)

### Example Output Structure

```
📊 Workflow Temporal Implementation - Status Report

══════════════════════════════════════════════════════════════

PROJECT OVERVIEW
Status: 🔄 In Progress
Current Phase: 3 / 13
Progress: [████████░░░░░░░░] 37.5%

SUMMARY
✅ Phases Completed: 2 / 13 (25%)
🔄 Phases In Progress: 1
⏳ Phases Pending: 5
👀 Phases Reviewed: X / Y completed (with grades)
📁 Files Created: 12
📝 Files Modified: 2

PLANNING STATUS
✅ Phase 1 Documentation Created
✅ Phase 2 Documentation Created
✅ Phase 3 Documentation Created
⏳ Phase 4 Documentation Pending
⏳ Phase 5 Documentation Pending

══════════════════════════════════════════════════════════════

PHASE BREAKDOWN

Phase 1: [Phase Name]
Status: ✅ Completed | 👀 Reviewed (B+)
Category: backend
Tasks: 7/7 completed (100%)

Phase 2: [Phase Name]
Status: ✅ Completed
Category: frontend
Tasks: 6/6 completed (100%)
Note: Not yet reviewed

Phase 3: [Phase Name]
Status: 🔄 In Progress
Category: fullstack
Tasks: 3/6 completed (50%)
Current Task: [Task description]

Phase 4: [Phase Name]
Status: ⏳ Pending
Category: frontend

══════════════════════════════════════════════════════════════

DEPENDENCIES

Internal Dependencies:
  Phase 3 depends on: Phase 1, Phase 2 ✅

External Dependencies:
  Depends on: BACKEND_API_PLAN ✅

══════════════════════════════════════════════════════════════

CURRENT FOCUS
Working on: Phase 3 - [Phase Name]
Next Task: [Next task description]
Recent Changes:
  - [Change 1]
  - [Change 2]
  - [Change 3]

Key Decisions:
  - [Decision 1]
  - [Decision 2]

══════════════════════════════════════════════════════════════

BLOCKERS
None currently

══════════════════════════════════════════════════════════════

MILESTONES
✅ Planning Complete ({{DATE}})
✅ Phase 1 Complete ({{DATE}})
✅ Phase 2 Complete ({{DATE}})
⏳ Phase 3 Complete
⏳ Phase 4 Complete
⏳ Project Complete

══════════════════════════════════════════════════════════════

NEXT STEPS
1. Complete Phase 3 (3 tasks remaining)
2. Run /workflow-temporal-review 3 to get code review (optional)
3. Run /workflow-temporal-validate to check Phase 3 completion criteria
4. Run /workflow-temporal-next to continue implementation
```

## Tips

- Be concise but comprehensive
- Use visual hierarchy (boxes, spacing, emojis)
- Highlight actionable items
- Show both high-level overview and detailed breakdown
- If blockers exist, make them very visible
- Include helpful next steps at the end
- Note phase categories for context

## Review Status Display

For each phase, check `reviewed` and `review_grade` fields:
- If `reviewed: true`: Show "👀 Reviewed ({{grade}})" after status
- If `reviewed: false` and status is `completed`: Show "Note: Not yet reviewed"
- Grades B- or below should be highlighted as needing re-review

In the summary section, show:
- Total phases reviewed vs completed
- Average grade (if tracking)
- Any phases needing re-review (grade < B)

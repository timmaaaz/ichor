# Universal Action Edge Enforcement Status Command

Read `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml` and display a comprehensive status report of the Universal Action Edge Enforcement implementation progress.

## Your Task

1. Read the PROGRESS.yaml file from `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
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
📊 Universal Action Edge Enforcement - Status Report

══════════════════════════════════════════════════════════════

PROJECT OVERVIEW
Status: 🔄 In Progress
Current Phase: 3 / 6
Progress: [████████░░░░░░░░] 33%

SUMMARY
✅ Phases Completed: 2 / 6 (33%)
🔄 Phases In Progress: 1
⏳ Phases Pending: 3
👀 Phases Reviewed: X / Y completed (with grades)
📁 Files Created: 12
📝 Files Modified: 2

PLANNING STATUS
✅ Phase 1 Documentation Created
✅ Phase 2 Documentation Created
✅ Phase 3 Documentation Created
⏳ Phase 4 Documentation Pending
⏳ Phase 5 Documentation Pending
⏳ Phase 6 Documentation Pending

══════════════════════════════════════════════════════════════

PHASE BREAKDOWN

Phase 1: Validation Layer Changes
Status: ✅ Completed | 👀 Reviewed (B+)
Category: backend
Tasks: 1/1 completed (100%)

Phase 2: Remove execution_order Field
Status: ✅ Completed
Category: backend
Tasks: 6/6 completed (100%)
Note: Not yet reviewed

Phase 3: Remove Linear Executor
Status: 🔄 In Progress
Category: backend
Tasks: 1/3 completed (33%)
Current Task: Delete ExecuteRuleActions() function

Phase 4: Test Updates
Status: ⏳ Pending
Category: testing

Phase 5: Seed Data Updates
Status: ⏳ Pending
Category: backend

Phase 6: Documentation Updates
Status: ⏳ Pending
Category: documentation

══════════════════════════════════════════════════════════════

DEPENDENCIES

Internal Dependencies:
  Phase 4 depends on: Phase 2, Phase 3 ✅

External Dependencies:
  None

══════════════════════════════════════════════════════════════

CURRENT FOCUS
Working on: Phase 3 - Remove Linear Executor
Next Task: Delete ExecuteRuleActions() function
Recent Changes:
  - Removed execution_order from all models
  - Added database migration

Key Decisions:
  - Require edges universally (Option B)
  - Remove execution_order field entirely

══════════════════════════════════════════════════════════════

BLOCKERS
None currently

══════════════════════════════════════════════════════════════

MILESTONES
✅ Planning Complete (2026-02-05)
⏳ Phase 1 Complete (Validation)
⏳ Phase 2-3 Complete (Remove Old Code)
⏳ Phase 4-5 Complete (Tests & Seeds)
⏳ Phase 6 Complete (Documentation)
⏳ Project Complete

══════════════════════════════════════════════════════════════

NEXT STEPS
1. Complete Phase 3 (2 tasks remaining)
2. Run /missing-action-edges-review 3 to get code review (optional)
3. Run /missing-action-edges-validate to check Phase 3 completion criteria
4. Run /missing-action-edges-next to continue implementation
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


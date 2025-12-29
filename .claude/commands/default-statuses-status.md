# Default Status Management Status Command

Read `.claude/plans/DEFAULT_STATUSES_PLAN/PROGRESS.yaml` and display a comprehensive status report of the Default Status Management implementation progress.

## Your Task

1. Read the PROGRESS.yaml file from `.claude/plans/DEFAULT_STATUSES_PLAN/PROGRESS.yaml`
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
📊 Default Status Management - Status Report

══════════════════════════════════════════════════════════════

PROJECT OVERVIEW
Status: 🔄 In Progress
Current Phase: 1 / 3
Progress: [████████░░░░░░░░] 33%

SUMMARY
✅ Phases Completed: 0 / 3 (0%)
🔄 Phases In Progress: 0
⏳ Phases Pending: 3
📁 Files Created: 0
📝 Files Modified: 0

PLANNING STATUS
⏳ Phase 1 Documentation Pending
⏳ Phase 2 Documentation Pending
⏳ Phase 3 Documentation Pending

══════════════════════════════════════════════════════════════

PHASE BREAKDOWN

Phase 1: Form Configuration FK Default Resolution
Status: ⏳ Pending
Category: backend
Tasks: 0/4 completed (0%)

Phase 2: Workflow Integration for Status Transitions
Status: ⏳ Pending
Category: backend
Tasks: 0/3 completed (0%)

Phase 3: Alert System Enhancement
Status: ⏳ Pending
Category: fullstack
Tasks: 0/4 completed (0%)

══════════════════════════════════════════════════════════════

CURRENT FOCUS
Working on: Planning phase documentation
Next Task: Create Phase 1 documentation using /default-statuses-build-phase

Key Decisions:
  - Use form config FK resolution (not template variables) for default status values
  - Names resolved to UUIDs at formdata processing time

══════════════════════════════════════════════════════════════

BLOCKERS
None currently

══════════════════════════════════════════════════════════════

MILESTONES
⏳ Planning Complete
⏳ Phase 1 Complete - FK Default Resolution
⏳ Phase 2 Complete - Workflow Integration
⏳ Phase 3 Complete - Alert System
⏳ Project Complete

══════════════════════════════════════════════════════════════

NEXT STEPS
1. Run /default-statuses-build-phase to create Phase 1 documentation
2. Run /default-statuses-next to begin implementation
```

## Tips

- Be concise but comprehensive
- Use visual hierarchy (boxes, spacing, emojis)
- Highlight actionable items
- Show both high-level overview and detailed breakdown
- If blockers exist, make them very visible
- Include helpful next steps at the end
- Note phase categories for context

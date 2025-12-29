# Default Status Management Dependencies Command

Display dependency information for the Default Status Management implementation, showing internal phase dependencies and external plan dependencies.

## Your Task

### 1. Read PROGRESS.yaml

Read `.claude/plans/DEFAULT_STATUSES_PLAN/PROGRESS.yaml` to gather:
- Phase internal dependencies
- External plan dependencies
- Current phase statuses
- Blocker information

### 2. Analyze Dependencies

For this plan, the dependency chain is:

```
Phase 1: Form Configuration FK Default Resolution
    └─► Phase 2: Workflow Integration for Status Transitions
        └─► Phase 3: Alert System Enhancement
```

- Phase 1 has no dependencies (starting phase)
- Phase 2 depends on Phase 1 (needs FK resolution working)
- Phase 3 depends on Phase 2 (needs workflow integration working)

### 3. Generate Dependency Report

```
🔗 Default Status Management - Dependency Report

══════════════════════════════════════════════════════════════

OVERVIEW

Total Phases: 3
Phases with Dependencies: 2
External Plans Referenced: 0

══════════════════════════════════════════════════════════════

INTERNAL DEPENDENCIES (Phase → Phase)

Phase 1: Form Configuration FK Default Resolution
  Status: [STATUS]
  Depends on: None (starting phase)
  Blocks: Phase 2

Phase 2: Workflow Integration for Status Transitions
  Status: [STATUS]
  Depends on: Phase 1 [STATUS_INDICATOR]
  Blocks: Phase 3

Phase 3: Alert System Enhancement
  Status: [STATUS]
  Depends on: Phase 2 [STATUS_INDICATOR]
  Blocks: None

Status Indicators:
  ✅ = Completed (dependency satisfied)
  🔄 = In Progress (dependency not yet satisfied)
  ⏳ = Pending (dependency not yet satisfied)
  🚫 = Blocked (dependency blocked)

══════════════════════════════════════════════════════════════

EXTERNAL DEPENDENCIES (Other Plans)

No external plan dependencies.

This plan is standalone and can be executed independently.

══════════════════════════════════════════════════════════════

DEPENDENCY GRAPH (ASCII)

┌─────────────────────────────────────────┐
│ Phase 1: Form Config FK Default Res.    │
│ Status: [STATUS]                         │
└────────────────────┬────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────┐
│ Phase 2: Workflow Integration           │
│ Status: [STATUS]                         │
└────────────────────┬────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────┐
│ Phase 3: Alert System Enhancement       │
│ Status: [STATUS]                         │
└─────────────────────────────────────────┘

══════════════════════════════════════════════════════════════

CRITICAL PATH ANALYSIS

Longest dependency chain: 3 phases (sequential)

All phases must be completed in order:
1. Phase 1 → 2. Phase 2 → 3. Phase 3

No parallelizable work in this plan.

══════════════════════════════════════════════════════════════

WARNINGS & BLOCKERS

[If any issues exist, list them here]

✅ No dependency issues detected. All dependencies will be satisfied in order.

══════════════════════════════════════════════════════════════

RECOMMENDATIONS

[Based on current status:]

If all dependencies OK:
✅ Current and next phases have all dependencies satisfied.
   Safe to proceed with /default-statuses-next

If Phase 1 not complete and trying Phase 2:
⚠️  Phase 2 depends on Phase 1 (Form Config FK Default Resolution)
   Recommendation: Complete Phase 1 first

══════════════════════════════════════════════════════════════

NEXT STEPS

1. Continue with current plan: /default-statuses-next
2. Monitor status: /default-statuses-status
```

## Dependency Details

### Why Phase 2 Depends on Phase 1

Phase 2 (Workflow Integration) sets up automation rules that update status fields. For these updates to work correctly:
- The FK default resolution from Phase 1 must be in place
- Orders must already be getting default "Pending" status
- The workflow can then transition from "Pending" to "Allocated"

### Why Phase 3 Depends on Phase 2

Phase 3 (Alert System) extends the alert action used in Phase 2's automation rules:
- Phase 2 seeds the "Allocation Failed - Alert Ops" rule
- Phase 3 enhances the alert action to support role-based recipients
- Without Phase 2's rules in place, there's no context for the alert enhancements

## Tips

- Always complete phases in order for this plan
- Use /default-statuses-status to check current phase status
- If you need to skip ahead, use /default-statuses-phase N with caution

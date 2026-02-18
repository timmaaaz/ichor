# Workflow Action Gap Remediation Dependencies Command

Display dependency information for the Workflow Action Gap Remediation implementation, showing internal phase dependencies and external plan dependencies.

## Your Task

### 1. Read PROGRESS.yaml

Read `.claude/plans/WORKFLOW_GAPS_PLAN/PROGRESS.yaml` to gather:
- Phase internal dependencies
- External plan dependencies
- Current phase statuses
- Blocker information

### 2. Analyze Dependencies

For each phase, check:
- `dependencies.internal` - which other phases in this plan it depends on
- `dependencies.external` - which other plans it depends on
- Current status of dependencies

### 3. Generate Dependency Report

Create a visual representation of dependencies:

```
🔗 Workflow Action Gap Remediation - Dependency Report

══════════════════════════════════════════════════════════════

OVERVIEW

Total Phases: 9
Phases with Dependencies: [COUNT]
External Plans Referenced: None

══════════════════════════════════════════════════════════════

INTERNAL DEPENDENCIES (Phase → Phase)

Phase 1: Add Missing Tables to Whitelist
  Status: [STATUS]
  Depends on: None (starting phase - independent)
  Blocks: Phase 6 (needs procurement tables first, recommended)

Phase 2: Fix FieldChanges Propagation
  Status: [STATUS]
  Depends on: None (independent bug fix)
  Blocks: None

Phase 3: Implement send_notification
  Status: [STATUS]
  Depends on: None (independent stub completion)
  Blocks: None

Phase 4: Implement send_email
  Status: [STATUS]
  Depends on: None (independent stub completion)
  Blocks: None

Phase 5: Implement seek_approval
  Status: [STATUS]
  Depends on: None (but Phase 1 recommended - approvals may need status updates)
  Blocks: None

Phase 6: Add create_purchase_order
  Status: [STATUS]
  Depends on: Phase 1 (recommended - needs procurement tables in whitelist)
  Blocks: None

Phase 7: Add receive_inventory
  Status: [STATUS]
  Depends on: None (independent - uses existing inventory buses)
  Blocks: None

Phase 8: Add call_webhook
  Status: [STATUS]
  Depends on: None (independent - no domain dependencies)
  Blocks: None

Phase 9: Add Template Arithmetic
  Status: [STATUS]
  Depends on: None (independent - template processor only)
  Blocks: None

Status Indicators:
  ✅ = Completed (dependency satisfied)
  🔄 = In Progress (dependency not yet satisfied)
  ⏳ = Pending (dependency not yet satisfied)
  🚫 = Blocked (dependency blocked)

══════════════════════════════════════════════════════════════

EXTERNAL DEPENDENCIES (Other Plans)

No external plan dependencies. All 9 phases are self-contained Go backend work.

══════════════════════════════════════════════════════════════

DEPENDENCY GRAPH (ASCII)

All phases are largely independent. Phase 1 is recommended before Phase 6:

Phase 1 (Tables) ──────────────────────→ Phase 6 (Create PO)
Phase 2 (FieldChanges)                   [Standalone]
Phase 3 (send_notification)              [Standalone]
Phase 4 (send_email)                     [Standalone]
Phase 5 (seek_approval)                  [Standalone]
Phase 7 (receive_inventory)              [Standalone]
Phase 8 (call_webhook)                   [Standalone]
Phase 9 (Template Arithmetic)            [Standalone]

══════════════════════════════════════════════════════════════

CRITICAL PATH ANALYSIS

Longest dependency chain: Phase 1 → Phase 6 (2 phases)

This means most phases can be done in any order.

Parallelizable work:
- Phases 2, 3, 4, 5, 7, 8, 9 are all independent — can be done in any order
- Phase 1 should precede Phase 6 if possible

══════════════════════════════════════════════════════════════

WARNINGS & BLOCKERS

[IF ANY ISSUES]
[FOR EACH ISSUE]
⚠️  [ISSUE_TYPE]: [ISSUE_DESCRIPTION]
    Affected: [AFFECTED_PHASES]
    Resolution: [SUGGESTED_RESOLUTION]
[/FOR]
[ELSE]
✅ No dependency issues detected. All dependencies are satisfied or will be satisfied in order.
[/IF]

══════════════════════════════════════════════════════════════

RECOMMENDATIONS

[IF CURRENT_PHASE_HAS_UNMET_DEPENDENCIES]
🚨 Current phase has unmet dependencies.
   Recommendation: Complete dependency phases before proceeding.

[ELSE]
✅ Current and next phases have all dependencies satisfied.
   Safe to proceed with /workflow-gaps-next

[/IF]

══════════════════════════════════════════════════════════════

NEXT STEPS

1. Continue with current plan: /workflow-gaps-next
2. Monitor status: /workflow-gaps-status

Suggested phase order for maximum efficiency:
  Phase 1 (10 min) → Phase 2 (30 min) → Phase 3 (1h) → Phase 4 (2h)
  → Phase 6 (4h) → Phase 7 (3h) → Phase 8 (2h) → Phase 9 (1h)
  → Phase 5 (8h, most complex)
```

## Tips for Dependency Management

1. **Minimize Dependencies**: This plan intentionally has few hard dependencies
2. **Phase 1 First**: Adding tables to the whitelist unblocks Phase 6
3. **Phase 5 Last**: seek_approval is the most complex — do easier phases first to build confidence
4. **Independent Phases**: Phases 2, 3, 4, 7, 8, 9 can be done in any order

## Example Usage

```bash
# Check all dependencies
/workflow-gaps-dependencies

# Review before jumping to a phase
/workflow-gaps-dependencies
/workflow-gaps-phase 5  # Only if dependencies satisfied
```

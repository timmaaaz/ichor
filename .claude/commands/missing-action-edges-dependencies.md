# Universal Action Edge Enforcement Dependencies Command

Display dependency information for the Universal Action Edge Enforcement implementation, showing internal phase dependencies and external plan dependencies.

## Your Task

### 1. Read PROGRESS.yaml

Read `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml` to gather:
- Phase internal dependencies
- External plan dependencies
- Current phase statuses
- Blocker information

### 2. Analyze Dependencies

For each phase, check:
- `dependencies.internal` - which other phases in this plan it depends on
- `dependencies.external` - which other plans it depends on
- Current status of dependencies

### 3. Check External Plan Status

For each external dependency:
1. Check if the plan exists in `.claude/plans/`
2. Read that plan's PROGRESS.yaml
3. Determine if the external plan is completed
4. Note any blockers in external plans

### 4. Generate Dependency Report

Create a visual representation of dependencies:

```
🔗 Universal Action Edge Enforcement - Dependency Report

══════════════════════════════════════════════════════════════

OVERVIEW

Total Phases: 6
Phases with Dependencies: 2 (Phase 4, Phase 5)
External Plans Referenced: 0

══════════════════════════════════════════════════════════════

INTERNAL DEPENDENCIES (Phase → Phase)

Phase 1: Validation Layer Changes
  Status: {{STATUS}}
  Depends on: None (starting phase)
  Blocks: None (can proceed independently)

Phase 2: Remove execution_order Field
  Status: {{STATUS}}
  Depends on: None (can run after Phase 1)
  Blocks: Phase 4

Phase 3: Remove Linear Executor
  Status: {{STATUS}}
  Depends on: None (can run after Phase 1)
  Blocks: Phase 4

Phase 4: Test Updates
  Status: {{STATUS}}
  Depends on: Phase 2 {{STATUS_INDICATOR}}, Phase 3 {{STATUS_INDICATOR}}
  Blocks: None

Phase 5: Seed Data Updates
  Status: {{STATUS}}
  Depends on: None (can run in parallel with Phase 4)
  Blocks: None

Phase 6: Documentation Updates
  Status: {{STATUS}}
  Depends on: None (best to run last)
  Blocks: None

Status Indicators:
  ✅ = Completed (dependency satisfied)
  🔄 = In Progress (dependency not yet satisfied)
  ⏳ = Pending (dependency not yet satisfied)
  🚫 = Blocked (dependency blocked)

══════════════════════════════════════════════════════════════

EXTERNAL DEPENDENCIES (Other Plans)

No external plan dependencies.

══════════════════════════════════════════════════════════════

DEPENDENCY GRAPH (ASCII)

Phase 1 (Validation)
    │
    ├─→ Phase 2 (Remove execution_order) ─┐
    │                                      │
    └─→ Phase 3 (Remove Linear Executor) ─┼─→ Phase 4 (Tests)
                                          │
    Phase 5 (Seeds) ──────────────────────┘

    Phase 6 (Documentation) ← Best to run last

══════════════════════════════════════════════════════════════

CRITICAL PATH ANALYSIS

Longest dependency chain:
Phase 1 → Phase 2/3 → Phase 4 → Phase 6

This is the minimum number of sequential phases that must be completed.

Parallelizable work:
- Phase 2 and Phase 3 can run in parallel (after Phase 1)
- Phase 4 and Phase 5 can run in parallel (after Phase 2 & 3)

══════════════════════════════════════════════════════════════

WARNINGS & BLOCKERS

{{IF ANY ISSUES}}

{{FOR EACH ISSUE}}
⚠️  {{ISSUE_TYPE}}: {{ISSUE_DESCRIPTION}}
    Affected: {{AFFECTED_PHASES}}
    Resolution: {{SUGGESTED_RESOLUTION}}
{{/FOR}}

{{ELSE}}
✅ No dependency issues detected. All dependencies are satisfied or will be satisfied in order.
{{/IF}}

══════════════════════════════════════════════════════════════

RECOMMENDATIONS

{{IF CURRENT_PHASE_HAS_UNMET_DEPENDENCIES}}
🚨 Current phase ({{CURRENT_PHASE}}) has unmet dependencies:
   - {{UNMET_DEP_1}}
   - {{UNMET_DEP_2}}

   Recommendation: Complete dependency phases before proceeding.
   Alternative: Use /missing-action-edges-phase to jump to a phase with satisfied dependencies.

{{ELSE IF NEXT_PHASE_HAS_UNMET_DEPENDENCIES}}
⚠️  Next phase ({{NEXT_PHASE}}) has unmet dependencies:
   - {{UNMET_DEP_1}}

   Recommendation: Consider reordering phases or completing dependencies first.

{{ELSE}}
✅ Current and next phases have all dependencies satisfied.
   Safe to proceed with /missing-action-edges-next

{{/IF}}

══════════════════════════════════════════════════════════════

NEXT STEPS

{{IF DEPENDENCIES_OK}}
1. Continue with current plan: /missing-action-edges-next
2. Monitor status: /missing-action-edges-status

{{ELSE}}
1. Review dependency warnings above
2. Complete blocking dependencies first
3. Or adjust phase order in PROGRESS.yaml
4. Re-run this command to verify: /missing-action-edges-dependencies

{{/IF}}
```

## Dependency Analysis Features

### Internal Dependency Tracking

For each phase, identify:
- Which phases it depends on (prerequisites)
- Which phases depend on it (blocks)
- Whether dependencies are satisfied (all prerequisite phases completed)

### External Dependency Tracking

For external plans:
- Verify plan exists
- Check completion status
- Identify which phases need the external plan
- Warn if external plan is incomplete or blocked

### Circular Dependency Detection

Check for circular dependencies:
```
⚠️  CIRCULAR DEPENDENCY DETECTED!

Phase A depends on Phase B
Phase B depends on Phase C
Phase C depends on Phase A

This creates an impossible-to-resolve cycle.
Please fix PROGRESS.yaml to remove circular dependencies.
```

### Critical Path Calculation

Identify the longest chain of sequential dependencies:
- This represents the minimum time to complete the project
- Helps identify bottlenecks
- Shows which phases can run in parallel

## Tips for Dependency Management

### Best Practices

1. **Minimize Dependencies**: Fewer dependencies = more flexibility
2. **Sequential When Necessary**: Some dependencies are unavoidable (e.g., remove code before updating tests)
3. **Parallel When Possible**: Independent phases can be worked on simultaneously
4. **External First**: Complete external plan dependencies early
5. **Document Why**: Explain why dependencies exist in phase documentation

### Common Dependency Patterns for This Plan

**Sequential Build** (Validation → Removal → Tests):
```
Phase 1 (validation) → Phase 2/3 (removal) → Phase 4 (tests)
```

**Parallel Development**:
```
Phase 2 (remove field) ┬→ Phase 4 (tests)
Phase 3 (remove exec) ─┘
```

## Example Usage

```bash
# Check all dependencies
/missing-action-edges-dependencies

# Review before jumping to a phase
/missing-action-edges-dependencies
/missing-action-edges-phase 5  # Only if dependencies satisfied
```

## When to Check Dependencies

**Check dependencies when**:
- Starting the project
- Before jumping to a specific phase
- When encountering blockers
- When planning parallel work
- When adding new phases

**Especially important if**:
- Working on phases out of order
- Multiple developers working on same plan
- Complex inter-phase dependencies
- External plan dependencies exist

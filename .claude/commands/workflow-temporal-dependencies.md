# Workflow Temporal Implementation Dependencies Command

Display dependency information for the Workflow Temporal Implementation, showing internal phase dependencies and external plan dependencies.

## Your Task

### 1. Read PROGRESS.yaml

Read `.claude/plans/WORKFLOW_TEMPORAL_PLAN/PROGRESS.yaml` to gather:
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
🔗 Workflow Temporal Implementation - Dependency Report

══════════════════════════════════════════════════════════════

OVERVIEW

Total Phases: 13
Phases with Dependencies: 11
External Plans Referenced: 0

══════════════════════════════════════════════════════════════

INTERNAL DEPENDENCIES (Phase → Phase)

Phase 1: Infrastructure Setup
  Status: {{STATUS}}
  Depends on: None (starting phase)
  Blocks: Phases 3-13

Phase 2: Temporalgraph Evaluation
  Status: {{STATUS}}
  Depends on: None (research phase)
  Blocks: Phase 3 (decision impacts implementation)

Phase 3: Core Models & Context
  Status: {{STATUS}}
  Depends on: Phase 2 {{STATUS_INDICATOR}} (for library vs custom decision)
  Blocks: Phases 4-9

Phase 4: Graph Executor
  Status: {{STATUS}}
  Depends on: Phase 3 {{STATUS_INDICATOR}}
  Blocks: Phases 5-9, 10

Phase 5: Workflow Implementation
  Status: {{STATUS}}
  Depends on: Phase 4 {{STATUS_INDICATOR}}
  Blocks: Phases 6-9

Phase 6: Activities & Async
  Status: {{STATUS}}
  Depends on: Phase 5 {{STATUS_INDICATOR}}
  Blocks: Phases 7-9

Phase 7: Trigger System
  Status: {{STATUS}}
  Depends on: Phase 6 {{STATUS_INDICATOR}}
  Blocks: Phase 9

Phase 8: Edge Store Adapter
  Status: {{STATUS}}
  Depends on: Phase 6 {{STATUS_INDICATOR}}
  Blocks: Phase 9

Phase 9: Worker Service & Wiring
  Status: {{STATUS}}
  Depends on: Phases 7, 8 {{STATUS_INDICATOR}}
  Blocks: Phases 10-13

Phase 10: Graph Executor Unit Tests
  Status: {{STATUS}}
  Depends on: Phase 4 {{STATUS_INDICATOR}}
  Blocks: Phase 13

Phase 11: Workflow Integration Tests
  Status: {{STATUS}}
  Depends on: Phase 9 {{STATUS_INDICATOR}}
  Blocks: Phase 13

Phase 12: Edge Case & Limit Tests
  Status: {{STATUS}}
  Depends on: Phase 9 {{STATUS_INDICATOR}}
  Blocks: Phase 13

Phase 13: Kubernetes Deployment
  Status: {{STATUS}}
  Depends on: Phases 1, 10, 11, 12 {{STATUS_INDICATOR}}
  Blocks: None (final phase)

Status Indicators:
  ✅ = Completed (dependency satisfied)
  🔄 = In Progress (dependency not yet satisfied)
  ⏳ = Pending (dependency not yet satisfied)
  🚫 = Blocked (dependency blocked)

══════════════════════════════════════════════════════════════

EXTERNAL DEPENDENCIES (Other Plans)

No external plan dependencies.

This plan is self-contained and does not depend on other plans.

══════════════════════════════════════════════════════════════

DEPENDENCY GRAPH (ASCII)

┌────────────────┐     ┌──────────────────────┐
│    Phase 1     │     │       Phase 2        │
│ Infrastructure │     │ Temporalgraph Eval   │
└───────┬────────┘     └──────────┬───────────┘
        │                         │
        │              ┌──────────┘
        │              │
        │      ┌───────▼────────┐
        │      │     Phase 3    │
        │      │  Core Models   │
        │      └───────┬────────┘
        │              │
        │      ┌───────▼────────┐
        │      │     Phase 4    │
        │      │ Graph Executor │──────────────────────┐
        │      └───────┬────────┘                      │
        │              │                               │
        │      ┌───────▼────────┐               ┌──────▼───────┐
        │      │     Phase 5    │               │   Phase 10   │
        │      │    Workflow    │               │  Unit Tests  │
        │      └───────┬────────┘               └──────┬───────┘
        │              │                               │
        │      ┌───────▼────────┐                      │
        │      │     Phase 6    │                      │
        │      │   Activities   │                      │
        │      └───────┬────────┘                      │
        │              │                               │
        │      ┌───────┴────────┐                      │
        │      │                │                      │
        │ ┌────▼───┐     ┌──────▼────┐                 │
        │ │Phase 7 │     │  Phase 8  │                 │
        │ │Trigger │     │Edge Store │                 │
        │ └────┬───┘     └──────┬────┘                 │
        │      │                │                      │
        │      └────────┬───────┘                      │
        │               │                              │
        │       ┌───────▼────────┐                     │
        │       │     Phase 9    │                     │
        │       │ Worker Wiring  │                     │
        │       └───────┬────────┘                     │
        │               │                              │
        │       ┌───────┴───────┐                      │
        │       │               │                      │
        │ ┌─────▼────┐   ┌──────▼─────┐                │
        │ │ Phase 11 │   │  Phase 12  │                │
        │ │Int Tests │   │ Edge Cases │                │
        │ └─────┬────┘   └──────┬─────┘                │
        │       │               │                      │
        │       └───────┬───────┘                      │
        │               │                              │
        └───────────────┼──────────────────────────────┘
                        │
                ┌───────▼────────┐
                │    Phase 13    │
                │ K8s Deployment │
                └────────────────┘

══════════════════════════════════════════════════════════════

CRITICAL PATH ANALYSIS

Longest dependency chain:
Phase 2 → Phase 3 → Phase 4 → Phase 5 → Phase 6 → Phase 7/8 → Phase 9 → Phase 11/12 → Phase 13

This is the minimum sequential path (10 phases).

Parallelizable work:
- Phase 1 (Infrastructure) can run parallel to Phase 2 (Evaluation)
- Phase 7 (Trigger) and Phase 8 (Edge Store) can run in parallel
- Phase 10, 11, 12 (Testing phases) can partially run in parallel

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
   Alternative: Use /workflow-temporal-phase to jump to a phase with satisfied dependencies.

{{ELSE IF NEXT_PHASE_HAS_UNMET_DEPENDENCIES}}
⚠️  Next phase ({{NEXT_PHASE}}) has unmet dependencies:
   - {{UNMET_DEP_1}}

   Recommendation: Consider reordering phases or completing dependencies first.

{{ELSE}}
✅ Current and next phases have all dependencies satisfied.
   Safe to proceed with /workflow-temporal-next

{{/IF}}

══════════════════════════════════════════════════════════════

NEXT STEPS

{{IF DEPENDENCIES_OK}}
1. Continue with current plan: /workflow-temporal-next
2. Monitor status: /workflow-temporal-status

{{ELSE}}
1. Review dependency warnings above
2. Complete blocking dependencies first
3. Or adjust phase order in PROGRESS.yaml
4. Re-run this command to verify: /workflow-temporal-dependencies

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
2. **Sequential When Necessary**: Some dependencies are unavoidable (e.g., models before executor)
3. **Parallel When Possible**: Independent phases can be worked on simultaneously
4. **Document Why**: Explain why dependencies exist in phase documentation

### Common Dependency Patterns

**Sequential Build** (Models → Executor → Workflow):
```
Phase 3 (models) → Phase 4 (executor) → Phase 5 (workflow)
```

**Parallel Development**:
```
Phase 7 (trigger)   ┬→ Phase 9 (wiring)
Phase 8 (edge store)┘
```

**Converging Work**:
```
Phase 10 (unit tests)    ┐
Phase 11 (integration)   ├→ Phase 13 (deployment)
Phase 12 (edge cases)    ┘
```

## Example Usage

```bash
# Check all dependencies
/workflow-temporal-dependencies

# Review before jumping to a phase
/workflow-temporal-dependencies
/workflow-temporal-phase 5  # Only if dependencies satisfied
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

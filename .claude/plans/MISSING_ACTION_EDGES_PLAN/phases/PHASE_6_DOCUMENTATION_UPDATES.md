# Phase 6: Documentation Updates

**Category**: Documentation
**Status**: Pending
**Dependencies**: Phase 1-5 (All implementation and test phases)

---

## Overview

Remove all references to linear execution mode, `execution_order`, and backwards compatibility from the workflow documentation. Replace with documentation stating that edges are required for all workflow rules with actions.

After Phases 1-5, the codebase has a single execution mode (graph-based BFS traversal via edges). The documentation must reflect this reality so developers don't rely on behaviors that no longer exist.

### Goals

1. Remove all mentions of `execution_order` field from docs (column definitions, examples, API reference)
2. Remove all "backwards compatibility" and "linear fallback" sections
3. Replace dual-mode language with single-mode "edges required" language
4. Update all code examples and JSON payloads to omit `execution_order`

### Why This Phase Matters

Stale documentation is worse than no documentation. If a developer reads that rules "fall back to linear execution when no edges exist" and relies on that behavior, their workflow will silently fail. This phase ensures the documentation matches the code.

---

## Prerequisites

Before starting this phase, ensure:

- [ ] Phases 1-5 are completed and all tests pass
- [ ] `make test` passes
- [ ] All documentation files exist at the expected paths

---

## Documentation Structure

All workflow documentation lives in `docs/workflow/`:

```
docs/workflow/
├── README.md                    ← Task 1: Execution Modes section
├── branching.md                 ← Task 2: Backwards compatibility
├── database-schema.md           ← Task 3: execution_order column/index
├── architecture.md              ← Task 4: Parallel execution section
├── configuration/
│   └── rules.md                 ← Task 5: Execution order examples
├── actions/
│   └── overview.md              ← Task 6: Linear execution section
└── api-reference.md             ← Task 7: API fallback notes
```

---

## Task Breakdown

### Task 1: Update Workflow README

**Files**:
- `docs/workflow/README.md` - Lines 78-100: "Execution Modes" section

**Current content** (to REPLACE):
```markdown
### 4. Execution Modes

Rules support two execution modes:

| Mode | Description |
|------|-------------|
| **Linear** | Actions execute based on `execution_order`. Default mode. |
| **Graph** | Actions execute based on `action_edges`. Enables branching with `evaluate_condition`. |

The executor automatically detects which mode to use:
- If the rule has edges in `workflow.action_edges`, graph mode is used
- If no edges exist, linear mode is used (backwards compatible)

**Graph mode** enables conditional workflows using the `evaluate_condition` action:
- The condition action sets `BranchTaken` to `"true_branch"` or `"false_branch"`
- Edges with matching types are followed; others are skipped
```

**Replace with**:
```markdown
### 4. Action Execution

Rules with actions execute using graph-based BFS traversal via `action_edges`. Edges define the execution order and flow:

| Edge Type | Description |
|-----------|-------------|
| **start** | Entry point — connects to the first action (no source) |
| **sequence** | Unconditional flow from one action to the next |
| **true_branch** | Conditional path taken when `evaluate_condition` returns true |
| **false_branch** | Conditional path taken when `evaluate_condition` returns false |
| **always** | Always-execute path from a condition (runs regardless of branch) |

**All rules with actions must have edges.** Rules without actions (trigger-only rules) are valid and saved as inactive drafts.

See [Branching](branching.md) for complete documentation on conditional workflows.
```

**Validation**:
- [ ] No mention of "linear" execution mode
- [ ] No mention of "backwards compatible"
- [ ] Edge types are documented
- [ ] Edge requirement is clearly stated

---

### Task 2: Update Branching Documentation

**Files**:
- `docs/workflow/branching.md` - Lines 7, 14, 94-108

**Change 1** — Line 7 (REPLACE):

Current:
```markdown
By default, workflow rules execute actions in a linear sequence based on `execution_order`. Graph-based execution enables:
```

Replace with:
```markdown
Workflow rules execute actions using graph-based traversal with action edges. This enables:
```

**Change 2** — Line 14 (REMOVE ENTIRELY):

Current:
```markdown
The graph executor is **backwards compatible** - rules without edges automatically fall back to linear `execution_order` execution.
```

Remove this line completely.

**Change 3** — Lines 94-108: "Backwards Compatibility" section (REMOVE ENTIRELY):

Current:
```markdown
### Backwards Compatibility

Rules execute using graph traversal only when edges exist:

```go
// Falls back to linear execution if:
// 1. No edges exist for the rule
// 2. Edges exist but no start edges defined

if len(edges) == 0 {
    return ae.ExecuteRuleActions(ctx, ruleID, executionContext)  // Linear
}
```

This means existing rules continue to work without modification.
```

Remove this entire section including the header.

**Validation**:
- [ ] No mention of "backwards compatible"
- [ ] No mention of "linear" execution
- [ ] No mention of `ExecuteRuleActions` (linear function)
- [ ] No mention of "fallback"

---

### Task 3: Update Database Schema Documentation

**Files**:
- `docs/workflow/database-schema.md` - Lines 117, 170-171, 316

**Change 1** — Line 117 (REMOVE row):

Current (in `workflow.rule_actions` table):
```markdown
| `execution_order` | INT | NO | - | Execution order (≥1) |
```

Remove this row entirely from the table.

**Change 2** — Lines 170-171 (REPLACE):

Current:
```markdown
**Usage notes:**
- Rules WITHOUT edges fall back to linear `execution_order` execution (backwards compatible)
- Rules WITH edges use graph-based BFS traversal
```

Replace with:
```markdown
**Usage notes:**
- All rules with actions require edges for execution
- Execution uses graph-based BFS traversal via `action_edges`
```

**Change 3** — Line 316 (REMOVE):

Current:
```sql
CREATE INDEX idx_rule_actions_order ON workflow.rule_actions(execution_order);
```

Remove this index reference entirely.

**Validation**:
- [ ] No `execution_order` column in table definition
- [ ] No `execution_order` index reference
- [ ] No "backwards compatible" or "fall back" language

---

### Task 4: Update Architecture Documentation

**Files**:
- `docs/workflow/architecture.md` - Lines 470-512

**Change 1** — Lines 470-476: "Parallel Execution" section (REMOVE ENTIRELY):

Current:
```markdown
### Parallel Execution

Actions with the same `execution_order` run in parallel:
```
execution_order=1: [email, alert]  ← run in parallel
execution_order=2: [update_field]  ← waits for order 1
execution_order=3: [approval]      ← waits for order 2
```
```

Remove this entire section.

**Change 2** — Lines 478-512: Update "Graph-Based Execution" section:

Find and replace the execution model description. Change:
```markdown
**Execution model:**
- Rules WITHOUT edges: Linear execution by `execution_order` (backwards compatible)
- Rules WITH edges: Graph traversal using directed edges
```

To:
```markdown
**Execution model:**
- All rules with actions use graph traversal via directed edges
- Rules without actions are saved as inactive drafts
```

Remove any remaining mentions of "backwards compatible" or linear fallback in this section.

**Validation**:
- [ ] "Parallel Execution" section removed
- [ ] No "backwards compatible" language
- [ ] No `execution_order` references
- [ ] Graph-based execution documented as the only mode

---

### Task 5: Update Rules Configuration Documentation

**Files**:
- `docs/workflow/configuration/rules.md` - Lines 112, 119-128, 154, 421-461

**Change 1** — Line 112 (REMOVE row):

Current (in RuleAction table):
```markdown
| `execution_order` | int | Yes | Order of execution (>= 1) |
```

Remove this row.

**Change 2** — Lines 119-128: "Execution Order" subsection (REMOVE ENTIRELY):

Current:
```markdown
### Execution Order

Actions with the **same** `execution_order` run in **parallel**.
Actions with **different** orders run **sequentially**.

```
Order 1: [send_email, create_alert]  ← parallel
Order 2: [update_field]               ← waits for 1
Order 3: [seek_approval]              ← waits for 2
```
```

Remove this entire subsection.

**Change 3** — Line 154: Remove `execution_order` from JSON example:

Remove this line from the example:
```json
  "execution_order": 1,
```

**Change 4** — Lines 421-430: "Backwards Compatibility" section (REMOVE ENTIRELY):

Current:
```markdown
### Backwards Compatibility

The graph executor is **fully backwards compatible**:

1. If a rule has **no edges**, the executor falls back to `execution_order`-based linear execution
2. Existing rules without edges continue to work exactly as before
3. You can migrate rules incrementally by adding edges

**Source**: `business/sdk/workflow/executor.go:222-227`
```

Remove this entire section.

**Change 5** — Lines 443-461: "Edges vs Execution Order" section (REPLACE):

Current section discusses choosing between linear and graph modes. Replace with:

```markdown
### Edge Design Best Practices

All rules with actions require edges. When designing workflows:

- **Always define a `start` edge** — this is the entry point (no source action)
- **Use `sequence` edges** for unconditional linear flows (Action A → Action B)
- **Use `true_branch` / `false_branch` edges** after `evaluate_condition` actions
- **Use `always` edges** for actions that should run regardless of condition outcome
- **Set `edge_order`** to control execution priority when multiple edges share the same source
- **Draft workflows** — rules without actions are saved as inactive and can have edges added later
```

**Change 6** — Lines 462-467: "Execution Order (Linear Mode)" section (REMOVE ENTIRELY):

Current:
```markdown
### Execution Order (Linear Mode)

1. Put independent actions at the same order (parallel)
2. Put dependent actions at higher orders (sequential)
3. Put validation/approval actions last
```

Remove this entire section.

**Validation**:
- [ ] No `execution_order` in table definition or examples
- [ ] "Execution Order" subsection removed
- [ ] "Backwards Compatibility" section removed
- [ ] "Edges vs Execution Order" replaced with "Edge Design Best Practices"
- [ ] "Execution Order (Linear Mode)" section removed
- [ ] JSON examples don't include `execution_order`

---

### Task 6: Update Actions Overview Documentation

**Files**:
- `docs/workflow/actions/overview.md` - Lines 178-209

**Change 1** — Lines 178-191: Remove "Linear Execution" subsection:

Current:
```markdown
## Execution Order

### Linear Execution (Default)

Actions execute based on their `execution_order` in the rule:

```
Order 1: [action_a, action_b]  ← Parallel (same order)
Order 2: [action_c]            ← Sequential (waits for order 1)
Order 3: [action_d, action_e]  ← Parallel (waits for order 2)
```

Actions with the same execution order run concurrently.
```

Remove the "Linear Execution (Default)" subsection entirely.

**Change 2** — Lines 193-209: Update "Graph-Based Execution" subsection:

Current:
```markdown
### Graph-Based Execution (Branching)

When `action_edges` are defined for a rule, the executor uses BFS graph traversal instead of `execution_order`:
...
Rules WITHOUT edges automatically fall back to linear `execution_order` execution (backwards compatible).
```

Replace with a single "Execution Order" section:
```markdown
## Execution Order

Actions execute using BFS graph traversal based on `action_edges`. All rules with actions must have edges defining the execution flow.

[Keep the existing graph traversal explanation, edge type descriptions, etc.]

Rules without actions are saved as inactive drafts and do not execute.
```

Remove the line about "fall back to linear execution" (around line 207).

**Validation**:
- [ ] "Linear Execution" subsection removed
- [ ] "Graph-Based Execution" renamed/simplified to "Execution Order"
- [ ] No mention of `execution_order` field
- [ ] No "fall back" or "backwards compatible" language
- [ ] Graph BFS traversal documented as the only mode

---

### Task 7: Update API Reference

**Files**:
- `docs/workflow/api-reference.md` - Line 625

**Current content** (in Edge API section):
```markdown
**Notes:**
- After deleting all edges, the rule falls back to linear `execution_order` execution
- Useful for resetting a rule's action graph
```

**Replace with**:
```markdown
**Notes:**
- After deleting all edges, the rule will not execute (edges are required for rules with actions)
- Useful for resetting a rule's action graph before rebuilding
```

Also search the file for any other `execution_order` mentions in request/response examples and remove them.

**Validation**:
- [ ] No "falls back to linear" language
- [ ] No `execution_order` in request/response examples
- [ ] Edge deletion behavior correctly documented

---

## Deliverables

- [ ] Updated `docs/workflow/README.md` — replaced "Execution Modes" with "Action Execution"
- [ ] Updated `docs/workflow/branching.md` — removed backwards compatibility sections
- [ ] Updated `docs/workflow/database-schema.md` — removed `execution_order` column, index, and fallback language
- [ ] Updated `docs/workflow/architecture.md` — removed "Parallel Execution", updated graph-based section
- [ ] Updated `docs/workflow/configuration/rules.md` — removed execution order fields, examples, and compat sections
- [ ] Updated `docs/workflow/actions/overview.md` — removed linear execution, updated to single graph mode
- [ ] Updated `docs/workflow/api-reference.md` — updated edge deletion behavior notes
- [ ] All 7 documentation files updated

---

## Validation Criteria

- [ ] `grep -rn "execution_order" docs/workflow/` returns zero results
- [ ] `grep -rn "linear execution" docs/workflow/` returns zero results (case-insensitive)
- [ ] `grep -rn "backwards compatible" docs/workflow/` returns zero results (case-insensitive)
- [ ] `grep -rn "fall back" docs/workflow/` returns zero results (case-insensitive)
- [ ] `grep -rn "fallback" docs/workflow/` returns zero results in execution context
- [ ] `grep -rn "ExecuteRuleActions" docs/workflow/` returns zero results (only `ExecuteRuleActionsGraph` if referenced)
- [ ] Edge requirement is clearly stated in README.md, branching.md, and configuration/rules.md
- [ ] All JSON examples omit `execution_order`
- [ ] Documentation is internally consistent (no section references a removed section)

---

## Testing Strategy

### What to Test

Documentation changes don't have automated tests, but verify:

- **Completeness**: Every `execution_order` reference is removed
- **Consistency**: No section references another section that was removed
- **Accuracy**: The documented behavior matches the code (edges required, graph-only execution)
- **Readability**: Replacement text reads naturally and doesn't leave orphaned fragments

### How to Verify

```bash
# Search for stale references across all workflow docs
grep -rni "execution_order" docs/workflow/
grep -rni "linear execution" docs/workflow/
grep -rni "backwards.compat" docs/workflow/
grep -rni "fall.back" docs/workflow/
grep -rni "ExecuteRuleActions[^G]" docs/workflow/

# Verify edge requirement is documented
grep -rni "edges.*required" docs/workflow/
grep -rni "must have edges" docs/workflow/

# Count total documentation files changed
git diff --stat docs/workflow/
```

---

## Gotchas and Tips

- **Don't remove too much**: Some sections discuss edges and graph execution in positive terms — keep those. Only remove the linear/fallback/backwards-compat language.
- **Check for orphaned references**: If you remove a "Backwards Compatibility" section, check whether any other section links to it (e.g., "See [Backwards Compatibility](#backwards-compatibility)").
- **Preserve useful content**: The branching documentation has valuable content about edge types, BFS traversal, and condition actions. Only remove the backwards-compat parts.
- **JSON example cleanup**: When removing `execution_order` from JSON examples, don't leave trailing commas or broken formatting.
- **Internal links**: Check for `#execution-order` anchor links that may break when the section is renamed or removed.
- **Source code references**: Some docs reference specific line numbers in source files (e.g., `executor.go:222-227`). These line numbers changed in Phases 2-3. Either update them or remove the specific line references.
- **`execution_order` vs `edge_order`**: Be careful not to accidentally remove `edge_order` references — that field still exists and is valid.
- **Case sensitivity**: Search for both `execution_order` (snake_case in SQL/JSON) and `ExecutionOrder` (PascalCase in Go references within docs).

---

## Reference

- Original plan: `.claude/plans/missing-action-edges.md`
- Progress tracking: `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
- Phase 1 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_1_VALIDATION_LAYER.md`
- Phase 2 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_2_REMOVE_EXECUTION_ORDER.md`
- Phase 3 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_3_REMOVE_LINEAR_EXECUTOR.md`
- Phase 4 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_4_TEST_UPDATES.md`
- Phase 5 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_5_SEED_DATA_UPDATES.md`

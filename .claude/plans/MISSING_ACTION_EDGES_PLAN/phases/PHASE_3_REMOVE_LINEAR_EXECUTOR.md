# Phase 3: Remove Linear Executor

**Category**: Backend
**Status**: Pending
**Dependencies**: Phase 1 (Validation Layer Changes), Phase 2 (Remove execution_order Field)

---

## Overview

Remove the linear execution fallback path from the workflow engine, keeping only the graph-based executor. Currently, `ExecuteRuleActionsGraph()` in `executor.go` falls back to `ExecuteRuleActions()` (linear execution) when no edges are found. With Phase 1 enforcing that all rules with actions must have edges, this fallback is dead code. Additionally, the engine's `executeRule()` method in `engine.go` still calls `ExecuteRuleActions()` directly instead of the graph executor.

This phase makes four changes:
1. **Replace fallback logic with errors**: In `ExecuteRuleActionsGraph()`, change the two fallback sites to return meaningful errors instead of calling the linear executor.
2. **Update the engine**: Change `engine.go` line 422 from calling `ExecuteRuleActions()` to calling `ExecuteRuleActionsGraph()`.
3. **Delete `ExecuteRuleActions()`**: Remove the entire linear executor function (lines 112-209 of `executor.go`).
4. **Remove dead commented-out code**: Delete the commented-out `loadRuleActions` function (lines 592-619).

### Goals

1. Replace backwards-compatibility fallback logic in `ExecuteRuleActionsGraph()` with clear error messages
2. Update `engine.go` to route all execution through the graph executor
3. Delete the `ExecuteRuleActions()` function entirely from `executor.go`
4. Remove dead commented-out code referencing the linear executor
5. Ensure no references to the linear executor remain anywhere in non-test production code

### Why This Phase Matters

Having two execution paths is a maintenance burden and a source of confusion. With Phase 1 requiring edges for all rules with actions, the linear executor can never be reached in normal operation. Keeping dead code risks someone accidentally re-enabling it or being confused by its existence. Removing it simplifies the codebase and makes the graph-based execution model the single source of truth.

---

## Prerequisites

Before starting this phase, ensure:

- [ ] Phase 1 (Validation Layer Changes) is completed
- [ ] Phase 2 (Remove execution_order Field) is completed
- [ ] **Database validation**: All existing rules with actions have edges (run validation query below)
- [ ] Go development environment is ready (`go version` shows 1.23+)
- [ ] You can build the project: `go build ./business/sdk/workflow/...`

### Database Validation Query

Run this query before starting Phase 3. It should return **zero rows**:

```sql
-- Find rules with actions but no edges (these would fail after Phase 3)
SELECT r.id, r.name, COUNT(ra.id) as action_count
FROM workflow.automation_rules r
JOIN workflow.rule_actions ra ON ra.automation_rules_id = r.id
WHERE NOT EXISTS (
    SELECT 1 FROM workflow.action_edges ae WHERE ae.rule_id = r.id
)
GROUP BY r.id, r.name;
```

If this returns results, those rules need edge definitions added before proceeding.

---

> **WARNING: Use `go build`, NOT `go test`, during this phase.**
>
> Test files will have compilation errors after removing `ExecuteRuleActions()`. This is expected and will be fixed in Phase 4.
>
> ```bash
> # CORRECT - verifies production code compiles
> go build ./business/sdk/workflow/...
>
> # WRONG - will fail due to test code references
> go test ./business/sdk/workflow/...
> ```

---

## Recommended Implementation Order

To avoid compilation errors during implementation:

1. **First**: Complete Task 1 (replace fallbacks with errors) — updates callers in `executor.go`
2. **Second**: Complete Task 2 (update engine.go) — updates caller in `engine.go`
3. **Third**: Complete Task 3 (delete `ExecuteRuleActions` function) — safe now that all callers are gone
4. **Fourth**: Complete Task 4 (remove dead commented-out code) — cleanup

**Why this order?** Steps 1-2 update all callers to stop referencing the function. Step 3 removes the function after all references are gone. This ensures `go build` passes at each step.

**Alternative** (single commit): If implementing all at once in a single commit, order doesn't matter since there are no intermediate builds. This is acceptable if you're confident in the changes.

---

## Task Breakdown

### Task 1: Replace Fallback Logic in `ExecuteRuleActionsGraph()` with Errors

**Files**:
- `business/sdk/workflow/executor.go` - Modify two fallback sites and doc comment

**Context**: `ExecuteRuleActionsGraph()` currently has two places where it falls back to the linear executor:

1. **Lines 222-227** - When `len(edges) == 0`:
   ```go
   // Backwards compatibility: if no edges, use linear execution
   if len(edges) == 0 {
       ae.log.Info(ctx, "No edges found for rule, falling back to linear execution",
           "ruleID", ruleID)
       return ae.ExecuteRuleActions(ctx, ruleID, executionContext)
   }
   ```

2. **Lines 264-268** - When `len(startEdges) == 0`:
   ```go
   if len(startEdges) == 0 {
       ae.log.Warn(ctx, "No start edges found for rule, falling back to linear execution",
           "ruleID", ruleID)
       return ae.ExecuteRuleActions(ctx, ruleID, executionContext)
   }
   ```

**Implementation**:

Replace the first fallback (no edges) with an error that returns a meaningful `BatchExecutionResult`:

```go
// All rules with actions must have edges (enforced by validation layer)
if len(edges) == 0 {
    return BatchExecutionResult{
        RuleID:       ruleID,
        TotalActions: 0,
        Status:       "failed",
        ErrorMessage: "rule has no edges - all rules with actions require edges",
        StartedAt:    startTime,
        CompletedAt:  time.Now(),
    }, fmt.Errorf("executeRuleActionsGraph: rule %s has no edges - all rules with actions require edges", ruleID)
}
```

Replace the second fallback (no start edges) with an error that returns a meaningful `BatchExecutionResult`:

```go
if len(startEdges) == 0 {
    return BatchExecutionResult{
        RuleID:       ruleID,
        TotalActions: len(actions),
        Status:       "failed",
        ErrorMessage: "no start edges found - at least one edge with nil source_action_id is required",
        StartedAt:    startTime,
        CompletedAt:  time.Now(),
    }, fmt.Errorf("executeRuleActionsGraph: rule %s has no start edges - at least one edge with nil source_action_id is required", ruleID)
}
```

Also update the function's doc comment to remove the backwards-compatibility note and document error conditions:

```go
// ExecuteRuleActionsGraph executes actions following the edge graph.
// All rules with actions must have edges defining execution flow (enforced by
// the validation layer). Returns an error if the rule has no edges or no start edges.
//
// Start edges are identified by source_action_id = NULL, indicating the beginning
// of the execution graph.
func (ae *ActionExecutor) ExecuteRuleActionsGraph(...) ...
```

**Validation**:
- [ ] No references to `ExecuteRuleActions` remain in `ExecuteRuleActionsGraph`
- [ ] Both fallback paths now return errors with meaningful `BatchExecutionResult` (Status="failed", ErrorMessage populated)
- [ ] Error messages include function context prefix (`executeRuleActionsGraph:`)
- [ ] Error messages are clear and actionable
- [ ] Doc comment updated to document error conditions

---

### Task 2: Update Engine to Use Graph Executor Only

**Files**:
- `business/sdk/workflow/engine.go` - Modify line 422

**Context**: The engine's `executeRule()` method currently calls the linear executor directly:

```go
// Line 422 in engine.go
batchResult, err := e.executor.ExecuteRuleActions(ctx, ruleID, executionContext)
```

**Pre-implementation verification**: Confirm line 422 is the only reference in `engine.go`:

```bash
grep -n "ExecuteRuleActions[^G]" business/sdk/workflow/engine.go
```

This should return exactly one result (line 422).

**Implementation**:

Change to use the graph executor:

```go
// Use graph-based execution (edges define execution flow)
batchResult, err := e.executor.ExecuteRuleActionsGraph(ctx, ruleID, executionContext)
```

**Validation**:
- [ ] `engine.go` calls `ExecuteRuleActionsGraph` instead of `ExecuteRuleActions`
- [ ] `grep -n "ExecuteRuleActions[^G]" business/sdk/workflow/engine.go` returns zero results

---

### Task 3: Delete `ExecuteRuleActions()` Function

**Files**:
- `business/sdk/workflow/executor.go` - Delete lines 112-209

**Context**: `ExecuteRuleActions()` is the original linear executor that iterates through actions in `execution_order`. With Tasks 1 and 2 complete, no production code references this function.

**Implementation**:

Delete the entire function:

```go
// DELETE THIS ENTIRE FUNCTION (lines 112-209):
func (ae *ActionExecutor) ExecuteRuleActions(ctx context.Context, ruleID uuid.UUID, executionContext ActionExecutionContext) (BatchExecutionResult, error) {
    // ... all contents ...
}
```

This function:
- Loads actions via `QueryRoleActionsViewByRuleID`
- Iterates linearly through actions
- Executes each action in sequence based on implicit ordering
- Tracks results, stats, and history

All of this is handled by `ExecuteRuleActionsGraph()` via edge traversal instead.

**Validation**:
- [ ] Function is completely removed
- [ ] `go build ./business/sdk/workflow/...` compiles successfully (production code only)

---

### Task 4: Remove Dead Commented-Out Code

**Files**:
- `business/sdk/workflow/executor.go` - Delete lines 592-619

**Context**: The commented-out `loadRuleActions()` function references `execution_order` in a SQL query and was used by the linear executor. With the linear executor removed, this serves no purpose and should be cleaned up.

**Implementation**:

Delete the entire commented-out block (lines 592-619):

```go
// DELETE THIS COMMENTED-OUT BLOCK:
// loadRuleActions loads actions for a rule from the database
// func (ae *ActionExecutor) loadRuleActions(ctx context.Context, ruleID uuid.UUID) ([]RuleActionView, error) {
//     ...
// }
```

**Validation**:
- [ ] Commented-out code block removed
- [ ] `go build ./business/sdk/workflow/...` still compiles

---

## Post-Implementation Verification

After completing all four tasks, verify there are no remaining references to the deleted function:

```bash
# Search for any remaining references to ExecuteRuleActions (not ExecuteRuleActionsGraph)
grep -rn "ExecuteRuleActions[^G]" business/sdk/workflow/ --include="*.go" | grep -v "_test.go"

# Verify no references in engine.go specifically
grep -n "\.ExecuteRuleActions(" business/sdk/workflow/engine.go

# Verify no commented-out loadRuleActions remains
grep -n "loadRuleActions" business/sdk/workflow/executor.go
```

All three commands should return zero results. Test files will be updated in Phase 4.

---

## Deliverables

- [ ] Updated `ExecuteRuleActionsGraph()` fallback at line 222-227 to return error with meaningful `BatchExecutionResult`
- [ ] Updated `ExecuteRuleActionsGraph()` fallback at line 264-268 to return error with meaningful `BatchExecutionResult`
- [ ] Updated `ExecuteRuleActionsGraph()` doc comment with error conditions
- [ ] Updated `engine.go` line 422 to call `ExecuteRuleActionsGraph()`
- [ ] Deleted `ExecuteRuleActions()` function from `executor.go` (lines 112-209)
- [ ] Deleted commented-out `loadRuleActions` function (lines 592-619)
- [ ] No production code references to `ExecuteRuleActions()` remain

---

## Validation Criteria

- [ ] `go build ./business/sdk/workflow/...` compiles successfully (production code only)
- [ ] `go vet ./business/sdk/workflow/...` passes with no warnings
- [ ] `grep -rn "ExecuteRuleActions[^G]" business/sdk/workflow/*.go | grep -v "_test.go"` returns zero results
- [ ] `grep -n "\.ExecuteRuleActions(" business/sdk/workflow/engine.go` returns zero results
- [ ] `ExecuteRuleActionsGraph()` returns error (not fallback) when edges are empty
- [ ] `ExecuteRuleActionsGraph()` returns error (not fallback) when startEdges are empty
- [ ] Error messages include function context: `"executeRuleActionsGraph: rule <uuid> has no edges..."`
- [ ] `BatchExecutionResult` in error cases has `Status="failed"` and populated `ErrorMessage`
- [ ] Doc comment for `ExecuteRuleActionsGraph()` does not mention backwards compatibility
- [ ] Doc comment documents error conditions and start edge definition
- [ ] Commented-out `loadRuleActions` function is removed

---

## Testing Strategy

### What to Test

- Graph execution with edges continues to work normally
- Missing edges returns a clear error (not a silent fallback)
- Missing start edges returns a clear error
- Engine routes all execution through graph executor
- Single-action rules with one start edge (minimal valid graph)
- Multi-action rules with complex graph (multiple paths, conditions)
- Rules with edges but none are start edges (all have non-nil source_action_id — should error)
- Error messages include rule ID for debugging
- `BatchExecutionResult` in error cases has `Status="failed"` and populated fields

### How to Test

```bash
# Compile check (most important - verifies no broken references)
go build ./business/sdk/workflow/...

# Vet check
go vet ./business/sdk/workflow/...

# Search for remaining references (should only find test files)
grep -rn "ExecuteRuleActions[^G]" business/sdk/workflow/ --include="*.go"

# Search for remaining commented-out code
grep -n "loadRuleActions" business/sdk/workflow/executor.go
```

### Integration Verification

After implementation, verify end-to-end execution:

```bash
# Run workflow engine tests that exercise rule execution via graph path
go test -v ./business/sdk/workflow/... -run TestGraphExec

# Note: Some graph tests that test fallback behavior will fail - this is expected
```

**Expected test failures** (to be fixed in Phase 4):
- `TestGraphExec_NoEdges_FallsBackToLinear` - tests the fallback we just removed
- `TestGraphExec_EmptyEdges_FallsBackToLinear` - tests the fallback we just removed
- `TestGraphExec_NoStartEdge_FallsBackToLinear` - tests the fallback we just removed
- `TestActionExecutor_Stats` - calls `ExecuteRuleActions()` directly
- `TestActionExecutor_ExecutionHistory` - calls `ExecuteRuleActions()` directly

---

## Security Note

Error messages include rule UUIDs for debugging purposes:
```
executeRuleActionsGraph: rule 550e8400-e29b-41d4-a716-446655440000 has no edges...
```

These errors are returned from the business layer and logged server-side. Ensure they are not returned directly to API clients without proper sanitization. The workflow API layer should wrap these errors using `errs.Newf()` before they reach clients.

---

## Rollback Strategy

If Phase 3 deployment causes unexpected issues:

### Immediate Rollback (Code Revert)
1. Revert the commit that removed `ExecuteRuleActions()`
2. Redeploy previous version
3. All rules continue to work (fallback path restored)

### Partial Rollback (Restore Fallback Only)
If only the fallback removal causes issues but the engine routing change is fine:
1. Restore the fallback logic in `ExecuteRuleActionsGraph()` (revert Task 1)
2. Keep the engine calling `ExecuteRuleActionsGraph()` (keep Task 2)
3. This provides backwards compatibility while investigating edge cases

### Prevention
- Run the database validation query (see Prerequisites) before deployment
- Monitor error rates after deployment for unexpected rule execution failures
- Have the rollback commit prepared before deploying

---

## Gotchas and Tips

- **Check for any other callers**: Use grep to find all callers of `ExecuteRuleActions` before deleting. The known callers are:
  1. `engine.go:422` (`executeRule` method)
  2. `executor.go:226` (first fallback in `ExecuteRuleActionsGraph`)
  3. `executor.go:267` (second fallback in `ExecuteRuleActionsGraph`)
  4. Test files (Phase 4)
- **Return meaningful results on error**: When returning errors from `ExecuteRuleActionsGraph`, always populate the `BatchExecutionResult` with `Status="failed"` and `ErrorMessage` so callers can inspect the result for logging/metrics.
- **Include function context in errors**: Prefix error messages with `executeRuleActionsGraph:` for better error traces.

---

## Reference

- Original plan: `.claude/plans/missing-action-edges.md`
- Progress tracking: `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
- Executor source: `business/sdk/workflow/executor.go`
- Engine source: `business/sdk/workflow/engine.go`

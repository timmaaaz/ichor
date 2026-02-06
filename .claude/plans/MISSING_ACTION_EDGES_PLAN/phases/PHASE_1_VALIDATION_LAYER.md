# Phase 1: Validation Layer Changes

**Category**: Backend
**Status**: Pending
**Dependencies**: None

---

## Overview

Enforce that workflow rules with actions must have edges connecting them, while allowing users to save incomplete workflows (trigger-only, no actions yet) for a consistent "you can always save" experience.

Currently, `ValidateGraph()` in `workflowsaveapp/graph.go` requires at least one action and at least one start edge. The `SaveWorkflowRequest` model also enforces `validate:"required,min=1"` on the `Actions` field, preventing saving workflows without actions entirely.

This phase makes three changes:
1. **Allow saving incomplete workflows**: Remove the `required,min=1` constraint on `Actions` so users can save a trigger with no actions (draft workflow).
2. **Auto-deactivate draft workflows**: Rules with no actions are forced inactive, preventing the engine from firing a rule that has nothing to do.
3. **Add explicit edge validation**: When actions *do* exist, require edges. Replace the misleading "exactly one start edge is required" error with a clear "rules with actions require edges" message.

### Goals

1. Allow incomplete/draft workflows to be saved (trigger configured, no actions yet)
2. Auto-deactivate rules with no actions (force `is_active = false`)
3. When actions exist but no edges are provided, return a clear error message
4. Skip graph validation entirely when no actions are present
5. Ensure the error message clearly communicates the requirement ("rules with actions require edges")

### Validation Matrix

| Actions | Edges | Result |
|---------|-------|--------|
| None    | None  | Save succeeds (draft workflow, forced inactive) |
| Present | Present | Validate graph structure, save succeeds if valid |
| Present | None  | Validation error: "rules with actions require at least one edge defining execution flow" |
| None    | Present | Edges are ignored (no actions to connect) — save succeeds (forced inactive) |

### Why This Phase Matters

This is the foundational change. Once validation rejects actions-without-edges, all subsequent phases (removing `execution_order`, removing the linear executor, updating seeds) become necessary to keep the system consistent. Without this validation, the other changes would be premature.

The draft workflow support also improves UX — non-technical users won't encounter confusing states where they can save sometimes but not others.

---

## Prerequisites

Before starting this phase, ensure:

- [ ] Go development environment is ready (`go version` shows 1.23+)
- [ ] You can build the project: `go build ./app/domain/workflow/workflowsaveapp/...`
- [ ] You understand the current `ValidateGraph()` function in `graph.go`
- [ ] You understand the current `SaveWorkflowRequest` model in `model.go`

---

## Task Breakdown

### Task 1: Allow Saving Workflows Without Actions

**Files**:
- `app/domain/workflow/workflowsaveapp/model.go` - Remove `required,min=1` from `Actions` validate tag

**Current Code** (line 19):
```go
Actions []SaveActionRequest `json:"actions" validate:"required,min=1,dive"`
```

**New Code**:
```go
Actions []SaveActionRequest `json:"actions" validate:"dive"`
```

**Key Considerations**:
- Removing `required,min=1` allows empty or nil `Actions` slices to pass request validation
- The `dive` tag still validates individual action items when present
- This enables saving draft/incomplete workflows (trigger only, no actions yet)

**Validation**:
- [ ] Code compiles: `go build ./app/domain/workflow/workflowsaveapp/...`

### Task 2: Skip Graph Validation When No Actions Exist

**Files**:
- `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` - Conditionally call `ValidateGraph()`

**Current Code** (lines 52-55 in `SaveWorkflow`, repeated at lines 123-126 in `CreateWorkflow`):
```go
// 3. Validate graph structure
if err := ValidateGraph(req.Actions, req.Edges); err != nil {
    return SaveWorkflowResponse{}, errs.Newf(errs.InvalidArgument, "graph: %s", err)
}
```

**New Code**:
```go
// 3. Validate graph structure (only when actions exist)
if len(req.Actions) > 0 {
    if err := ValidateGraph(req.Actions, req.Edges); err != nil {
        return SaveWorkflowResponse{}, errs.Newf(errs.InvalidArgument, "graph: %s", err)
    }
}
```

**Key Considerations**:
- Both `SaveWorkflow()` and `CreateWorkflow()` need the same change
- When no actions are present, graph validation is meaningless — skip it entirely
- `ValidateActionConfigs()` (called just before) should also be wrapped in the same guard since it validates action configs

**Where to apply** (both call sites):
- `workflowsaveapp.go:48-55` - `SaveWorkflow()` (validate action configs + graph)
- `workflowsaveapp.go:119-126` - `CreateWorkflow()` (validate action configs + graph)

**Validation**:
- [ ] Code compiles: `go build ./app/domain/workflow/workflowsaveapp/...`

### Task 3: Auto-Deactivate Rules With No Actions

**Files**:
- `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` - Force `is_active = false` when no actions

**Implementation**:

In both `SaveWorkflow()` and `CreateWorkflow()`, after request validation but before the transaction, force the rule to be inactive when there are no actions:

```go
// If no actions, force rule to inactive (draft workflow)
if len(req.Actions) == 0 {
	req.IsActive = false
}
```

**Key Considerations**:
- This prevents the engine from ever firing a rule that has no actions to execute
- The user can still save the workflow — they just can't activate it until actions are added
- When the user later adds actions + edges and saves again, they can set `is_active = true`
- This is a silent override, not a validation error — the save succeeds, the rule is just forced inactive
- Place this before the transaction begins so the forced value flows through to the rule create/update

**Validation**:
- [ ] Code compiles: `go build ./app/domain/workflow/workflowsaveapp/...`
- [ ] Saving a draft workflow with `is_active: true` results in `is_active: false` in the response

### Task 4: Add Edge Requirement Validation to `ValidateGraph()`

**Files**:
- `app/domain/workflow/workflowsaveapp/graph.go` - Add early check for actions without edges

**Current Code** (lines 14-17):
```go
func ValidateGraph(actions []SaveActionRequest, edges []SaveEdgeRequest) error {
	if len(actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}
```

**New Code**:
```go
func ValidateGraph(actions []SaveActionRequest, edges []SaveEdgeRequest) error {
	if len(actions) == 0 {
		return nil // No actions means no graph to validate
	}

	// Rules with actions must have edges defining the execution flow.
	if len(edges) == 0 {
		return fmt.Errorf("rules with actions require at least one edge defining execution flow")
	}

	// Count start edges
	startEdgeCount := 0
	// ... rest of existing logic unchanged
```

**Key Considerations**:
- The `len(actions) == 0` case now returns `nil` instead of an error — this is a defensive guard since the caller already skips `ValidateGraph` when there are no actions (Task 2), but keeps the function safe if called directly
- When actions exist but edges don't, we return a clear error before reaching the "exactly one start edge is required" check
- The new error message is distinct and clearly communicates what's needed
- **Nil vs empty slices**: `len(edges) == 0` correctly handles both `nil` and empty slices. JSON unmarshaling of `"edges": []` creates an empty slice, which is the expected case

**Validation**:
- [ ] Code compiles: `go build ./app/domain/workflow/workflowsaveapp/...`

### Task 5: Verify Integration Test Impact

**Files**:
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/*_test.go` - Identify affected tests

**Implementation**:

Before running tests after the above changes:

1. Search for test cases that create actions without edges:
   ```bash
   grep -rn "SaveActionRequest" api/cmd/services/ichor/tests/workflow/workflowsaveapi/
   grep -rn "Edges" api/cmd/services/ichor/tests/workflow/workflowsaveapi/
   ```

2. Document which test functions will fail after implementing validation

3. These failures are expected and will be addressed in Phase 4 (Test Updates) and Phase 5 (Seed Data Updates)

**Validation**:
- [ ] List of affected test functions documented
- [ ] Expected failure pattern confirmed (should error with "rules with actions require...")

---

## Deliverables

- [ ] Updated `model.go`: Removed `required,min=1` from `Actions` validate tag
- [ ] Updated `workflowsaveapp.go`: Skip graph/action-config validation when no actions present (both call sites)
- [ ] Updated `workflowsaveapp.go`: Force `is_active = false` when no actions (both call sites)
- [ ] Updated `graph.go`: Return `nil` for empty actions, add explicit edge requirement check
- [ ] Documented list of tests expected to break (for Phase 4/5)

---

## Validation Criteria

- [ ] Go compilation passes: `go build ./app/domain/workflow/workflowsaveapp/...`
- [ ] Saving a workflow with no actions and no edges succeeds (draft workflow)
- [ ] Draft workflow is forced inactive regardless of `is_active` in request
- [ ] Saving a workflow with actions and edges succeeds (existing behavior)
- [ ] Saving a workflow with actions but no edges returns clear error
- [ ] New validation error message: "rules with actions require at least one edge defining execution flow"

---

## Testing Strategy

### What to Test

- **No actions, no edges**: Should save successfully (new behavior — draft workflow)
- **No actions, is_active=true**: Should save successfully but `is_active` forced to `false` in response
- **Actions with edges**: Should pass validation (existing behavior, unchanged)
- **Actions without edges**: Should fail with the new error message
- **Cycle detection**: Should still work (existing behavior, unchanged)
- **Unreachable actions**: Should still be caught (existing behavior, unchanged)

### How to Test

```bash
# Build the package
go build ./app/domain/workflow/workflowsaveapp/...

# Run unit tests for the package
go test -v ./app/domain/workflow/workflowsaveapp/...

# Run the full workflow test suite
go test -v ./api/cmd/services/ichor/tests/workflow/...

# Run all tests to check for regressions
make test
```

### Manual Verification

After implementation, verify that `ValidateGraph`:
1. Returns `nil` when `actions` is empty (draft workflow)
2. Returns error "rules with actions require at least one edge defining execution flow" when actions exist but edges is empty
3. Returns error "exactly one start edge is required" when edges exist but none is a start edge
4. Returns `nil` for valid graphs (actions + proper edges)

---

## Gotchas and Tips

- **Three files, not just one**: This phase touches `model.go`, `workflowsaveapp.go`, and `graph.go`. Don't forget the model change.
- **Two call sites in workflowsaveapp.go**: Both `SaveWorkflow()` and `CreateWorkflow()` need the `len(req.Actions) > 0` guard. Missing one creates an inconsistency.
- **Don't change the ValidateGraph function signature**: `ValidateGraph(actions []SaveActionRequest, edges []SaveEdgeRequest) error` stays the same.
- **Existing tests may break**: Tests that create actions without edges will now fail. That's expected — Phase 4 and Phase 5 address test updates. For Phase 1, focus on the validation changes and verify with `go build`.
- **Error message matters**: The error message should be user-facing friendly. "rules with actions require at least one edge defining execution flow" is clear about what's wrong and what's needed.
- **Nil vs empty slices**: `len(edges) == 0` correctly handles both `nil` and empty `[]SaveEdgeRequest{}`. No special handling needed.
- **Production data risk**: If production contains workflow rules with actions but no edges, updates to those rules will be blocked by the new validation. Run this query before deploying:
  ```sql
  SELECT r.id, r.name,
         COUNT(DISTINCT a.id) as action_count,
         COUNT(DISTINCT e.id) as edge_count
  FROM workflow.automation_rules r
  LEFT JOIN workflow.rule_actions a ON r.id = a.automation_rule_id
  LEFT JOIN workflow.action_edges e ON e.rule_id = r.id
  GROUP BY r.id, r.name
  HAVING COUNT(DISTINCT a.id) > 0 AND COUNT(DISTINCT e.id) = 0;
  ```

---

## Reference

- Original plan: `.claude/plans/missing-action-edges.md`
- Progress tracking: `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
- Source files:
  - `app/domain/workflow/workflowsaveapp/model.go` (line 19, `Actions` field)
  - `app/domain/workflow/workflowsaveapp/graph.go` (line 14, `ValidateGraph` function)
  - `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` (lines 53 and 124, call sites)

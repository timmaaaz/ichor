# Phase 4: Test Updates

**Category**: Testing
**Status**: Pending
**Dependencies**: Phase 1 (Validation Layer), Phase 2 (Remove execution_order), Phase 3 (Remove Linear Executor)

---

## Overview

Remove obsolete tests that validate linear fallback behavior, update existing tests to use the graph executor with proper edges, and add a new validation test confirming that actions without edges are rejected.

After Phases 1-3, the linear executor no longer exists and `execution_order` has been removed. Tests that rely on either will fail or reference deleted code. This phase brings the test suite into alignment with the new single-execution-mode architecture.

### Goals

1. Delete 3 linear fallback test functions from `executor_graph_test.go`
2. Update `createTestRule` helper in `executor_graph_test.go` to remove `ExecutionOrder`
3. Update 2 test functions in `executor_test.go` to use `ExecuteRuleActionsGraph()` with edges
4. Add 1 new validation test confirming actions-without-edges rejection
5. Verify no remaining `ExecutionOrder` references in existing validation tests

### Why This Phase Matters

Without updating the tests, `make test` will fail after Phases 1-3. The test suite is the primary safety net for the workflow engine, and broken tests undermine confidence in the refactoring. This phase also adds the missing negative test case that validates the new edge requirement from Phase 1.

---

## Prerequisites

Before starting this phase, ensure:

- [ ] Phase 1 (Validation Layer Changes) is completed
- [ ] Phase 2 (Remove execution_order Field) is completed
- [ ] Phase 3 (Remove Linear Executor) is completed
- [ ] Go development environment is ready (`go version` shows 1.23+)
- [ ] Project compiles after Phases 1-3: `go build ./...`
- [ ] Phase 1 wraps `ValidateGraph` errors with `errs.InvalidArgument` (confirmed: `workflowsaveapp.go` line 54 and 131)
- [ ] Phase 1 implements draft workflow auto-deactivation (`is_active=false` when no actions) (confirmed: `workflowsaveapp.go` lines 59-61)

---

## Task Breakdown

### Task 1: Delete Linear Fallback Tests

**Files**:
- `business/sdk/workflow/executor_graph_test.go` - Delete 3 test functions

**Implementation**:

Delete these three test functions entirely:

1. **`TestGraphExec_NoEdges_FallsBackToLinear`** (lines 246-293)
   - Tests that when a rule has actions but NO edges, execution falls back to linear mode
   - This behavior no longer exists — actions without edges now produce an error at validation time

2. **`TestGraphExec_EmptyEdges_FallsBackToLinear`** (lines 295-328)
   - Tests that when edges slice is explicitly empty (`[]testEdge{}`), execution falls back to linear
   - Same as above — empty edges with actions is now a validation error

3. **`TestGraphExec_NoStartEdge_FallsBackToLinear`** (lines 421-459)
   - Tests that when edges exist but no start edge is present, execution falls back to linear
   - No start edge is now a graph validation error, not a fallback scenario

Also update the section comment above these tests. The "Backwards Compatibility Tests" header (line 242-244) should be removed since the section no longer exists.

**Validation**:
- [ ] All three functions are completely removed
- [ ] The "Backwards Compatibility Tests" section header is removed
- [ ] The file still compiles: `go build ./business/sdk/workflow/...`
- [ ] Remaining graph tests still pass conceptually (full test run later)

---

### Task 2: Update `createTestRule` Helper to Remove `ExecutionOrder`

**Files**:
- `business/sdk/workflow/executor_graph_test.go` - Modify line 196

**Implementation**:

The `createTestRule` helper function (line 108) is used by 20+ tests in this file. It creates rule actions in a loop and currently sets `ExecutionOrder: i + 1` (line 196). After Phase 2, this field no longer exists on `NewRuleAction`.

Remove the `ExecutionOrder` line:
```go
// BEFORE (line 196):
ExecutionOrder:   i + 1,

// AFTER: simply delete this line
```

**Important**: This helper is used extensively. After this change, run a broad compilation check to ensure all callers still work. The function signature and return values do NOT change — only the internal `NewRuleAction` struct loses a field.

**Validation**:
- [ ] `ExecutionOrder` reference removed from `createTestRule` helper
- [ ] File compiles: `go build ./business/sdk/workflow/...`
- [ ] All existing graph tests that use `createTestRule` still compile

---

### Task 3: Update `TestActionExecutor_Stats` to Use Graph Executor

**Files**:
- `business/sdk/workflow/executor_test.go` - Modify lines 765-977

**Implementation**:

This test currently:
- Creates 3 `NewRuleAction` structs with `ExecutionOrder` fields (lines 877-913)
- Calls `ae.ExecuteRuleActions()` (line 936) — the linear executor

Changes needed:

1. **Remove `ExecutionOrder` from `NewRuleAction` structs** (3 occurrences):
   ```go
   // BEFORE:
   {
       AutomationRuleID: rule.ID,
       Name:             "Success Action 1",
       ActionConfig:     json.RawMessage(`{...}`),
       ExecutionOrder:   1,  // REMOVE THIS LINE
       IsActive:         true,
       TemplateID:       &template.ID,
   },

   // AFTER:
   {
       AutomationRuleID: rule.ID,
       Name:             "Success Action 1",
       ActionConfig:     json.RawMessage(`{...}`),
       IsActive:         true,
       TemplateID:       &template.ID,
   },
   ```

2. **Collect action IDs** returned from `CreateRuleAction`. Replace the existing loop (lines 915-920) that discards the return value:
   ```go
   // BEFORE:
   for _, action := range actions {
       _, err := workflowBus.CreateRuleAction(ctx, action)
       if err != nil {
           t.Fatalf("Failed to create rule action: %v", err)
       }
   }

   // AFTER:
   var actionIDs []uuid.UUID
   for _, action := range actions {
       created, err := workflowBus.CreateRuleAction(ctx, action)
       if err != nil {
           t.Fatalf("Failed to create rule action: %v", err)
       }
       actionIDs = append(actionIDs, created.ID)
   }
   ```

3. **Create edges** connecting the actions. Use `RuleID` (NOT `AutomationRuleID`) and note there is NO `CreatedBy` field on `NewActionEdge`:
   ```go
   // Start -> Action 0
   _, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
       RuleID:         rule.ID,
       TargetActionID: actionIDs[0],
       EdgeType:       workflow.EdgeTypeStart,
       EdgeOrder:      1,
   })
   if err != nil {
       t.Fatalf("Failed to create start edge: %v", err)
   }

   // Action 0 -> Action 1
   _, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
       RuleID:         rule.ID,
       SourceActionID: &actionIDs[0],
       TargetActionID: actionIDs[1],
       EdgeType:       workflow.EdgeTypeSequence,
       EdgeOrder:      1,
   })
   if err != nil {
       t.Fatalf("Failed to create edge: %v", err)
   }

   // Action 1 -> Action 2
   _, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
       RuleID:         rule.ID,
       SourceActionID: &actionIDs[1],
       TargetActionID: actionIDs[2],
       EdgeType:       workflow.EdgeTypeSequence,
       EdgeOrder:      1,
   })
   if err != nil {
       t.Fatalf("Failed to create edge: %v", err)
   }
   ```

4. **Change executor call** from linear to graph:
   ```go
   // BEFORE:
   result, err := ae.ExecuteRuleActions(ctx, rule.ID, execContext)

   // AFTER:
   result, err := ae.ExecuteRuleActionsGraph(ctx, rule.ID, execContext)
   ```

5. **Add `TriggerSource` to `ActionExecutionContext`**. This field is currently missing from this test's `execContext` (lines 923-934) but exists on the struct. Add it:
   ```go
   execContext := workflow.ActionExecutionContext{
       EntityID:    entity.ID,
       EntityName:  "customers",
       EventType:   "on_create",
       UserID:      uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
       RuleID:      &rule.ID,
       ExecutionID: uuid.New(),
       Timestamp:   time.Now(),
       TriggerSource: workflow.TriggerSourceAutomation,
       RawData: map[string]interface{}{
           "test": "data",
       },
   }
   ```

**Note**: The test expectations (TotalActions=3, SuccessfulActions=2, FailedActions=1) should remain the same since graph execution preserves action count and success/failure semantics.

**Validation**:
- [ ] No references to `ExecutionOrder` remain in the test
- [ ] No calls to `ExecuteRuleActions()` (linear) remain — uses `ExecuteRuleActionsGraph()` instead
- [ ] Edges are created forming a chain: Start -> Action0 -> Action1 -> Action2
- [ ] Uses `RuleID` (not `AutomationRuleID`) and no `CreatedBy` in edge creation
- [ ] `TriggerSource` is set on `ActionExecutionContext`
- [ ] Test still validates the same stats expectations

---

### Task 4: Update `TestActionExecutor_ExecutionHistory` to Use Graph Executor

**Files**:
- `business/sdk/workflow/executor_test.go` - Modify lines 979-1166

**Implementation**:

This test creates 5 rules in a loop, each with `i+1` actions. Changes needed:

1. **Remove `ExecutionOrder` from `NewRuleAction` and capture return value** (line 1100-1107). Replace the existing `CreateRuleAction` call that discards the return value with one that captures it:
   ```go
   // BEFORE:
   _, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
       AutomationRuleID: rule.ID,
       Name:             fmt.Sprintf("Action %d", j),
       ActionConfig:     actionConfig,
       ExecutionOrder:   j + 1,  // REMOVE THIS LINE
       IsActive:         true,
       TemplateID:       &template.ID,
   })

   // AFTER:
   created, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
       AutomationRuleID: rule.ID,
       Name:             fmt.Sprintf("Action %d", j),
       ActionConfig:     actionConfig,
       IsActive:         true,
       TemplateID:       &template.ID,
   })
   ```

2. **Collect action IDs and create edges after the inner action creation loop**. Declare `actionIDs` before the inner loop, append inside it, then create edges after:
   ```go
   // Declare before inner loop
   var actionIDs []uuid.UUID

   // Inside inner loop (j := 0; j <= i; j++), after creating each action:
   actionIDs = append(actionIDs, created.ID)

   // After inner loop completes, create edges: Start -> Action0 -> Action1 -> ... -> ActionN
   _, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
       RuleID:         rule.ID,
       TargetActionID: actionIDs[0],
       EdgeType:       workflow.EdgeTypeStart,
       EdgeOrder:      1,
   })
   if err != nil {
       t.Fatalf("Failed to create start edge for rule %d: %v", i, err)
   }

   for k := 0; k < len(actionIDs)-1; k++ {
       _, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
           RuleID:         rule.ID,
           SourceActionID: &actionIDs[k],
           TargetActionID: actionIDs[k+1],
           EdgeType:       workflow.EdgeTypeSequence,
           EdgeOrder:      1,
       })
       if err != nil {
           t.Fatalf("Failed to create edge for rule %d: %v", i, err)
       }
   }
   ```

3. **Change executor call** from linear to graph:
   ```go
   // BEFORE:
   _, err = ae.ExecuteRuleActions(ctx, rule.ID, execContext)

   // AFTER:
   _, err = ae.ExecuteRuleActionsGraph(ctx, rule.ID, execContext)
   ```

4. **Add `TriggerSource` to `ActionExecutionContext`** (lines 1114-1122). This field is currently missing from this test:
   ```go
   execContext := workflow.ActionExecutionContext{
       EntityID:    entity.ID,
       EntityName:  "customers",
       EventType:   "on_create",
       UserID:      userID,
       RuleID:      &rule.ID,
       ExecutionID: uuid.New(),
       Timestamp:   time.Now(),
       TriggerSource: workflow.TriggerSourceAutomation,
   }
   ```

**Validation**:
- [ ] No references to `ExecutionOrder` remain in the test
- [ ] No calls to `ExecuteRuleActions()` (linear) remain
- [ ] Each rule gets edges: Start -> Action0, Action0 -> Action1, etc.
- [ ] Uses `RuleID` (not `AutomationRuleID`) and no `CreatedBy` in edge creation
- [ ] `TriggerSource` is set on `ActionExecutionContext`
- [ ] History expectations remain unchanged (5 rules, 3 most recent, clear history)

---

### Task 5: Add Validation Test for Actions Without Edges

**Origin**: Phase 1 re-review recommendation (B+ → A upgrade; these tests would bring Phase 1 to A+).

**Files**:
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/validation_test.go` - Add new test table function

**Implementation**:

Add a new test table function that validates the Phase 1 validation matrix. This test should be added to the existing `validation_test.go` file alongside the existing `validationActionConfig` and `validationGraph` functions.

These two test cases directly verify the Phase 1 validation matrix and were identified as the only remaining gap in the Phase 1 re-review (upgraded B+ → A, would be A+ with these tests).

```go
// =============================================================================
// Edge Requirement Validation Tests
// =============================================================================

// validationEdgeRequirement tests the Phase 1 validation matrix:
//   - Actions without edges → rejected (InvalidArgument)
//   - No actions, no edges  → allowed (draft workflow, forced inactive)
//
// Added per Phase 1 re-review recommendation (B+ → A upgrade).
func validationEdgeRequirement(sd SaveSeedData) []apitest.Table {
    if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
        return nil
    }

    return []apitest.Table{
        {
            Name:       "actions-without-edges-rejected",
            URL:        "/v1/workflow/rules/full",
            Token:      sd.Users[0].Token,
            StatusCode: http.StatusBadRequest,
            Method:     http.MethodPost,
            Input: workflowsaveapp.SaveWorkflowRequest{
                Name:          "Actions Without Edges",
                IsActive:      true,
                EntityID:      sd.Entities[0].ID.String(),
                TriggerTypeID: sd.TriggerTypes[0].ID.String(),
                Actions: []workflowsaveapp.SaveActionRequest{
                    {
                        Name:       "Orphan Action",
                        ActionType: "create_alert",
                        IsActive:   true,
                        ActionConfig: json.RawMessage(`{
                            "alert_type": "test",
                            "severity": "info",
                            "title": "Test",
                            "message": "This should fail"
                        }`),
                    },
                },
                // No Edges field — actions exist but no edges
            },
            GotResp: &errs.Error{},
            ExpResp: &errs.Error{},
            CmpFunc: func(got any, exp any) string {
                gotErr, ok := got.(*errs.Error)
                if !ok {
                    return "failed to cast to error"
                }
                if gotErr.Code != errs.InvalidArgument {
                    return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
                }
                if !strings.Contains(strings.ToLower(gotErr.Error()), "edge") {
                    return "expected error message to mention 'edge'"
                }
                return ""
            },
        },
        {
            Name:       "no-actions-no-edges-allowed",
            URL:        "/v1/workflow/rules/full",
            Token:      sd.Users[0].Token,
            StatusCode: http.StatusOK,
            Method:     http.MethodPost,
            Input: workflowsaveapp.SaveWorkflowRequest{
                Name:          "Draft Workflow No Actions",
                IsActive:      true, // User requests active — should be forced inactive
                EntityID:      sd.Entities[0].ID.String(),
                TriggerTypeID: sd.TriggerTypes[0].ID.String(),
                // No Actions, No Edges — draft workflow
            },
            GotResp: &workflowsaveapp.SaveWorkflowResponse{},
            ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
            CmpFunc: func(got any, exp any) string {
                gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
                if !ok {
                    return "failed to cast response"
                }
                if gotResp.ID == "" {
                    return "expected draft workflow to be created"
                }
                if gotResp.IsActive {
                    return "expected draft workflow to be forced inactive (is_active should be false)"
                }
                if len(gotResp.Actions) != 0 {
                    return fmt.Sprintf("expected zero actions, got %d", len(gotResp.Actions))
                }
                if len(gotResp.Edges) != 0 {
                    return fmt.Sprintf("expected zero edges, got %d", len(gotResp.Edges))
                }
                return ""
            },
        },
    }
}
```

**Wire the new test function**: Find the test runner in the main test file (likely `workflowsave_test.go` or similar) and add:
```go
test.Run(t, validationEdgeRequirement(sd), "validation-edge-requirement")
```

**Note**: The `ExecutionOrder` field in `SaveActionRequest` will have been removed by Phase 2, so the new test function should NOT include it.

**Note on error wrapping**: The validation test expects `errs.InvalidArgument` because `workflowsaveapp.go` wraps `ValidateGraph` errors with `errs.Newf(errs.InvalidArgument, "graph: %s", err)` (lines 54 and 131). This is already implemented in Phase 1.

**Note on draft deactivation**: The "no-actions-no-edges-allowed" test expects `is_active=false` because `workflowsaveapp.go` forces `req.IsActive = false` when `len(req.Actions) == 0` (lines 59-61). This is already implemented in Phase 1.

**Validation**:
- [ ] `actions-without-edges-rejected` test case returns 400 with InvalidArgument and error message mentions "edge"
- [ ] `no-actions-no-edges-allowed` test case returns 200, creates an inactive draft workflow, and verifies zero actions/edges in response
- [ ] New test function is wired into the test runner
- [ ] File compiles: `go build ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...`

---

### Task 6: Verify No Remaining `ExecutionOrder` in Test Files

**Files**:
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/validation_test.go` - Check for stale `ExecutionOrder` references

**Implementation**:

After Phase 2, all test files referencing `ExecutionOrder` will fail to compile. Run a compilation check and grep to identify any files needing updates beyond the explicit tasks above.

```bash
# Check for any remaining ExecutionOrder references in test files
grep -rn "ExecutionOrder" business/sdk/workflow/*_test.go
grep -rn "ExecutionOrder" api/cmd/services/ichor/tests/workflow/

# Compile check
go build ./...
```

Known files that may still reference `ExecutionOrder` in validation tests or seed data:
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/validation_test.go` (multiple occurrences in existing `validationActionConfig` and `validationGraph` functions)

If any references remain, remove them. The compiler will catch all of them since `ExecutionOrder` no longer exists on the struct after Phase 2.

**Validation**:
- [ ] `grep -rn "ExecutionOrder" business/sdk/workflow/*_test.go` returns zero results (excluding `ExecutionOrderBy*` constants)
- [ ] `grep -rn "ExecutionOrder" api/cmd/services/ichor/tests/workflow/` returns zero results
- [ ] `go build ./...` passes

---

## Deliverables

- [ ] Removed 3 obsolete linear fallback test functions from `executor_graph_test.go`
- [ ] Removed "Backwards Compatibility Tests" section header from `executor_graph_test.go`
- [ ] Updated `createTestRule` helper in `executor_graph_test.go` to remove `ExecutionOrder` (line 196)
- [ ] Updated `TestActionExecutor_Stats` to use `ExecuteRuleActionsGraph()` with edges
- [ ] Updated `TestActionExecutor_ExecutionHistory` to use `ExecuteRuleActionsGraph()` with edges
- [ ] Added `validationEdgeRequirement` test function with 2 test cases
- [ ] Wired new test function into the test runner
- [ ] Verified no remaining `ExecutionOrder` references in test files

---

## Validation Criteria

- [ ] Go compilation passes: `go build ./...`
- [ ] All workflow tests pass: `go test ./business/sdk/workflow/...`
- [ ] All workflowsaveapi tests pass: `go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...`
- [ ] No test references `ExecuteRuleActions()` (the linear executor)
- [ ] No test references `ExecutionOrder` (the removed field, except `ExecutionOrderBy*` constants)
- [ ] No test helper functions reference `ExecutionOrder`
- [ ] `createTestRule` in `executor_graph_test.go` compiles without `ExecutionOrder`
- [ ] New validation test passes (actions without edges rejected, draft workflows allowed)
- [ ] Draft workflow test passes (verifies Phase 1 auto-deactivation)
- [ ] Full test suite: `make test`

---

## Testing Strategy

### What to Test

- **Deletion safety**: Ensure removing fallback tests doesn't break other tests in the same file
- **Helper update safety**: `createTestRule` is used by 20+ tests — verify all still compile after removing `ExecutionOrder`
- **Edge creation**: Verify that the edge creation code uses `RuleID` (not `AutomationRuleID`) and has no `CreatedBy` field
- **Stats correctness**: `TestActionExecutor_Stats` should produce identical stat values with graph execution
- **History correctness**: `TestActionExecutor_ExecutionHistory` should produce identical history entries with graph execution
- **Validation rejection**: New test confirms API returns 400 when actions have no edges
- **Draft workflow**: New test confirms saving a trigger-only workflow succeeds and forces `is_active=false`

### How to Test

```bash
# Compile check
go build ./...

# Run specific test files
go test -v ./business/sdk/workflow/ -run TestGraphExec
go test -v ./business/sdk/workflow/ -run TestActionExecutor_Stats
go test -v ./business/sdk/workflow/ -run TestActionExecutor_ExecutionHistory
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/ -run validation

# Full workflow test suite
go test ./business/sdk/workflow/...
go test ./api/cmd/services/ichor/tests/workflow/...

# Full test suite
make test
```

---

## Gotchas and Tips

- **`ExecutionOrder` is gone**: After Phase 2, the `NewRuleAction` struct no longer has an `ExecutionOrder` field. If you try to set it, the code won't compile. This is intentional — the compiler will catch any missed references.
- **`ExecuteRuleActions()` is gone**: After Phase 3, the linear executor function is deleted. Any call to it will be a compile error. Use `ExecuteRuleActionsGraph()` instead.
- **Use `RuleID`, not `AutomationRuleID`**: The `NewActionEdge` struct field is `RuleID`. Do NOT use `AutomationRuleID` — that field does not exist on this struct.
- **No `CreatedBy` on `NewActionEdge`**: The `NewActionEdge` struct does NOT have a `CreatedBy` field. The struct fields are: `RuleID`, `SourceActionID`, `TargetActionID`, `EdgeType`, `EdgeOrder`.
- **Start edge has no SourceActionID**: When creating a start edge, `SourceActionID` should be `nil` (it's a `*uuid.UUID`). Only `TargetActionID` is set.
- **Test isolation**: Each test creates its own database via `dbtest.NewDatabase()`, so changes to one test won't affect others.
- **Do NOT use `t.Parallel()`**: Both `TestActionExecutor_Stats` and `TestActionExecutor_ExecutionHistory` share RabbitMQ test infrastructure and must run serially. Do not add `t.Parallel()`.
- **RabbitMQ dependency**: Both `TestActionExecutor_Stats` and `TestActionExecutor_ExecutionHistory` require a RabbitMQ test container. Don't remove the RabbitMQ setup code.
- **`createTestRule` is used by 20+ tests**: When modifying this helper, verify broadly that all callers still compile. The function signature does NOT change — only the internal `NewRuleAction` struct loses the `ExecutionOrder` field.
- **Action result order may differ**: Graph execution follows edges, not creation order. However, since we're creating a linear chain (Start -> A -> B -> C), the order should match the previous linear execution order.
- **Validation tests depend on Phase 1 error wrapping**: The new validation test expects `errs.InvalidArgument`. This works because `workflowsaveapp.go` wraps `ValidateGraph` errors with `errs.Newf(errs.InvalidArgument, ...)` — confirmed already implemented in Phase 1.
- **Check `SaveActionResponse` for `ExecutionOrder`**: After Phase 2, the `SaveActionResponse` no longer contains `ExecutionOrder`. If validation test assertions check response fields, ensure they don't reference it.
- **Existing validation tests may need `ExecutionOrder` removed**: The existing `validationActionConfig` and `validationGraph` functions in `validation_test.go` may reference `ExecutionOrder` — the compiler will catch these. Task 6 handles this.

---

## Reference

- Original plan: `.claude/plans/missing-action-edges.md`
- Progress tracking: `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
- Phase 1 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_1_VALIDATION_LAYER.md`
- Phase 2 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_2_REMOVE_EXECUTION_ORDER.md`
- Phase 3 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_3_REMOVE_LINEAR_EXECUTOR.md`
- `NewActionEdge` struct: `business/sdk/workflow/models.go` lines 408-414
- `createTestRule` helper: `business/sdk/workflow/executor_graph_test.go` lines 108-240
- Error wrapping: `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` lines 53-54, 130-131
- Draft deactivation: `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` lines 58-61

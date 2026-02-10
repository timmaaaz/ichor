# Phase 5: Seed Data Updates

**Category**: Backend
**Status**: Pending
**Dependencies**: Phase 1 (Validation Layer), Phase 2 (Remove execution_order), Phase 3 (Remove Linear Executor), Phase 4 (Test Updates)

---

## Overview

All seeded workflow rules that create actions must also create edges connecting those actions. After Phases 1-3, the engine requires edges for execution and the save API rejects actions without edges. Any seed function that creates actions without edges will either fail at runtime or leave the database in an inconsistent state.

This phase updates 6 seed functions and reviews 5 additional test files to ensure every action has a corresponding edge chain.

### Goals

1. Add edge creation to `TestSeedRuleActions()` and `TestSeedFullWorkflow()` in `testutil.go`
2. Add edge creation to `ruleapi/seed_test.go`, `cascade_seed_test.go`, and `executionapi/seed_test.go`
3. Add edge creation to inline rule creation in `ordersapi/workflow_test.go` and `formdataapi/workflow_test.go`
4. Verify workflowsaveapi test files already have edges (they do — no changes needed)

### Why This Phase Matters

Seed functions are the foundation of the test suite. If they create actions without edges, then:
- Tests using `TestSeedFullWorkflow()` will fail when the engine tries to execute rules
- Tests using `TestSeedRuleActions()` will create orphaned actions that violate the new edge requirement
- Integration tests in ruleapi, executionapi, ordersapi, and formdataapi will fail
- `make seed` may produce an inconsistent database state

---

## Prerequisites

Before starting this phase, ensure:

- [ ] Phase 1 (Validation Layer Changes) is completed
- [ ] Phase 2 (Remove execution_order Field) is completed
- [ ] Phase 3 (Remove Linear Executor) is completed
- [ ] Phase 4 (Test Updates) is completed
- [ ] Go development environment is ready (`go version` shows 1.23+)
- [ ] Project compiles after Phases 1-4: `go build ./...`
- [ ] All Phase 4 tests pass: `make test`

---

## Edge Creation Pattern Reference

All edge creation in this phase follows the same pattern established in existing code. Use this as a reference for every task.

**`NewActionEdge` struct** (`business/sdk/workflow/models.go` lines 408-414):
```go
type NewActionEdge struct {
    RuleID         uuid.UUID
    SourceActionID *uuid.UUID  // nil for start edges
    TargetActionID uuid.UUID
    EdgeType       string      // "start", "sequence", "true_branch", "false_branch", "always"
    EdgeOrder      int
}
```

**`CreateActionEdge` method** (`business/sdk/workflow/workflowbus.go` lines 1170-1180):
```go
func (b *Business) CreateActionEdge(ctx context.Context, nae NewActionEdge) (ActionEdge, error)
```

**Standard linear chain pattern** (from `workflowsaveapi/seed_test.go` `seedEdgesForRule` helper):
```go
// Start edge: nil -> first action
_, err := wfBus.CreateActionEdge(ctx, workflow.NewActionEdge{
    RuleID:         ruleID,
    SourceActionID: nil,
    TargetActionID: actions[0].ID,
    EdgeType:       "start",
    EdgeOrder:      0,
})

// Sequence edges: action[i] -> action[i+1]
for i := 0; i < len(actions)-1; i++ {
    sourceID := actions[i].ID
    _, err := wfBus.CreateActionEdge(ctx, workflow.NewActionEdge{
        RuleID:         ruleID,
        SourceActionID: &sourceID,
        TargetActionID: actions[i+1].ID,
        EdgeType:       "sequence",
        EdgeOrder:      i + 1,
    })
}
```

**Key rules**:
- Use `RuleID`, NOT `AutomationRuleID`
- No `CreatedBy` field on `NewActionEdge`
- Start edges have `SourceActionID: nil`
- Edge types: use string constants `"start"`, `"sequence"` (or `workflow.EdgeTypeStart`, `workflow.EdgeTypeSequence`)

---

## Task Breakdown

### Task 1: Update `TestSeedRuleActions()` in `testutil.go`

**Files**:
- `business/sdk/workflow/testutil.go` - Lines 228-243

**Current behavior**:

`TestSeedRuleActions()` creates `n` actions distributed across the provided `ruleIDs` via `TestNewRuleActions()` and `CreateRuleAction()`. It returns the created actions but does NOT create any edges.

```go
func TestSeedRuleActions(ctx context.Context, n int, ruleIDs []uuid.UUID, templateIDs *[]uuid.UUID, api *Business) ([]RuleAction, error) {
    newActions := TestNewRuleActions(n, ruleIDs, templateIDs)
    actions := make([]RuleAction, len(newActions))
    for i, nra := range newActions {
        action, err := api.CreateRuleAction(ctx, nra)
        // ...
        actions[i] = action
    }
    return actions, nil
}
```

**Implementation**:

After creating all actions, group them by rule ID and create edge chains for each rule:

```go
func TestSeedRuleActions(ctx context.Context, n int, ruleIDs []uuid.UUID, templateIDs *[]uuid.UUID, api *Business) ([]RuleAction, error) {
    newActions := TestNewRuleActions(n, ruleIDs, templateIDs)

    actions := make([]RuleAction, len(newActions))
    for i, nra := range newActions {
        action, err := api.CreateRuleAction(ctx, nra)
        if err != nil {
            return nil, fmt.Errorf("seeding rule action: idx: %d : %w", i, err)
        }
        actions[i] = action
    }

    // Group actions by rule ID for edge creation
    actionsByRule := make(map[uuid.UUID][]RuleAction)
    for _, a := range actions {
        actionsByRule[a.AutomationRuleID] = append(actionsByRule[a.AutomationRuleID], a)
    }

    // Create edge chains for each rule
    for ruleID, ruleActions := range actionsByRule {
        // Start edge
        _, err := api.CreateActionEdge(ctx, NewActionEdge{
            RuleID:         ruleID,
            SourceActionID: nil,
            TargetActionID: ruleActions[0].ID,
            EdgeType:       "start",
            EdgeOrder:      0,
        })
        if err != nil {
            return nil, fmt.Errorf("seeding start edge for rule %s: %w", ruleID, err)
        }

        // Sequence edges
        for i := 0; i < len(ruleActions)-1; i++ {
            sourceID := ruleActions[i].ID
            _, err := api.CreateActionEdge(ctx, NewActionEdge{
                RuleID:         ruleID,
                SourceActionID: &sourceID,
                TargetActionID: ruleActions[i+1].ID,
                EdgeType:       "sequence",
                EdgeOrder:      i + 1,
            })
            if err != nil {
                return nil, fmt.Errorf("seeding edge for rule %s action %d: %w", ruleID, i, err)
            }
        }
    }

    return actions, nil
}
```

**Important**: `TestNewRuleActions()` (line 196) distributes actions across rules using `ruleIDs[i%len(ruleIDs)]`. After Phase 2, `ExecutionOrder` is removed from `NewRuleAction`. Verify `TestNewRuleActions` no longer sets `ExecutionOrder` — if it does, remove it here too.

**Validation**:
- [ ] Edge creation added after action creation
- [ ] Actions grouped by rule ID correctly
- [ ] Start edge + sequence edges created for each rule's actions
- [ ] Function still returns `[]RuleAction` (signature unchanged)
- [ ] Code compiles: `go build ./business/sdk/workflow/...`

---

### Task 2: Update `TestSeedFullWorkflow()` in `testutil.go`

**Files**:
- `business/sdk/workflow/testutil.go` - Lines 400-511

**Current behavior**:

`TestSeedFullWorkflow()` calls `TestSeedRuleActions(ctx, 10, ruleIDs, &templateIDs, api)` at line 469 to create 10 actions across 5 rules. After Task 1, `TestSeedRuleActions()` already creates edges internally.

**Implementation**:

**No changes should be needed** if Task 1 is implemented correctly. `TestSeedFullWorkflow()` delegates action creation to `TestSeedRuleActions()`, which now handles edge creation internally.

However, verify:
1. The call at line 469 passes the correct `ruleIDs`
2. The returned `actions` are still stored in `data.RuleActions`
3. No additional action creation happens outside `TestSeedRuleActions()`

**Validation**:
- [ ] Confirm `TestSeedFullWorkflow()` calls `TestSeedRuleActions()` (which now creates edges)
- [ ] No additional action creation outside the helper
- [ ] `TestWorkflowData` struct does not need an `ActionEdges` field (edges are infrastructure, not test data)
- [ ] Code compiles

---

### Task 3: Update `ruleapi/seed_test.go`

**Files**:
- `api/cmd/services/ichor/tests/workflow/ruleapi/seed_test.go` - `insertSeedData()` function

**Current behavior**:

Creates 3 automation rules (line 85) and 5 actions distributed across them (line 96). No edges are created.

```go
// Current: creates actions but no edges
actions, err := workflow.TestSeedRuleActions(ctx, 5, ruleIDs, &templateIDs, busDomain.Workflow)
```

**Implementation**:

After Task 1, `TestSeedRuleActions()` now creates edges internally. **No changes should be needed** to this file.

However, verify:
1. The `TestSeedRuleActions` call signature still matches
2. The test doesn't create additional actions outside the helper
3. The seed data struct doesn't need edge fields

If for some reason this file creates actions outside `TestSeedRuleActions()`, add edge creation following the standard pattern.

**Validation**:
- [ ] Confirm seed function uses `TestSeedRuleActions()` (which now creates edges)
- [ ] No orphaned actions created outside the helper
- [ ] Code compiles

---

### Task 4: Update `cascade_seed_test.go`

**Files**:
- `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_seed_test.go` - `insertCascadeSeedData()` function (lines 58-387)

**Current behavior**:

This is the most complex seed function. It creates 7 rules with varying numbers of actions:
- **PrimaryRule** (line 159): actions created inline
- **DownstreamTriggerRules** x2 (lines 182-213): actions created inline
- **NonModifyingRule** (line 227): actions created inline
- **MixedActionsRule** (line 258): actions created inline
- **SelfTriggerRule** (line 303): actions created inline
- **InactiveDownstreamRule** (line 334): actions created inline

Total: ~7 actions across 7 rules, **NO edges created anywhere**.

**Implementation**:

This file creates actions directly via `busDomain.Workflow.CreateRuleAction()` (NOT through `TestSeedRuleActions()`). After each action creation, add edge creation.

For each rule that has actions, add a simple start edge (since most rules have 1 action):

```go
// After creating an action for a rule:
action, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
    AutomationRuleID: rule.ID,
    Name:             "Some Action",
    ActionConfig:     actionConfig,
    IsActive:         true,
    TemplateID:       &template.ID,
})
if err != nil {
    return CascadeSeedData{}, fmt.Errorf("creating action: %w", err)
}

// Add start edge for single-action rule
_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
    RuleID:         rule.ID,
    SourceActionID: nil,
    TargetActionID: action.ID,
    EdgeType:       "start",
    EdgeOrder:      0,
})
if err != nil {
    return CascadeSeedData{}, fmt.Errorf("creating edge for action: %w", err)
}
```

For rules with multiple actions, create a full chain (Start -> Action0 -> Action1 -> ...).

**Important**: Also remove any `ExecutionOrder` fields from `NewRuleAction` structs if still present after Phase 2.

**Validation**:
- [ ] Every rule that has actions also has edges
- [ ] Single-action rules get a start edge
- [ ] Multi-action rules get start + sequence edges
- [ ] No `ExecutionOrder` references remain
- [ ] Code compiles

---

### Task 5: Update `executionapi/seed_test.go`

**Files**:
- `api/cmd/services/ichor/tests/workflow/executionapi/seed_test.go` - `insertSeedData()` function

**Current behavior**:

Creates 2 automation rules (line 78) and 3 actions (line 89). No edges created.

**Implementation**:

Check whether this file uses `TestSeedRuleActions()` or creates actions directly:

- **If it uses `TestSeedRuleActions()`**: No changes needed (Task 1 handles it)
- **If it creates actions directly**: Add edge creation after each action, following the standard pattern

For 3 actions across 2 rules, expect:
- Rule 1 might get 2 actions: Start -> Action0 -> Action1
- Rule 2 might get 1 action: Start -> Action0

(Actual distribution depends on `TestNewRuleActions` modulo logic)

**Validation**:
- [ ] All actions have corresponding edges
- [ ] No `ExecutionOrder` references remain
- [ ] Code compiles

---

### Task 6: Update `ordersapi/workflow_test.go`

**Files**:
- `api/cmd/services/ichor/tests/sales/ordersapi/workflow_test.go` - Lines 87-177

**Current behavior**:

Creates 3 rules inline (on_create, on_update, on_delete) with 1 action each, directly using `busDomain.Workflow.CreateRuleAction()`. **No edges created.**

**Implementation**:

After each `CreateRuleAction()` call, add a start edge:

```go
// After creating the action:
action, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
    AutomationRuleID: rule.ID,
    Name:             "...",
    ActionConfig:     actionConfig,
    IsActive:         true,
    TemplateID:       &template.ID,
})
if err != nil {
    t.Fatalf("Failed to create rule action: %v", err)
}

// Add start edge (single action per rule)
_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
    RuleID:         rule.ID,
    SourceActionID: nil,
    TargetActionID: action.ID,
    EdgeType:       "start",
    EdgeOrder:      0,
})
if err != nil {
    t.Fatalf("Failed to create edge: %v", err)
}
```

Repeat for all 3 rules. Also remove any `ExecutionOrder` from `NewRuleAction` structs.

**Validation**:
- [ ] All 3 rules have start edges for their actions
- [ ] No `ExecutionOrder` references
- [ ] Code compiles

---

### Task 7: Update `formdataapi/workflow_test.go`

**Files**:
- `api/cmd/services/ichor/tests/formdata/formdataapi/workflow_test.go` - Lines 49-198

**Current behavior**:

Creates 2 rules inline with 1 action each, directly using `busDomain.Workflow.CreateRuleAction()`. **No edges created.**

**Implementation**:

Same pattern as Task 6. After each `CreateRuleAction()` call, add a start edge:

```go
_, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
    RuleID:         rule.ID,
    SourceActionID: nil,
    TargetActionID: action.ID,
    EdgeType:       "start",
    EdgeOrder:      0,
})
if err != nil {
    t.Fatalf("Failed to create edge: %v", err)
}
```

Repeat for both rules. Also remove any `ExecutionOrder` from `NewRuleAction` structs.

**Validation**:
- [ ] Both rules have start edges for their actions
- [ ] No `ExecutionOrder` references
- [ ] Code compiles

---

### Task 9: Update `seedFrontend.go` Workflow Seeding

**Files**:
- `business/sdk/dbtest/seedFrontend.go` - Lines 3768-3939

**Current behavior**:

Creates 3 automation rules with 1 action each using direct `CreateRuleAction()` calls (NOT through `TestSeedRuleActions()`). No edges are created. Error handling uses `log.Error()` (non-fatal, seeding continues).

Rules:
- "Line Item Created - Allocate Inventory" (1 action: Allocate Inventory for Line Item)
- "Allocation Success - Update Line Items" (1 action: Update Line Items to ALLOCATED)
- "Allocation Failed - Alert Operations" (1 action: Create Alert for Operations)

**Implementation**:

After each `CreateRuleAction()` call succeeds, add a start edge. Since each rule has exactly 1 action, only a start edge is needed:

```go
allocateAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
    AutomationRuleID: rule.ID,
    // ...
})
if err != nil {
    log.Error(ctx, "Failed to create action", "error", err)
} else {
    _, err = busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
        RuleID:         rule.ID,
        SourceActionID: nil,
        TargetActionID: allocateAction.ID,
        EdgeType:       "start",
        EdgeOrder:      0,
    })
    if err != nil {
        log.Error(ctx, "Failed to create edge for action", "error", err)
    }
    log.Info(ctx, "✅ Created rule")
}
```

Key changes:
- Change `_, err =` to `action, err :=` to capture the created action
- Add `CreateActionEdge()` call in the `else` block before the success log
- Use `log.Error()` for edge creation failures (consistent with file pattern)

**Validation**:
- [ ] All 3 rules have start edges for their single action
- [ ] Error handling follows existing pattern (`log.Error`, non-fatal)
- [ ] Code compiles: `go build ./business/sdk/dbtest/...`

---

### Task 8: Verify workflowsaveapi Test Files (No Changes Expected)

**Files**:
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go`
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go`
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go`

**Current behavior**:

All three files **already include edges** in their inline `SaveWorkflowRequest` structs. No changes needed.

**Implementation**:

Verify by inspection that:
1. Every `SaveWorkflowRequest` with actions also has edges
2. No `ExecutionOrder` references remain (Phase 2 should have caught these)

```bash
# Quick verification
grep -n "ExecutionOrder" api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go
grep -n "ExecutionOrder" api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go
grep -n "ExecutionOrder" api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go
```

**Validation**:
- [ ] Confirmed: all three files already have edges
- [ ] No remaining `ExecutionOrder` references
- [ ] No changes needed

---

## Deliverables

- [ ] Updated `TestSeedRuleActions()` with edge creation after actions (testutil.go)
- [ ] Verified `TestSeedFullWorkflow()` works via updated `TestSeedRuleActions()` (testutil.go)
- [ ] Verified `ruleapi/seed_test.go` works via updated `TestSeedRuleActions()` (or added edges if direct creation)
- [ ] Updated `cascade_seed_test.go` with edge creation for all 7 rules (~7 actions)
- [ ] Verified or updated `executionapi/seed_test.go` with edge creation for 3 actions
- [ ] Updated `ordersapi/workflow_test.go` with edge creation for 3 inline rules
- [ ] Updated `formdataapi/workflow_test.go` with edge creation for 2 inline rules
- [ ] Updated `seedFrontend.go` with edge creation for 3 inline rules (1 action each)
- [ ] Verified workflowsaveapi test files already have edges (no changes)
- [ ] No remaining `ExecutionOrder` references in any seed/test file

---

## Validation Criteria

- [ ] Go compilation passes: `go build ./...`
- [ ] All workflow tests pass: `go test ./business/sdk/workflow/...`
- [ ] All ruleapi tests pass: `go test ./api/cmd/services/ichor/tests/workflow/ruleapi/...`
- [ ] All executionapi tests pass: `go test ./api/cmd/services/ichor/tests/workflow/executionapi/...`
- [ ] All workflowsaveapi tests pass: `go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...`
- [ ] All ordersapi tests pass: `go test ./api/cmd/services/ichor/tests/sales/ordersapi/...`
- [ ] All formdataapi tests pass: `go test ./api/cmd/services/ichor/tests/formdata/formdataapi/...`
- [ ] No seed function creates actions without corresponding edges (including seedFrontend.go)
- [ ] No remaining `ExecutionOrder` references in seed/test files
- [ ] `make seed` works correctly (if applicable)
- [ ] Full test suite: `make test`

---

## Testing Strategy

### What to Test

- **Seed function completeness**: Every `CreateRuleAction()` call has a corresponding edge chain
- **Edge chain validity**: Start edge has `SourceActionID: nil`, sequence edges have correct source/target
- **Rule isolation**: Edges use the correct `RuleID` for each rule (not mixed up)
- **Single-action rules**: Only need a start edge
- **Multi-action rules**: Need start + (N-1) sequence edges
- **Test independence**: Changes to shared helpers (`TestSeedRuleActions`) don't break callers

### How to Test

```bash
# Compile check
go build ./...

# Run tests for each affected area
go test -v ./business/sdk/workflow/ -run TestSeed
go test -v ./api/cmd/services/ichor/tests/workflow/ruleapi/...
go test -v ./api/cmd/services/ichor/tests/workflow/executionapi/...
go test -v ./api/cmd/services/ichor/tests/sales/ordersapi/...
go test -v ./api/cmd/services/ichor/tests/formdata/formdataapi/...
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...

# Full test suite
make test
```

---

## Gotchas and Tips

- **`TestSeedRuleActions` distributes actions via modulo**: `TestNewRuleActions()` assigns actions to rules using `ruleIDs[i%len(ruleIDs)]`. When grouping by rule ID to create edges, the order within each group matters — edges should chain actions in the order they were created.
- **Use `RuleID`, not `AutomationRuleID`**: The `NewActionEdge` struct field is `RuleID`. The `RuleAction` struct has `AutomationRuleID`. When creating edges for an action, use the action's `AutomationRuleID` value but put it in the edge's `RuleID` field.
- **No `CreatedBy` on `NewActionEdge`**: The struct fields are: `RuleID`, `SourceActionID`, `TargetActionID`, `EdgeType`, `EdgeOrder`. No `CreatedBy`.
- **Start edge `SourceActionID` is nil**: It's a `*uuid.UUID`. For start edges, simply omit it or set to `nil`.
- **`cascade_seed_test.go` is the hardest task**: It creates actions inline across 7 different rules with different test scenarios. Take care to add edges to each rule individually and preserve the cascade testing logic.
- **Check `TestNewRuleActions` for `ExecutionOrder`**: After Phase 2, this helper (testutil.go line ~196) should no longer set `ExecutionOrder`. Verify this is the case.
- **Order matters for edge chains**: When actions are grouped by rule, create edges in the same order actions were created. The first action becomes the start edge target, subsequent actions chain via sequence edges.
- **Don't add edges for rules with zero actions**: Some test scenarios may intentionally create rules without actions (draft/trigger-only rules). Only add edges when actions exist.
- **`seedEdgesForRule` helper exists**: If you find yourself writing the same edge creation loop repeatedly, consider importing or adapting the `seedEdgesForRule` helper from `workflowsaveapi/seed_test.go` (lines 151-191). However, this helper is in a test package and may not be importable — in that case, inline the pattern.

---

## Reference

- Original plan: `.claude/plans/missing-action-edges.md`
- Progress tracking: `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml`
- Phase 1 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_1_VALIDATION_LAYER.md`
- Phase 2 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_2_REMOVE_EXECUTION_ORDER.md`
- Phase 3 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_3_REMOVE_LINEAR_EXECUTOR.md`
- Phase 4 (dependency): `.claude/plans/MISSING_ACTION_EDGES_PLAN/phases/PHASE_4_TEST_UPDATES.md`
- `NewActionEdge` struct: `business/sdk/workflow/models.go` lines 408-414
- `CreateActionEdge` method: `business/sdk/workflow/workflowbus.go` lines 1170-1180
- `seedEdgesForRule` helper: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/seed_test.go` lines 151-191
- `TestNewRuleActions` helper: `business/sdk/workflow/testutil.go` lines ~196-226
- `TestSeedRuleActions` helper: `business/sdk/workflow/testutil.go` lines 228-243
- `TestSeedFullWorkflow` helper: `business/sdk/workflow/testutil.go` lines 400-511

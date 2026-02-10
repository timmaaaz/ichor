# Phase 5: Workflow Implementation

**Category**: backend
**Status**: Pending
**Dependencies**: Phase 3 (Core Models & Context - COMPLETED, amended with `ContinuationState`), Phase 4 (Graph Executor - COMPLETED, provides `GraphExecutor`, `GetStartActions`, `GetNextActions`, `FindConvergencePoint`)

---

## Overview

Implement the main Temporal workflow function (`ExecuteGraphWorkflow`) that orchestrates the execution of visual workflow graphs. This is the bridge between the pure graph traversal logic (Phase 4) and the activity execution layer (Phase 6). The workflow handles sequential execution, parallel branching with convergence detection, fire-and-forget branches, Continue-As-New for long-running workflows, versioning for safe in-flight upgrades, and action-type-aware timeout configuration.

This is the first file that directly imports the Temporal SDK (`go.temporal.io/sdk/workflow` and `go.temporal.io/sdk/temporal`). All operations within workflow code must be deterministic - no `time.Now()`, no `rand`, no direct I/O, no goroutines (use `workflow.Go` instead).

## Goals

1. **Implement the main Temporal workflow function (`ExecuteGraphWorkflow`)** that orchestrates sequential and parallel action execution using the `GraphExecutor`, with input validation and Continue-As-New at 10K events to prevent unbounded history growth
2. **Build parallel branch execution with convergence detection** - child workflows per branch, selector-based wait-for-all, per-action-name result merging at convergence point, and fire-and-forget support using child workflows with `PARENT_CLOSE_POLICY_ABANDON`
3. **Add workflow versioning for safe in-flight upgrades and action-type-aware timeout configuration** - `workflow.GetVersion` for breaking changes, async actions get 30min timeout, human-in-the-loop actions get 7-day timeout

## Prerequisites

- Phase 3 complete (amended): `WorkflowInput` (with `ContinuationState *MergedContext`), `MergedContext`, `BranchInput`/`BranchOutput`, `ActionActivityInput`/`ActionActivityOutput`, constants (`TaskQueue`, `HistoryLengthThreshold`, etc.)
- Phase 4 complete: `GraphExecutor` with `GetStartActions`, `GetNextActions`, `FindConvergencePoint`, `Graph`
- Temporal SDK available in `go.mod`/`vendor` (from Phase 1)
- Understanding of Temporal determinism constraints

---

## Go Package Structure

```
business/sdk/workflow/temporal/
    models.go              <- Phase 3 (COMPLETED, amended with ContinuationState)
    models_test.go         <- Phase 3 (COMPLETED)
    graph_executor.go      <- Phase 4 (COMPLETED)
    graph_executor_test.go <- Phase 4 (COMPLETED)
    workflow.go            <- THIS PHASE
    activities_stub.go     <- THIS PHASE (temporary, replaced by Phase 6)
    activities.go          <- Phase 6
    activities_async.go    <- Phase 6
    async_completer.go     <- Phase 6
    trigger.go             <- Phase 7
```

---

## Task Breakdown

### Task 1: Implement workflow.go Core Functions

**Status**: Pending

**Description**: Implement the main workflow entry point and its core execution functions. `ExecuteGraphWorkflow` is registered with Temporal and called when a workflow starts. It validates input, restores continuation state if present, initializes the execution context, builds the graph executor, finds start actions, and kicks off execution. `checkContinueAsNew` prevents unbounded history growth while preserving the full `MergedContext` structure. `executeActions` dispatches between sequential and parallel execution. `executeSingleAction` executes a single action as a Temporal activity with contextual error wrapping and recurses.

**Notes**:
- `ExecuteGraphWorkflow` - Main entry registered with Temporal worker; calls `input.Validate()` first
- `checkContinueAsNew` - Preserves full `MergedContext` via `input.ContinuationState` (not flattened into TriggerData)
- `executeActions` - Dispatcher: 0 actions = done, 1 action = sequential, >1 = parallel (check convergence)
- `executeSingleAction` - Executes activity, merges result, gets next actions, recurses
- `activityOptions` - Helper to build `workflow.ActivityOptions` based on action type (avoids duplication with branch child workflow)

**Files**:
- `business/sdk/workflow/temporal/workflow.go`

**Implementation Guide**:

```go
package temporal

import (
    "fmt"
    "time"

    enumspb "go.temporal.io/api/enums/v1"
    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"
)

// Package-level action type classification maps.
// Defined at package scope to avoid per-call allocation.
var (
    asyncActionTypes = map[string]bool{
        "allocate_inventory":   true,
        "send_email":           true,
        "credit_check":         true,
        "fraud_detection":      true,
        "third_party_api_call": true,
        "reserve_shipping":     true,
    }

    humanActionTypes = map[string]bool{
        "manager_approval":   true,
        "manual_review":      true,
        "human_verification": true,
        "approval_request":   true,
    }
)

// ExecuteGraphWorkflow interprets any graph definition dynamically.
// This is the core workflow registered with the Temporal worker.
//
// Determinism requirements:
//   - No time.Now(), rand, or direct I/O
//   - Use workflow.GetLogger, workflow.Go, workflow.ExecuteActivity
//   - All map iterations in GraphExecutor are pre-sorted
func ExecuteGraphWorkflow(ctx workflow.Context, input WorkflowInput) error {
    logger := workflow.GetLogger(ctx)

    // Validate input before proceeding.
    if err := input.Validate(); err != nil {
        logger.Error("Invalid workflow input", "error", err)
        return fmt.Errorf("validate workflow input: %w", err)
    }

    // Version the interpreter logic for safe deployments.
    // Increment maxVersion when making breaking changes to execution logic.
    v := workflow.GetVersion(ctx, "graph-interpreter", workflow.DefaultVersion, 1)

    logger.Info("Starting graph workflow",
        "rule_id", input.RuleID,
        "execution_id", input.ExecutionID,
        "interpreter_version", v,
        "is_continuation", input.ContinuationState != nil,
    )

    // Initialize or restore execution context.
    // ContinuationState is non-nil after Continue-As-New, preserving the full
    // MergedContext (ActionResults + TriggerData + Flattened) across boundaries.
    var mergedCtx *MergedContext
    if input.ContinuationState != nil {
        mergedCtx = input.ContinuationState
    } else {
        mergedCtx = NewMergedContext(input.TriggerData)
    }

    // Build graph executor with pre-sorted edge indexes
    executor := NewGraphExecutor(input.Graph)

    // Find start actions (edges with source_action_id = nil)
    startActions := executor.GetStartActions()
    if len(startActions) == 0 {
        logger.Info("Empty workflow - no start actions found")
        return nil
    }

    // Execute from start (may be multiple parallel start actions)
    return executeActions(ctx, executor, startActions, mergedCtx, input)
}

// checkContinueAsNew returns a ContinueAsNewError if history has grown too large.
// This prevents hitting Temporal's 50K event limit on long-running workflows.
// The full MergedContext is preserved via ContinuationState, maintaining the
// structured ActionResults map and Flattened template resolution data.
func checkContinueAsNew(ctx workflow.Context, input WorkflowInput, mergedCtx *MergedContext) error {
    info := workflow.GetInfo(ctx)
    if info.GetCurrentHistoryLength() > HistoryLengthThreshold {
        logger := workflow.GetLogger(ctx)
        logger.Info("History threshold exceeded, continuing as new workflow",
            "history_length", info.GetCurrentHistoryLength(),
            "threshold", HistoryLengthThreshold,
        )

        // Preserve the full MergedContext structure across Continue-As-New.
        // This maintains ActionResults (structured) and Flattened (template resolution)
        // without losing the distinction between trigger data and action results.
        input.ContinuationState = mergedCtx

        return workflow.NewContinueAsNewError(ctx, ExecuteGraphWorkflow, input)
    }
    return nil
}

// executeActions handles both sequential and parallel execution.
// This is the main dispatcher called recursively as the workflow progresses.
func executeActions(ctx workflow.Context, executor *GraphExecutor, actions []ActionNode, mergedCtx *MergedContext, input WorkflowInput) error {
    if len(actions) == 0 {
        return nil
    }

    // Check if we need to continue-as-new to avoid history size limits
    if err := checkContinueAsNew(ctx, input, mergedCtx); err != nil {
        return err
    }

    if len(actions) == 1 {
        // Sequential execution - single path
        return executeSingleAction(ctx, executor, actions[0], mergedCtx, input)
    }

    // Parallel execution - check for convergence
    convergencePoint := executor.FindConvergencePoint(actions)

    if convergencePoint == nil {
        // Fire-and-forget parallel branches - no convergence
        return executeFireAndForget(ctx, executor, actions, mergedCtx, input)
    }

    // Parallel with convergence - must wait for all branches
    return executeParallelWithConvergence(ctx, executor, actions, convergencePoint, mergedCtx, input)
}

// executeSingleAction executes one action and continues to the next.
// This is the workhorse function - it executes a Temporal activity,
// merges the result into context, and recurses with the next actions.
func executeSingleAction(ctx workflow.Context, executor *GraphExecutor, action ActionNode, mergedCtx *MergedContext, input WorkflowInput) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Executing action",
        "action_id", action.ID,
        "action_name", action.Name,
        "action_type", action.ActionType,
    )

    // Prepare activity input
    activityInput := ActionActivityInput{
        ActionID:   action.ID,
        ActionName: action.Name,
        ActionType: action.ActionType,
        Config:     action.Config,
        Context:    mergedCtx.Flattened,
    }

    // Configure activity options based on action type
    activityCtx := workflow.WithActivityOptions(ctx, activityOptions(action.ActionType))

    // Execute the action as a Temporal activity
    // ExecuteActionActivity is stubbed in activities_stub.go until Phase 6
    var result ActionActivityOutput
    if err := workflow.ExecuteActivity(activityCtx, ExecuteActionActivity, activityInput).Get(ctx, &result); err != nil {
        logger.Error("Action failed",
            "action_id", action.ID,
            "action_name", action.Name,
            "error", err,
        )
        return fmt.Errorf("execute action %s (%s): %w", action.Name, action.ID, err)
    }

    // Merge result into context for subsequent actions
    mergedCtx.MergeResult(action.Name, result.Result)

    logger.Info("Action completed",
        "action_id", action.ID,
        "action_name", action.Name,
        "success", result.Success,
    )

    // Get next actions based on result and edge types
    nextActions := executor.GetNextActions(action.ID, result.Result)

    if len(nextActions) == 0 {
        // End of this path
        return nil
    }

    // Continue execution (recursive)
    return executeActions(ctx, executor, nextActions, mergedCtx, input)
}

// activityOptions builds ActivityOptions based on the action type.
// This avoids duplicating timeout logic between executeSingleAction and
// ExecuteBranchUntilConvergence.
func activityOptions(actionType string) workflow.ActivityOptions {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    3,
        },
    }

    // Async actions (external APIs, email, inventory) get longer timeouts
    if isAsyncAction(actionType) {
        ao.StartToCloseTimeout = 30 * time.Minute
        ao.HeartbeatTimeout = time.Minute
    }

    // Human-in-the-loop actions (approvals, manual review) can take days
    if isHumanAction(actionType) {
        ao.StartToCloseTimeout = 7 * 24 * time.Hour // 7 days
        ao.HeartbeatTimeout = time.Hour
    }

    return ao
}
```

---

### Task 2: Implement Parallel Execution Functions

**Status**: Pending

**Description**: Implement the two parallel execution patterns: fire-and-forget (branches run as child workflows with `PARENT_CLOSE_POLICY_ABANDON`, parent doesn't wait) and parallel-with-convergence (branches run as child workflows, parent waits for all at convergence point, then merges results by individual action name and continues).

**Notes**:
- `executeFireAndForget` - Uses `workflow.ExecuteChildWorkflow` with `PARENT_CLOSE_POLICY_ABANDON` so branches survive parent completion. Uses `workflow.Go` only to consume the child future without blocking.
- `executeParallelWithConvergence` - Uses `workflow.ExecuteChildWorkflow` per branch, `workflow.NewSelector` to wait for all, merges results per action name into shared context at convergence
- `ExecuteBranchUntilConvergence` - Child workflow function, executes actions linearly until reaching convergence point ID. Uses `activityOptions()` helper to avoid duplicating timeout logic.

**Result Merging Strategy**: When branches converge, each branch's `BranchOutput.ActionResults` (keyed by action name) is iterated and merged individually via `mergedCtx.MergeResult(actionName, actionResult)`. This means the convergence point action and all subsequent actions can access any upstream action's result by name (e.g., `{{branch_left_action.field}}`). If two branches happen to execute actions with the same name, the later merge overwrites the earlier - action names should be unique within a graph.

**Files**:
- `business/sdk/workflow/temporal/workflow.go` (same file as Task 1)

**Implementation Guide**:

```go
// executeFireAndForget launches parallel branches as child workflows that
// survive the parent's completion. Uses PARENT_CLOSE_POLICY_ABANDON so
// branches continue running independently even after the parent workflow ends.
//
// Branch errors do not fail the parent workflow - they are logged by the
// child workflow itself and visible in the Temporal UI.
func executeFireAndForget(ctx workflow.Context, executor *GraphExecutor, branches []ActionNode, mergedCtx *MergedContext, input WorkflowInput) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Executing fire-and-forget parallel branches",
        "branch_count", len(branches),
    )

    for i, branch := range branches {
        branchAction := branch
        branchIndex := i

        // Each fire-and-forget branch runs as a child workflow with ABANDON policy.
        // This means the child continues even if the parent completes or fails.
        childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
            WorkflowID: fmt.Sprintf("%s-fire-forget-%d-%s",
                workflow.GetInfo(ctx).WorkflowExecution.ID,
                branchIndex,
                branchAction.ID,
            ),
            ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
        })

        // ExecuteBranchUntilConvergence handles the fire-and-forget case when
        // ConvergencePoint is uuid.Nil: it executes until no more next actions.
        future := workflow.ExecuteChildWorkflow(childCtx, ExecuteBranchUntilConvergence,
            BranchInput{
                StartAction:      branchAction,
                ConvergencePoint: uuid.Nil, // No convergence - run until end of path
                Graph:            executor.Graph(),
                InitialContext:   mergedCtx.Clone(),
            },
        )

        // Consume the future in a goroutine to avoid blocking the parent.
        // We don't wait for completion - the child is abandoned on parent close.
        workflow.Go(ctx, func(gCtx workflow.Context) {
            var output BranchOutput
            if err := future.Get(gCtx, &output); err != nil {
                logger.Warn("Fire-and-forget branch failed",
                    "branch_index", branchIndex,
                    "start_action", branchAction.Name,
                    "error", err,
                )
            }
        })
    }

    // Parent returns immediately - fire-and-forget branches run independently
    return nil
}

// executeParallelWithConvergence executes branches as child workflows
// and waits for all of them at the convergence point.
// After all branches complete, their results are merged into the shared
// context (keyed by individual action name) and execution continues from
// the convergence point.
func executeParallelWithConvergence(
    ctx workflow.Context,
    executor *GraphExecutor,
    branches []ActionNode,
    convergencePoint *ActionNode,
    mergedCtx *MergedContext,
    input WorkflowInput,
) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Executing parallel branches with convergence",
        "branch_count", len(branches),
        "convergence_point", convergencePoint.Name,
    )

    // Create selector for waiting on all branches
    selector := workflow.NewSelector(ctx)
    branchResults := make([]BranchOutput, len(branches))
    branchErrors := make([]error, len(branches))

    for i, branch := range branches {
        branchIndex := i
        branchAction := branch

        // Each branch runs as a child workflow
        childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
            WorkflowID: fmt.Sprintf("%s-branch-%d-%s",
                workflow.GetInfo(ctx).WorkflowExecution.ID,
                branchIndex,
                branchAction.ID,
            ),
        })

        future := workflow.ExecuteChildWorkflow(childCtx, ExecuteBranchUntilConvergence,
            BranchInput{
                StartAction:      branchAction,
                ConvergencePoint: convergencePoint.ID,
                Graph:            executor.Graph(),
                InitialContext:   mergedCtx,
            },
        )

        selector.AddFuture(future, func(f workflow.Future) {
            var output BranchOutput
            branchErrors[branchIndex] = f.Get(ctx, &output)
            branchResults[branchIndex] = output
        })
    }

    // Wait for all branches
    for i := 0; i < len(branches); i++ {
        selector.Select(ctx)
    }

    // Check for errors - any branch failure fails the entire parallel group
    for i, err := range branchErrors {
        if err != nil {
            logger.Error("Branch failed",
                "branch_index", i,
                "error", err,
            )
            return fmt.Errorf("branch %d (%s) failed: %w", i, branches[i].Name, err)
        }
    }

    // Merge all branch results into context.
    // Results are keyed by individual action name, not by branch index.
    // This allows the convergence point and subsequent actions to access
    // any upstream action's result by name (e.g., {{send_email.status}}).
    for _, br := range branchResults {
        for actionName, actionResult := range br.ActionResults {
            mergedCtx.MergeResult(actionName, actionResult)
        }
    }

    logger.Info("All branches converged, continuing from convergence point",
        "convergence_point", convergencePoint.Name,
    )

    // Continue from convergence point
    return executeSingleAction(ctx, executor, *convergencePoint, mergedCtx, input)
}

// ExecuteBranchUntilConvergence executes actions until reaching the convergence point.
// This is a child workflow function - one instance runs per parallel branch.
// It clones the initial context to prevent cross-branch data leakage.
//
// When ConvergencePoint is uuid.Nil (fire-and-forget), the branch executes
// until there are no more next actions (end of path).
//
// Branch linearity assumption: Within a branch leading to a convergence point,
// each action should resolve to a single next action. Conditional edges
// (true_branch/false_branch) are resolved by GetNextActions based on the
// action result, so even conditionals yield a single next action. If multiple
// next actions are returned (nested parallelism), only the first is followed -
// nested parallel branches are a Phase 10+ concern.
func ExecuteBranchUntilConvergence(ctx workflow.Context, input BranchInput) (BranchOutput, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting branch execution",
        "start_action", input.StartAction.Name,
        "convergence_point", input.ConvergencePoint,
    )

    executor := NewGraphExecutor(input.Graph)
    mergedCtx := input.InitialContext.Clone()

    currentAction := input.StartAction

    for {
        // Check if we've reached convergence point.
        // For fire-and-forget branches (ConvergencePoint == uuid.Nil), this
        // check never triggers - the branch runs until no more next actions.
        if currentAction.ID == input.ConvergencePoint {
            logger.Info("Branch reached convergence point")
            break
        }

        // Execute action
        activityInput := ActionActivityInput{
            ActionID:   currentAction.ID,
            ActionName: currentAction.Name,
            ActionType: currentAction.ActionType,
            Config:     currentAction.Config,
            Context:    mergedCtx.Flattened,
        }

        activityCtx := workflow.WithActivityOptions(ctx, activityOptions(currentAction.ActionType))

        var result ActionActivityOutput
        if err := workflow.ExecuteActivity(activityCtx, ExecuteActionActivity, activityInput).Get(ctx, &result); err != nil {
            return BranchOutput{}, fmt.Errorf("execute action %s (%s): %w", currentAction.Name, currentAction.ID, err)
        }

        mergedCtx.MergeResult(currentAction.Name, result.Result)

        // Get next action. Conditional edges are resolved by GetNextActions based
        // on result, so even conditionals should yield a single next action.
        nextActions := executor.GetNextActions(currentAction.ID, result.Result)
        if len(nextActions) == 0 {
            // End of branch path (fire-and-forget) or dead end before convergence
            if input.ConvergencePoint != uuid.Nil {
                logger.Warn("Branch ended before reaching convergence point",
                    "last_action", currentAction.Name,
                    "convergence_point", input.ConvergencePoint,
                )
            }
            break
        }

        // Follow first next action. Multiple next actions within a branch
        // (nested parallelism) is not supported in this phase.
        if len(nextActions) > 1 {
            logger.Warn("Multiple next actions in branch - following first only",
                "action", currentAction.Name,
                "next_count", len(nextActions),
            )
        }
        currentAction = nextActions[0]
    }

    return BranchOutput{
        ActionResults: mergedCtx.ActionResults,
    }, nil
}
```

---

### Task 3: Implement Workflow Versioning

**Status**: Pending

**Description**: Add Temporal workflow versioning using `workflow.GetVersion` to support safe in-flight workflow upgrades. When the interpreter logic changes, incrementing the max version allows new workflows to use new logic while replaying in-flight workflows with old logic.

**Notes**:
- `workflow.GetVersion(ctx, "graph-interpreter", workflow.DefaultVersion, 1)` in `ExecuteGraphWorkflow`
- Version switch placeholder for future logic changes
- This is integrated into the `ExecuteGraphWorkflow` function (Task 1), not a separate function

**Files**:
- `business/sdk/workflow/temporal/workflow.go` (same file)

**Implementation Guide**:

Versioning is already included in the Task 1 implementation via:

```go
v := workflow.GetVersion(ctx, "graph-interpreter", workflow.DefaultVersion, 1)
```

When making future breaking changes, add a version switch:

```go
v := workflow.GetVersion(ctx, "graph-interpreter", workflow.DefaultVersion, 2)

switch v {
case workflow.DefaultVersion:
    // Original logic (for replaying old workflows)
    // ...
case 1:
    // First change
    // ...
case 2:
    // Current logic
    // ...
}
```

For Phase 5, no version switch is needed yet - just the initial version marker.

---

### Task 4: Add Action Type Helpers

**Status**: Pending

**Description**: Implement helper functions that classify action types for timeout configuration. These use package-level map variables (not per-call allocation) for efficiency. Async actions (external APIs, email sending) get longer timeouts. Human actions (approvals, manual review) get multi-day timeouts.

**Notes**:
- `asyncActionTypes` and `humanActionTypes` defined as package-level `var` declarations
- `isAsyncAction(actionType string) bool` - Simple map lookup
- `isHumanAction(actionType string) bool` - Simple map lookup
- `activityOptions(actionType string) workflow.ActivityOptions` - Shared helper used by both `executeSingleAction` and `ExecuteBranchUntilConvergence`

**Files**:
- `business/sdk/workflow/temporal/workflow.go` (same file)

**Implementation Guide**:

```go
// Package-level maps are defined in the Task 1 implementation guide (top of file).

// isAsyncAction returns true for actions that involve external async operations.
// These get longer timeouts (30min) and heartbeat requirements.
func isAsyncAction(actionType string) bool {
    return asyncActionTypes[actionType]
}

// isHumanAction returns true for actions that require human interaction.
// These get multi-day timeouts (7 days) to allow for review cycles.
func isHumanAction(actionType string) bool {
    return humanActionTypes[actionType]
}

// activityOptions is defined in the Task 1 implementation guide.
```

---

### Task 5: Create Activity Stub File

**Status**: Pending

**Description**: Create `activities_stub.go` with a minimal `ExecuteActionActivity` stub so that Phase 5 compiles. This file is deleted entirely when Phase 6 implements the real activity.

**Files**:
- `business/sdk/workflow/temporal/activities_stub.go`

**Implementation Guide**:

```go
package temporal

import (
    "context"
    "fmt"
)

// ExecuteActionActivity is the Temporal activity that executes a single action.
// This is a stub for Phase 5 compilation. Full implementation in Phase 6 (activities.go).
func ExecuteActionActivity(_ context.Context, _ ActionActivityInput) (ActionActivityOutput, error) {
    return ActionActivityOutput{}, fmt.Errorf("not implemented: ExecuteActionActivity requires Phase 6")
}
```

---

## Validation Criteria

- [ ] `go build ./business/sdk/workflow/temporal/...` passes (with `activities_stub.go`)
- [ ] `go test ./business/sdk/workflow/temporal/...` passes (existing Phase 3+4 tests still green)
- [ ] No non-deterministic operations in workflow code:
  - [ ] No `time.Now()` (use `workflow.Now(ctx)` if needed)
  - [ ] No `rand` (use `workflow.SideEffect` if needed)
  - [ ] No direct I/O or goroutines (use `workflow.Go`)
  - [ ] No `map` iteration without sorted keys in workflow functions
- [ ] `input.Validate()` called at start of `ExecuteGraphWorkflow`
- [ ] Continue-As-New preserves full `MergedContext` via `input.ContinuationState` (not flattened)
- [ ] Continue-As-New configured at 10K events (`HistoryLengthThreshold`)
- [ ] `ExecuteGraphWorkflow` is exported (registered with Temporal worker in Phase 9)
- [ ] `ExecuteBranchUntilConvergence` is exported (used as child workflow)
- [ ] `executeActions`, `executeSingleAction`, `executeFireAndForget`, `executeParallelWithConvergence` are unexported (internal dispatch)
- [ ] `isAsyncAction` and `isHumanAction` use package-level maps (not per-call allocation)
- [ ] `activityOptions` helper shared between `executeSingleAction` and `ExecuteBranchUntilConvergence`
- [ ] Parallel branches with convergence use `workflow.ExecuteChildWorkflow` (not `workflow.Go`)
- [ ] Fire-and-forget branches use `workflow.ExecuteChildWorkflow` with `PARENT_CLOSE_POLICY_ABANDON`
- [ ] Child workflow IDs are deterministic (derived from parent workflow ID + branch index + action ID)
- [ ] Branch result merging iterates `BranchOutput.ActionResults` by individual action name
- [ ] `ExecuteBranchUntilConvergence` handles `ConvergencePoint == uuid.Nil` for fire-and-forget
- [ ] All `return err` statements include contextual wrapping via `fmt.Errorf`
- [ ] Activity stub in separate `activities_stub.go` file (not workflow.go)
- [ ] `go vet ./business/sdk/workflow/temporal/...` passes

---

## Deliverables

- `business/sdk/workflow/temporal/workflow.go` - Main workflow implementation
- `business/sdk/workflow/temporal/activities_stub.go` - Temporary stub (replaced in Phase 6)

---

## Gotchas & Tips

### Common Pitfalls

- **`ExecuteActionActivity` stub in `activities_stub.go`**: This function is implemented in Phase 6 (`activities.go`). The stub lives in a separate file (`activities_stub.go`) so Phase 6 simply deletes the stub file and creates `activities.go`. Do NOT put the stub in `workflow.go`.

- **Temporal determinism violations**: The most common mistakes in workflow code:
  - Using `time.Now()` instead of `workflow.Now(ctx)` (not needed here, but worth noting)
  - Using Go's `go` keyword instead of `workflow.Go(ctx, func(...))`
  - Using `rand` for any purpose
  - Making HTTP calls directly (must be in activities)
  - Iterating maps without sorting (GraphExecutor handles this, but be careful in workflow.go)

- **Fire-and-forget uses child workflows, not `workflow.Go`**: Raw `workflow.Go` goroutines don't truly detach - Temporal waits for all goroutines before completing the workflow. Use `workflow.ExecuteChildWorkflow` with `enumspb.PARENT_CLOSE_POLICY_ABANDON` for true fire-and-forget. The `workflow.Go` goroutine in `executeFireAndForget` is only used to consume the child future without blocking the parent.

- **Continue-As-New preserves full MergedContext**: When Continue-As-New fires, the full `MergedContext` is stored in `input.ContinuationState`. This preserves the structured `ActionResults` map (which action produced which output) and the `Flattened` map (for template resolution). On restart, the workflow checks `input.ContinuationState != nil` and restores the full context rather than creating a fresh one from TriggerData. This avoids losing the distinction between trigger data and action results.

- **Selector wait count**: `selector.Select(ctx)` is called `len(branches)` times to wait for ALL branches. If you call it fewer times, some branches won't be waited on. If you call it more times, it will block forever.

- **Child workflow ID determinism**: The child workflow ID format `{parentID}-branch-{index}-{actionID}` (convergence) or `{parentID}-fire-forget-{index}-{actionID}` (fire-and-forget) must be deterministic for replay. Since `branches` comes from `GetNextActions` (which returns sorted results), the index is deterministic. Don't add timestamps or random components.

- **Recursive execution depth**: `executeSingleAction` -> `executeActions` -> `executeSingleAction` is recursive. For linear chains this means one stack frame per action. Continue-As-New fires at 10K history events (~3K-5K actions at ~2-3 events per action), which bounds the recursion to ~5K frames. Go stack starts at 1MB and grows to 1GB, so ~5K frames (~25MB) is well within limits. Not a concern for typical workflows.

- **Branch linearity assumption**: `ExecuteBranchUntilConvergence` expects each action within a branch to resolve to a single next action. Conditional edges (true_branch/false_branch) are evaluated by `GetNextActions` based on the action result, so they yield one next action. If multiple next actions are returned (nested parallelism before convergence), only the first is followed with a warning. Nested parallel branches are a Phase 10+ concern.

- **Payload size risk**: If `mergedCtx` (stored in `ContinuationState` during Continue-As-New) exceeds Temporal's 2MB payload limit, the continue-as-new will fail. Activities must enforce `MaxResultValueSize` truncation (Phase 6). Monitor `ContextSizeWarningBytes` (200KB) in Phase 6 activities and log warnings.

- **No global workflow timeout configured**: If a human approval action never completes, the workflow runs indefinitely (up to the 7-day activity timeout, then retries). Consider adding `WorkflowExecutionTimeout` in Phase 9 when wiring the worker, or rely on Temporal's namespace-level default timeout.

### Tips

- Phase 5 produces two files: `workflow.go` (~400 lines) and `activities_stub.go` (~10 lines)
- Test this code in Phase 11 (Workflow Integration Tests) using Temporal's test suite (`testsuite.WorkflowTestSuite`). Phase 5 itself doesn't include tests because testing workflow code requires the activity implementation from Phase 6
- The `isAsyncAction`/`isHumanAction` maps are intentionally simple. As the action registry grows, these could be moved to configuration. For now, hardcoded maps are fine
- `ExecuteGraphWorkflow` and `ExecuteBranchUntilConvergence` are the only two exported workflow functions. They get registered with the Temporal worker in Phase 9
- The `activityOptions()` helper is extracted to avoid duplicating timeout logic between `executeSingleAction` and `ExecuteBranchUntilConvergence`

---

## Testing Strategy

### Why No Tests in Phase 5

Unlike Phases 3-4, Phase 5 does NOT include unit tests because:

1. **Temporal workflow code requires the Temporal test environment** (`testsuite.WorkflowTestSuite`) to execute
2. **Activities must be mocked or stubbed** - `ExecuteActionActivity` (Phase 6) must exist with real logic
3. **Integration testing is more valuable** than unit testing for workflow orchestration

Testing is deferred to:
- **Phase 11**: Workflow integration tests (simple workflows, sequential actions, template resolution)
- **Phase 11**: Parallel execution tests (convergence, fire-and-forget, nested)
- **Phase 11**: Replay tests (determinism verification)
- **Phase 12**: Continue-As-New tests, error handling tests

### Manual Verification

After writing `workflow.go` and `activities_stub.go`, verify:
1. `go build ./business/sdk/workflow/temporal/...` passes
2. `go test ./business/sdk/workflow/temporal/...` passes (existing tests)
3. `go vet ./business/sdk/workflow/temporal/...` passes
4. No `time.Now()`, `rand`, or bare `go` keywords in workflow.go (search manually)
5. All exported functions have doc comments
6. `input.Validate()` is called before any execution logic

---

## Temporal-Specific Patterns

### Continue-As-New

```
Workflow starts -> executes actions -> history grows
At 10K events -> ContinueAsNewError
New workflow starts with:
  - Same RuleID, ExecutionID
  - Same Graph (unchanged)
  - Same TriggerData (unchanged)
  - ContinuationState = full MergedContext (ActionResults + Flattened)
New workflow detects ContinuationState != nil -> restores full context
Execution continues from current position
```

The full `MergedContext` is preserved via `input.ContinuationState`, maintaining:
- `ActionResults` (structured: which action produced which output)
- `Flattened` (template resolution: `action_name.field` keys)
- `TriggerData` (original event data)

### Parallel Execution: Convergence vs Fire-and-Forget

```
Convergence Pattern:                Fire-and-Forget Pattern:
    A                                   A
   / \                                 / \
  B   C  (child workflows)           B   C  (child workflows + ABANDON)
   \ /                                |
    D  (waits for B & C)              D  (doesn't wait for C)
       (merges by action name)           (C runs independently)
```

**Convergence**: Uses `workflow.ExecuteChildWorkflow` with default parent close policy because the parent needs to:
- Wait for all branches to complete (via `workflow.NewSelector`)
- Collect `BranchOutput.ActionResults` from each branch
- Merge results by individual action name into shared context
- Continue execution from the convergence point

**Fire-and-forget**: Uses `workflow.ExecuteChildWorkflow` with `PARENT_CLOSE_POLICY_ABANDON` because:
- Branches must survive parent completion
- Parent doesn't need results
- Branches run independently with their own cloned context
- `workflow.Go` only used to consume the child future without blocking

### Workflow Versioning

```go
v := workflow.GetVersion(ctx, "graph-interpreter", workflow.DefaultVersion, 1)
```

When replaying an existing workflow:
- `v` equals the version recorded when the workflow first ran
- New workflows get `maxVersion` (currently 1)
- Old workflows get `workflow.DefaultVersion` (0)

To make a breaking change to execution logic:
1. Increment max version: `workflow.DefaultVersion, 2`
2. Add a switch statement for old vs new logic
3. Deploy new worker code
4. Old in-flight workflows replay with old logic
5. New workflows use new logic

### Activity Options by Action Type

| Action Category | StartToCloseTimeout | HeartbeatTimeout | MaxAttempts | Examples |
|----------------|--------------------|--------------------|-------------|----------|
| Default | 5 minutes | None | 3 | `set_field`, `create_alert`, `evaluate_condition` |
| Async | 30 minutes | 1 minute | 3 | `send_email`, `allocate_inventory`, `credit_check` |
| Human | 7 days | 1 hour | 3 | `manager_approval`, `manual_review` |

---

## Existing Code Reference

### Workflow SDK Imports
```go
import (
    enumspb "go.temporal.io/api/enums/v1"  // PARENT_CLOSE_POLICY_ABANDON
    "go.temporal.io/sdk/temporal"            // RetryPolicy
    "go.temporal.io/sdk/workflow"            // Context, ExecuteActivity, Go, Selector, etc.
)
```

### Key Temporal SDK Functions Used
- `workflow.GetLogger(ctx)` - Deterministic logger
- `workflow.GetInfo(ctx)` - Workflow metadata (history length, execution ID)
- `workflow.GetVersion(ctx, changeID, minVersion, maxVersion)` - Versioning
- `workflow.WithActivityOptions(ctx, ao)` - Activity configuration
- `workflow.ExecuteActivity(ctx, fn, input)` - Execute activity
- `workflow.WithChildOptions(ctx, opts)` - Child workflow configuration
- `workflow.ExecuteChildWorkflow(ctx, fn, input)` - Execute child workflow
- `workflow.NewSelector(ctx)` - Create selector for waiting on futures
- `workflow.Go(ctx, fn)` - Launch concurrent goroutine (deterministic)
- `workflow.NewContinueAsNewError(ctx, fn, input)` - Continue-As-New
- `enumspb.PARENT_CLOSE_POLICY_ABANDON` - Child survives parent completion

### Phase 3 Amendment: ContinuationState

`WorkflowInput` in `models.go` was amended to add:
```go
ContinuationState *MergedContext `json:"continuation_state,omitempty"`
```

This field is:
- `nil` on initial workflow execution
- Set to the full `MergedContext` on Continue-As-New
- Checked in `ExecuteGraphWorkflow` to restore or create fresh context

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 5

# Review plan before implementing
/workflow-temporal-plan-review 5

# Review code after implementing
/workflow-temporal-review 5
```

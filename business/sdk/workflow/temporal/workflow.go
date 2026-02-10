package temporal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Package-level action type classification maps.
// Defined at package scope to avoid per-call allocation.
//
// All actions run through ExecuteActionActivity (synchronous Temporal activity).
// Temporal handles retries, timeouts, and failure recovery natively.
// These maps only affect timeout configuration, not routing.
var (
	// longRunningActionTypes get extended timeouts (30 min) for external operations.
	// Examples: API calls, email delivery, inventory allocation.
	longRunningActionTypes = map[string]bool{
		"allocate_inventory":   true,
		"reserve_inventory":    true,
		"send_email":           true,
		"credit_check":         true,
		"fraud_detection":      true,
		"third_party_api_call": true,
		"reserve_shipping":     true,
	}

	// humanActionTypes require human interaction and get multi-day timeouts.
	// Examples: manager approvals, manual reviews.
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

	// Build graph executor with pre-sorted edge indexes.
	executor := NewGraphExecutor(input.Graph)

	// Find start actions (edges with source_action_id = nil).
	startActions := executor.GetStartActions()
	if len(startActions) == 0 {
		logger.Info("Empty workflow - no start actions found")
		return nil
	}

	// Execute from start (may be multiple parallel start actions).
	return executeActions(ctx, executor, startActions, mergedCtx, input)
}

// shouldContinueAsNew is the testable core of the threshold check.
// Extracted so unit tests can verify boundary behavior without needing
// workflow.Context or mocked GetInfo().
func shouldContinueAsNew(currentHistoryLength int) bool {
	return currentHistoryLength > HistoryLengthThreshold
}

// checkContinueAsNew returns a ContinueAsNewError if history has grown too large.
// This prevents hitting Temporal's 50K event limit on long-running workflows.
// The full MergedContext is preserved via ContinuationState, maintaining the
// structured ActionResults map and Flattened template resolution data.
func checkContinueAsNew(ctx workflow.Context, input WorkflowInput, mergedCtx *MergedContext) error {
	info := workflow.GetInfo(ctx)
	if shouldContinueAsNew(int(info.GetCurrentHistoryLength())) {
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

	// Check if we need to continue-as-new to avoid history size limits.
	if err := checkContinueAsNew(ctx, input, mergedCtx); err != nil {
		return err
	}

	if len(actions) == 1 {
		// Sequential execution - single path.
		return executeSingleAction(ctx, executor, actions[0], mergedCtx, input)
	}

	// Parallel execution - check for convergence.
	convergencePoint := executor.FindConvergencePoint(actions)

	if convergencePoint == nil {
		// Fire-and-forget parallel branches - no convergence.
		return executeFireAndForget(ctx, executor, actions, mergedCtx, input)
	}

	// Parallel with convergence - must wait for all branches.
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

	// Prepare activity input.
	activityInput := ActionActivityInput{
		ActionID:    action.ID,
		ActionName:  action.Name,
		ActionType:  action.ActionType,
		Config:      action.Config,
		Context:     mergedCtx.Flattened,
		RuleID:      input.RuleID,
		ExecutionID: input.ExecutionID,
		RuleName:    input.RuleName,
	}

	// Configure activity options based on action type.
	activityCtx := workflow.WithActivityOptions(ctx, activityOptions(action.ActionType))

	// Route to the appropriate activity function.
	// Async and human actions use ExecuteAsyncActionActivity which returns
	// ErrResultPending and waits for external completion via AsyncCompleter.
	activityFunc := selectActivityFunc(action.ActionType)

	var result ActionActivityOutput
	if err := workflow.ExecuteActivity(activityCtx, activityFunc, activityInput).Get(ctx, &result); err != nil {
		logger.Error("Action failed",
			"action_id", action.ID,
			"action_name", action.Name,
			"error", err,
		)
		return fmt.Errorf("execute action %s (%s): %w", action.Name, action.ID, err)
	}

	// Merge result into context for subsequent actions.
	mergedCtx.MergeResult(action.Name, result.Result)

	logger.Info("Action completed",
		"action_id", action.ID,
		"action_name", action.Name,
		"success", result.Success,
	)

	// Get next actions based on result and edge types.
	nextActions := executor.GetNextActions(action.ID, result.Result)

	if len(nextActions) == 0 {
		// End of this path.
		return nil
	}

	// Continue execution (recursive).
	return executeActions(ctx, executor, nextActions, mergedCtx, input)
}

// =============================================================================
// Parallel Execution
// =============================================================================

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
				RuleID:           input.RuleID,
				ExecutionID:      input.ExecutionID,
				RuleName:         input.RuleName,
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

	// Parent returns immediately - fire-and-forget branches run independently.
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

	// Create selector for waiting on all branches.
	selector := workflow.NewSelector(ctx)
	branchResults := make([]BranchOutput, len(branches))
	branchErrors := make([]error, len(branches))

	for i, branch := range branches {
		branchIndex := i
		branchAction := branch

		// Each branch runs as a child workflow.
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
				RuleID:           input.RuleID,
				ExecutionID:      input.ExecutionID,
				RuleName:         input.RuleName,
			},
		)

		selector.AddFuture(future, func(f workflow.Future) {
			var output BranchOutput
			branchErrors[branchIndex] = f.Get(ctx, &output)
			branchResults[branchIndex] = output
		})
	}

	// Wait for all branches.
	for range len(branches) {
		selector.Select(ctx)
	}

	// Check for errors - any branch failure fails the entire parallel group.
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

	// Continue from convergence point.
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

		// Execute action.
		activityInput := ActionActivityInput{
			ActionID:    currentAction.ID,
			ActionName:  currentAction.Name,
			ActionType:  currentAction.ActionType,
			Config:      currentAction.Config,
			Context:     mergedCtx.Flattened,
			RuleID:      input.RuleID,
			ExecutionID: input.ExecutionID,
			RuleName:    input.RuleName,
		}

		activityCtx := workflow.WithActivityOptions(ctx, activityOptions(currentAction.ActionType))
		activityFunc := selectActivityFunc(currentAction.ActionType)

		var result ActionActivityOutput
		if err := workflow.ExecuteActivity(activityCtx, activityFunc, activityInput).Get(ctx, &result); err != nil {
			return BranchOutput{}, fmt.Errorf("execute action %s (%s): %w", currentAction.Name, currentAction.ID, err)
		}

		mergedCtx.MergeResult(currentAction.Name, result.Result)

		// Get next action. Conditional edges are resolved by GetNextActions based
		// on result, so even conditionals should yield a single next action.
		nextActions := executor.GetNextActions(currentAction.ID, result.Result)
		if len(nextActions) == 0 {
			// End of branch path (fire-and-forget) or dead end before convergence.
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

// =============================================================================
// Action Type Helpers
// =============================================================================

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

	// Long-running actions (external APIs, email, inventory) get longer timeouts.
	// Temporal handles retries natively - no need for RabbitMQ.
	if isLongRunningAction(actionType) {
		ao.StartToCloseTimeout = 30 * time.Minute
		ao.HeartbeatTimeout = time.Minute
		ao.RetryPolicy.MaximumAttempts = 3 // Temporal retries are safe
	}

	// Human-in-the-loop actions (approvals, manual review) can take days.
	// MaximumAttempts=1 prevents duplicate approval requests on retry.
	if isHumanAction(actionType) {
		ao.StartToCloseTimeout = 7 * 24 * time.Hour // 7 days
		ao.HeartbeatTimeout = time.Hour
		ao.RetryPolicy.MaximumAttempts = 1
	}

	return ao
}

// selectActivityFunc returns the activity method name based on action type.
// Temporal resolves struct method names by string when Activities is registered
// via worker.RegisterActivity(&Activities{...}).
//
// All actions route to ExecuteActionActivity (synchronous execution).
// Temporal handles retries and timeouts natively - no need for async completion pattern.
func selectActivityFunc(actionType string) string {
	// All actions use the synchronous activity. Temporal handles:
	// - Retries with backoff (configured in activityOptions)
	// - Timeouts (30min for long-running, 7 days for human)
	// - Durable execution and failure recovery
	_ = actionType // unused but kept for signature consistency
	return "ExecuteActionActivity"
}

// isLongRunningAction returns true for actions that involve external operations.
// These get longer timeouts (30min), heartbeat requirements, and limited retries.
func isLongRunningAction(actionType string) bool {
	return longRunningActionTypes[actionType]
}

// isHumanAction returns true for actions that require human interaction.
// These get multi-day timeouts (7 days), heartbeat requirements, and no retries.
func isHumanAction(actionType string) bool {
	return humanActionTypes[actionType]
}

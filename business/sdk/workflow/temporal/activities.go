package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Activities Struct
// =============================================================================

// Activities holds both registries needed by Temporal activity methods.
// Register an instance on the worker (Phase 9) via worker.RegisterActivity(&Activities{...}).
// Temporal resolves struct method names by string, so workflow.go references
// "ExecuteActionActivity" and "ExecuteAsyncActionActivity" as strings.
type Activities struct {
	Registry      *workflow.ActionRegistry
	AsyncRegistry *AsyncRegistry
}

// =============================================================================
// Synchronous Activity
// =============================================================================

// ExecuteActionActivity dispatches to the appropriate handler based on action type.
// This wraps existing workflow.ActionHandler implementations as Temporal activities.
func (a *Activities) ExecuteActionActivity(ctx context.Context, input ActionActivityInput) (ActionActivityOutput, error) {
	logger := activity.GetLogger(ctx)
	info := activity.GetInfo(ctx)

	logger.Info("Executing action activity",
		"action_id", input.ActionID,
		"action_name", input.ActionName,
		"action_type", input.ActionType,
		"workflow_id", info.WorkflowExecution.ID,
	)

	if err := input.Validate(); err != nil {
		return ActionActivityOutput{}, fmt.Errorf("action %s (%s): invalid input: %w", input.ActionName, input.ActionID, err)
	}

	if a.Registry == nil {
		return ActionActivityOutput{}, fmt.Errorf("action %s (%s): activity dependencies not initialized", input.ActionName, input.ActionID)
	}

	handler, exists := a.Registry.Get(input.ActionType)
	if !exists {
		return ActionActivityOutput{}, fmt.Errorf("action %s (%s): no handler registered for type %s", input.ActionName, input.ActionID, input.ActionType)
	}

	execCtx := buildExecContext(input)

	result, err := handler.Execute(ctx, input.Config, execCtx)
	if err != nil {
		logger.Error("Action execution failed",
			"action_id", input.ActionID,
			"action_name", input.ActionName,
			"action_type", input.ActionType,
			"workflow_id", info.WorkflowExecution.ID,
			"error", err,
		)
		return ActionActivityOutput{}, fmt.Errorf("action %s (%s): execute: %w", input.ActionName, input.ActionID, err)
	}

	resultMap := toResultMap(result)

	// Inject default output if the handler didn't set one.
	// This ensures every action result has an "output" key for the graph executor.
	if _, hasOutput := resultMap["output"]; !hasOutput {
		resultMap["output"] = "success"
	}

	logger.Info("Action execution succeeded",
		"action_id", input.ActionID,
		"action_name", input.ActionName,
		"workflow_id", info.WorkflowExecution.ID,
	)

	return ActionActivityOutput{
		ActionID:   input.ActionID,
		ActionName: input.ActionName,
		Result:     resultMap,
		Success:    true,
	}, nil
}

// =============================================================================
// Async Activity
// =============================================================================

// ExecuteAsyncActionActivity handles actions that complete asynchronously.
// The activity returns ErrResultPending; completion happens via AsyncCompleter.
//
// Flow:
//  1. Get task token from activity context
//  2. Call handler.StartAsync with task token
//  3. Return ErrResultPending (activity stays open)
//  4. External system processes work
//  5. External system calls AsyncCompleter.Complete(taskToken, result)
//  6. Temporal resumes workflow with the result
func (a *Activities) ExecuteAsyncActionActivity(ctx context.Context, input ActionActivityInput) (ActionActivityOutput, error) {
	logger := activity.GetLogger(ctx)
	info := activity.GetInfo(ctx)

	logger.Info("Starting async action activity",
		"action_id", input.ActionID,
		"action_name", input.ActionName,
		"action_type", input.ActionType,
		"workflow_id", info.WorkflowExecution.ID,
	)

	if err := input.Validate(); err != nil {
		return ActionActivityOutput{}, fmt.Errorf("async action %s (%s): invalid input: %w", input.ActionName, input.ActionID, err)
	}

	if a.AsyncRegistry == nil {
		return ActionActivityOutput{}, fmt.Errorf("async action %s (%s): async activity dependencies not initialized", input.ActionName, input.ActionID)
	}

	taskToken := info.TaskToken

	execCtx := buildExecContext(input)

	handler, exists := a.AsyncRegistry.Get(input.ActionType)
	if !exists {
		return ActionActivityOutput{}, fmt.Errorf("async action %s (%s): no handler registered for type %s", input.ActionName, input.ActionID, input.ActionType)
	}

	if err := handler.StartAsync(ctx, input.Config, execCtx, taskToken); err != nil {
		logger.Error("Async action start failed",
			"action_id", input.ActionID,
			"action_name", input.ActionName,
			"action_type", input.ActionType,
			"workflow_id", info.WorkflowExecution.ID,
			"error", err,
		)
		return ActionActivityOutput{}, fmt.Errorf("async action %s (%s): start: %w", input.ActionName, input.ActionID, err)
	}

	logger.Info("Async action started, awaiting external completion",
		"action_id", input.ActionID,
		"action_name", input.ActionName,
		"workflow_id", info.WorkflowExecution.ID,
	)

	return ActionActivityOutput{}, activity.ErrResultPending
}

// =============================================================================
// Helpers
// =============================================================================

// buildExecContext converts ActionActivityInput into the existing
// workflow.ActionExecutionContext used by all action handlers.
//
// Typed fields (RuleID, ExecutionID, RuleName) are populated from explicit
// ActionActivityInput fields rather than string-parsing from the context map.
// EntityID, EntityName, EventType, UserID, and FieldChanges are extracted
// from the Context map (originating from trigger data in Phase 7).
func buildExecContext(input ActionActivityInput) workflow.ActionExecutionContext {
	execCtx := workflow.ActionExecutionContext{
		RawData:       input.Context,
		ExecutionID:   input.ExecutionID,
		RuleName:      input.RuleName,
		ActionName:    input.ActionName,
		TriggerSource: "automation",
		Timestamp:     time.Now(),
	}

	if input.RuleID != uuid.Nil {
		ruleID := input.RuleID
		execCtx.RuleID = &ruleID
	}

	// Extract typed fields from merged context when available.
	// These fields originate from trigger data (Phase 7 buildTriggerData)
	// and flow through MergedContext.Flattened -> ActionActivityInput.Context.
	if v, ok := input.Context["entity_id"].(string); ok {
		if id, err := uuid.Parse(v); err == nil {
			execCtx.EntityID = id
		}
	}
	if v, ok := input.Context["entity_name"].(string); ok {
		execCtx.EntityName = v
	}
	if v, ok := input.Context["event_type"].(string); ok {
		execCtx.EventType = v
	}
	if v, ok := input.Context["user_id"].(string); ok {
		if id, err := uuid.Parse(v); err == nil {
			execCtx.UserID = id
		}
	}

	// Extract FieldChanges from context map for update events.
	if changes, ok := input.Context["field_changes"].(map[string]any); ok {
		fieldChanges := make(map[string]workflow.FieldChange, len(changes))
		for field, change := range changes {
			if changeMap, ok := change.(map[string]any); ok {
				fieldChanges[field] = workflow.FieldChange{
					OldValue: changeMap["old_value"],
					NewValue: changeMap["new_value"],
				}
			}
		}
		execCtx.FieldChanges = fieldChanges
	}

	return execCtx
}

// toResultMap converts any result type to a map for context merging.
// Handles: nil, map[string]any, and struct types.
// Struct types are marshaled to JSON then unmarshaled into a map.
//
// IMPORTANT: The JSON roundtrip converts all numeric types to float64.
// This means int64 values (e.g., database IDs, timestamps) lose precision
// for values > 2^53. This is acceptable for context merging where values
// are primarily used for template resolution (string interpolation).
// Handlers that need precise large integers should return string representations.
func toResultMap(result any) map[string]any {
	if result == nil {
		return map[string]any{}
	}

	if m, ok := result.(map[string]any); ok {
		return m
	}

	data, err := json.Marshal(result)
	if err != nil {
		return map[string]any{"raw": fmt.Sprintf("%v", result)}
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return map[string]any{"raw": fmt.Sprintf("%v", result)}
	}

	return m
}

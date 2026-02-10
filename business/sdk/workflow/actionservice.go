package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// Error Definitions
// =============================================================================

var (
	// ErrActionNotFound is returned when an action type is not registered
	ErrActionNotFound = errors.New("action type not found")

	// ErrManualExecutionNotSupported is returned when trying to manually execute
	// an action that doesn't support manual execution (e.g., update_field)
	ErrManualExecutionNotSupported = errors.New("manual execution not supported for this action type")
)

// =============================================================================
// Request/Response Types
// =============================================================================

// ExecuteRequest represents a request to execute an action
type ExecuteRequest struct {
	ActionType string                 `json:"action_type"`
	Config     json.RawMessage        `json:"config"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
	EntityName string                 `json:"entity_name,omitempty"`
	RawData    map[string]interface{} `json:"raw_data,omitempty"`
	UserID     uuid.UUID              `json:"user_id"`
}

// ExecuteResult represents the result of an action execution
type ExecuteResult struct {
	ExecutionID uuid.UUID   `json:"execution_id"`
	ActionType  string      `json:"action_type"`
	Status      string      `json:"status"` // "completed", "queued", "failed"
	Result      interface{} `json:"result,omitempty"`
	TrackingURL string      `json:"tracking_url,omitempty"`
	Error       string      `json:"error,omitempty"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}

// ExecutionStatusInfo represents the status of an execution
type ExecutionStatusInfo struct {
	ExecutionID uuid.UUID   `json:"execution_id"`
	ActionType  string      `json:"action_type"`
	Status      string      `json:"status"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}

// ActionInfo represents metadata about an available action
type ActionInfo struct {
	Type                     string `json:"type"`
	Description              string `json:"description"`
	IsAsync                  bool   `json:"is_async"`
	SupportsManualExecution  bool   `json:"supports_manual_execution"`
}

// =============================================================================
// ActionService
// =============================================================================

// ActionService provides unified action execution for both automation and manual triggers.
// It wraps the ActionRegistry and provides execution recording, permission checking,
// and a consistent interface for the API layer.
type ActionService struct {
	log      *logger.Logger
	db       *sqlx.DB
	registry *ActionRegistry
}

// NewActionService creates a new ActionService
func NewActionService(log *logger.Logger, db *sqlx.DB, registry *ActionRegistry) *ActionService {
	return &ActionService{
		log:      log,
		db:       db,
		registry: registry,
	}
}

// Execute executes an action, either for automation or manual trigger.
// For manual executions, it validates that the action supports manual execution.
func (s *ActionService) Execute(ctx context.Context, req ExecuteRequest, triggerSource string) (*ExecuteResult, error) {
	startTime := time.Now()
	executionID := uuid.New()

	s.log.Info(ctx, "ActionService.Execute starting",
		"execution_id", executionID,
		"action_type", req.ActionType,
		"trigger_source", triggerSource,
		"user_id", req.UserID)

	// Look up the handler
	handler, exists := s.registry.Get(req.ActionType)
	if !exists {
		return &ExecuteResult{
			ExecutionID: executionID,
			ActionType:  req.ActionType,
			Status:      "failed",
			Error:       ErrActionNotFound.Error(),
			StartedAt:   startTime,
		}, ErrActionNotFound
	}

	// For manual executions, verify the action supports it
	if triggerSource == TriggerSourceManual && !handler.SupportsManualExecution() {
		s.log.Warn(ctx, "Manual execution attempted on unsupported action type",
			"action_type", req.ActionType)
		return &ExecuteResult{
			ExecutionID: executionID,
			ActionType:  req.ActionType,
			Status:      "failed",
			Error:       ErrManualExecutionNotSupported.Error(),
			StartedAt:   startTime,
		}, ErrManualExecutionNotSupported
	}

	// Validate the configuration
	if err := handler.Validate(req.Config); err != nil {
		s.log.Error(ctx, "Action configuration validation failed",
			"execution_id", executionID,
			"action_type", req.ActionType,
			"error", err)
		return &ExecuteResult{
			ExecutionID: executionID,
			ActionType:  req.ActionType,
			Status:      "failed",
			Error:       fmt.Sprintf("configuration validation failed: %v", err),
			StartedAt:   startTime,
		}, err
	}

	// Build execution context
	execContext := ActionExecutionContext{
		EntityName:    req.EntityName,
		EventType:     EventTypeManualTrigger,
		RawData:       req.RawData,
		Timestamp:     startTime,
		UserID:        req.UserID,
		RuleID:        nil, // nil for manual executions
		RuleName:      "",
		ExecutionID:   executionID,
		TriggerSource: triggerSource,
	}

	// Set EntityID if provided
	if req.EntityID != nil {
		execContext.EntityID = *req.EntityID
	}

	// For automation triggers, we might have a rule context
	// (this will be set by the caller when coming from the workflow engine)
	if triggerSource == TriggerSourceAutomation {
		execContext.EventType = "" // Will be set by caller
	}

	// Execute the action
	result, err := handler.Execute(ctx, req.Config, execContext)

	completedAt := time.Now()
	execResult := &ExecuteResult{
		ExecutionID: executionID,
		ActionType:  req.ActionType,
		StartedAt:   startTime,
		CompletedAt: &completedAt,
	}

	if err != nil {
		execResult.Status = "failed"
		execResult.Error = err.Error()
		s.log.Error(ctx, "Action execution failed",
			"execution_id", executionID,
			"action_type", req.ActionType,
			"error", err)
	} else {
		// For async actions, status is "queued" if result indicates it was queued
		if handler.IsAsync() {
			execResult.Status = "queued"
			// For async actions, the tracking URL would be set based on the result
			execResult.TrackingURL = fmt.Sprintf("/v1/workflow/executions/%s", executionID)
		} else {
			execResult.Status = "completed"
		}
		execResult.Result = result
	}

	// Record the execution
	if recordErr := s.recordExecution(ctx, execResult, req, triggerSource); recordErr != nil {
		s.log.Error(ctx, "Failed to record execution",
			"execution_id", executionID,
			"error", recordErr)
		// Don't fail the execution if recording fails
	}

	return execResult, err
}

// ExecuteForAutomation executes an action in the context of workflow automation.
// This is called by the ActionExecutor when processing rules.
func (s *ActionService) ExecuteForAutomation(ctx context.Context, req ExecuteRequest, ruleID uuid.UUID, ruleName string, eventType string) (*ExecuteResult, error) {
	startTime := time.Now()
	executionID := uuid.New()

	// Look up the handler
	handler, exists := s.registry.Get(req.ActionType)
	if !exists {
		return &ExecuteResult{
			ExecutionID: executionID,
			ActionType:  req.ActionType,
			Status:      "failed",
			Error:       ErrActionNotFound.Error(),
			StartedAt:   startTime,
		}, ErrActionNotFound
	}

	// Validate the configuration
	if err := handler.Validate(req.Config); err != nil {
		return &ExecuteResult{
			ExecutionID: executionID,
			ActionType:  req.ActionType,
			Status:      "failed",
			Error:       fmt.Sprintf("configuration validation failed: %v", err),
			StartedAt:   startTime,
		}, err
	}

	// Build execution context with rule information
	execContext := ActionExecutionContext{
		EntityName:    req.EntityName,
		EventType:     eventType,
		RawData:       req.RawData,
		Timestamp:     startTime,
		UserID:        req.UserID,
		RuleID:        &ruleID,
		RuleName:      ruleName,
		ExecutionID:   executionID,
		TriggerSource: TriggerSourceAutomation,
	}

	if req.EntityID != nil {
		execContext.EntityID = *req.EntityID
	}

	// Execute the action
	result, err := handler.Execute(ctx, req.Config, execContext)

	completedAt := time.Now()
	execResult := &ExecuteResult{
		ExecutionID: executionID,
		ActionType:  req.ActionType,
		StartedAt:   startTime,
		CompletedAt: &completedAt,
	}

	if err != nil {
		execResult.Status = "failed"
		execResult.Error = err.Error()
	} else {
		if handler.IsAsync() {
			execResult.Status = "queued"
		} else {
			execResult.Status = "completed"
		}
		execResult.Result = result
	}

	return execResult, err
}

// GetExecutionStatus retrieves the status of an execution by ID
func (s *ActionService) GetExecutionStatus(ctx context.Context, executionID uuid.UUID) (*ExecutionStatusInfo, error) {
	query := `
		SELECT
			id,
			action_type,
			status,
			actions_executed,
			error_message,
			executed_at
		FROM workflow.automation_executions
		WHERE id = $1
	`

	var row struct {
		ID              uuid.UUID       `db:"id"`
		ActionType      string          `db:"action_type"`
		Status          string          `db:"status"`
		ActionsExecuted json.RawMessage `db:"actions_executed"`
		ErrorMessage    string          `db:"error_message"`
		ExecutedAt      time.Time       `db:"executed_at"`
	}

	if err := s.db.GetContext(ctx, &row, query, executionID); err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}

	statusInfo := &ExecutionStatusInfo{
		ExecutionID: row.ID,
		ActionType:  row.ActionType,
		Status:      row.Status,
		Error:       row.ErrorMessage,
		StartedAt:   row.ExecutedAt,
	}

	// Parse the result from actions_executed if present
	if len(row.ActionsExecuted) > 0 {
		var result interface{}
		if err := json.Unmarshal(row.ActionsExecuted, &result); err == nil {
			statusInfo.Result = result
		}
	}

	return statusInfo, nil
}

// ListAvailableActions returns all actions that can be executed manually
func (s *ActionService) ListAvailableActions() []ActionInfo {
	allTypes := s.registry.GetAll()
	available := make([]ActionInfo, 0, len(allTypes))

	for _, actionType := range allTypes {
		handler, exists := s.registry.Get(actionType)
		if !exists {
			continue
		}

		info := ActionInfo{
			Type:                    actionType,
			Description:             handler.GetDescription(),
			IsAsync:                 handler.IsAsync(),
			SupportsManualExecution: handler.SupportsManualExecution(),
		}
		available = append(available, info)
	}

	return available
}

// ListManuallyExecutableActions returns only actions that support manual execution
func (s *ActionService) ListManuallyExecutableActions() []ActionInfo {
	all := s.ListAvailableActions()
	manual := make([]ActionInfo, 0)

	for _, info := range all {
		if info.SupportsManualExecution {
			manual = append(manual, info)
		}
	}

	return manual
}

// GetRegistry returns the underlying ActionRegistry
func (s *ActionService) GetRegistry() *ActionRegistry {
	return s.registry
}

// recordExecution records an execution to the database
func (s *ActionService) recordExecution(ctx context.Context, result *ExecuteResult, req ExecuteRequest, triggerSource string) error {
	// Serialize the result for storage
	var actionsExecuted json.RawMessage
	if result.Result != nil {
		data, err := json.Marshal(result.Result)
		if err == nil {
			actionsExecuted = data
		}
	}

	// Build trigger data
	triggerData, _ := json.Marshal(map[string]interface{}{
		"entity_id":   req.EntityID,
		"entity_name": req.EntityName,
		"raw_data":    req.RawData,
	})

	// Calculate execution time
	var executionTimeMs int
	if result.CompletedAt != nil {
		executionTimeMs = int(result.CompletedAt.Sub(result.StartedAt).Milliseconds())
	}

	// Map result status to ExecutionStatus
	var status ExecutionStatus
	switch result.Status {
	case "completed":
		status = StatusCompleted
	case "queued":
		status = StatusRunning
	case "failed":
		status = StatusFailed
	default:
		status = StatusPending
	}

	query := `
		INSERT INTO workflow.automation_executions (
			id,
			automation_rules_id,
			entity_type,
			trigger_data,
			actions_executed,
			status,
			error_message,
			execution_time_ms,
			executed_at,
			trigger_source,
			executed_by,
			action_type
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := s.db.ExecContext(ctx, query,
		result.ExecutionID,
		nil, // automation_rules_id is null for manual executions
		req.EntityName,
		triggerData,
		actionsExecuted,
		status,
		result.Error,
		executionTimeMs,
		result.StartedAt,
		triggerSource,
		req.UserID,
		req.ActionType,
	)

	return err
}

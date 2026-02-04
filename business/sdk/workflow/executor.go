package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ActionExecution represents an active action execution
type ActionExecution struct {
	ActionID  uuid.UUID
	RuleID    *uuid.UUID // Pointer to support nil for manual executions
	StartedAt time.Time
	Status    string // TODO: Make into execution status
	Context   ActionExecutionContext
}

// ActionExecutionStats tracks action execution statistics
type ActionExecutionStats struct {
	TotalActionsExecuted   int
	SuccessfulExecutions   int
	FailedExecutions       int
	AverageExecutionTimeMs float64
	LastExecutedAt         *time.Time
}

// ActionExecutor manages the execution of rule actions
type ActionExecutor struct {
	log         *logger.Logger
	db          *sqlx.DB
	workflowBus *Business

	registry *ActionRegistry

	// Template processor
	templateProc *TemplateProcessor

	// Execution tracking
	activeExecs sync.Map // map[string]*ActionExecution
	history     []BatchExecutionResult
	stats       ActionExecutionStats
	mu          sync.RWMutex

	// Configuration
	maxParallelActions int
	retryEnabled       bool
	maxRetries         int
}

// BatchExecutionResult represents the result of executing all actions for a rule
type BatchExecutionResult struct {
	RuleID             uuid.UUID              `json:"rule_id"`
	RuleName           string                 `json:"rule_name"`
	TotalActions       int                    `json:"total_actions"`
	SuccessfulActions  int                    `json:"successful_actions"`
	FailedActions      int                    `json:"failed_actions"`
	SkippedActions     int                    `json:"skipped_actions"`
	ActionResults      []ActionResult         `json:"action_results"`
	TotalExecutionTime time.Duration          `json:"total_execution_time_ms"`
	StartedAt          time.Time              `json:"started_at"`
	CompletedAt        time.Time              `json:"completed_at"`
	Context            ActionExecutionContext `json:"context"`
	Status             string                 `json:"status"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
}

// ActionValidationResult represents the result of validating an action configuration
type ActionValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(log *logger.Logger, db *sqlx.DB, workflowBus *Business) *ActionExecutor {
	ae := &ActionExecutor{
		log:                log,
		db:                 db,
		workflowBus:        workflowBus,
		registry:           NewActionRegistry(),
		templateProc:       NewTemplateProcessor(DefaultTemplateProcessingOptions()),
		history:            make([]BatchExecutionResult, 0),
		maxParallelActions: 10,
		retryEnabled:       true,
		maxRetries:         3,
	}

	return ae
}

// GetRegistry returns the action registry for external registration
func (ae *ActionExecutor) GetRegistry() *ActionRegistry {
	return ae.registry
}

// Initialize initializes the action executor
func (ae *ActionExecutor) Initialize(ctx context.Context) error {
	ae.log.Info(ctx, "Initializing action executor...")

	// Any initialization logic here

	ae.log.Info(ctx, "Action executor initialized successfully")
	return nil
}

// ExecuteRuleActions executes all actions for a given rule
func (ae *ActionExecutor) ExecuteRuleActions(ctx context.Context, ruleID uuid.UUID, executionContext ActionExecutionContext) (BatchExecutionResult, error) {
	startTime := time.Now()

	// Load actions for the rule
	actions, err := ae.workflowBus.QueryRoleActionsViewByRuleID(ctx, ruleID)
	if err != nil {
		return BatchExecutionResult{}, fmt.Errorf("failed to load rule actions: %w", err)
	}

	// Get rule name for result
	ruleName := ae.getRuleName(ctx, ruleID)
	executionContext.RuleName = ruleName

	ae.log.Info(ctx, "Executing actions for rule",
		"ruleID", ruleID,
		"ruleName", ruleName,
		"actionCount", len(actions))

	// Execute actions in order
	actionResults := make([]ActionResult, 0, len(actions))
	successCount := 0
	failedCount := 0
	skippedCount := 0

	for _, action := range actions {
		if !action.IsActive {
			skippedCount++
			actionResults = append(actionResults, ActionResult{
				ActionID:   action.ID,
				ActionName: action.Name,
				ActionType: ae.getActionType(action),
				Status:     "skipped",
				StartedAt:  time.Now(),
			})
			continue
		}

		result := ae.executeAction(ctx, action, executionContext)
		actionResults = append(actionResults, result)

		switch result.Status {
		case "success":
			successCount++
		case "failed":
			failedCount++
			// Stop on critical failure if configured
			if ae.shouldStopOnFailure(action) {
				ae.log.Error(ctx, "Critical action failed, stopping execution",
					"actionID", action.ID,
					"actionName", action.Name)
				break
			}
		case "skipped":
			skippedCount++
		}
	}

	completedAt := time.Now()
	duration := completedAt.Sub(startTime)

	// Determine overall status
	status := "success"
	errorMessage := ""
	if failedCount > 0 {
		status = "failed"
		errorMessage = fmt.Sprintf("%d actions failed", failedCount)
	}

	result := BatchExecutionResult{
		RuleID:             ruleID,
		RuleName:           ruleName,
		TotalActions:       len(actions),
		SuccessfulActions:  successCount,
		FailedActions:      failedCount,
		SkippedActions:     skippedCount,
		ActionResults:      actionResults,
		TotalExecutionTime: duration,
		StartedAt:          startTime,
		CompletedAt:        completedAt,
		Context:            executionContext,
		Status:             status,
		ErrorMessage:       errorMessage,
	}

	// Update statistics
	ae.updateStats(result)

	// Add to history
	ae.mu.Lock()
	ae.history = append(ae.history, result)
	if len(ae.history) > 1000 { // Limit history size
		ae.history = ae.history[len(ae.history)-1000:]
	}
	ae.mu.Unlock()

	return result, nil
}

// ExecuteRuleActionsGraph executes actions following the edge graph.
// Falls back to linear execution_order if no edges exist (backwards compatible).
func (ae *ActionExecutor) ExecuteRuleActionsGraph(ctx context.Context, ruleID uuid.UUID, executionContext ActionExecutionContext) (BatchExecutionResult, error) {
	startTime := time.Now()

	// Load edges for this rule
	edges, err := ae.workflowBus.QueryEdgesByRuleID(ctx, ruleID)
	if err != nil {
		return BatchExecutionResult{}, fmt.Errorf("failed to load edges: %w", err)
	}

	// Backwards compatibility: if no edges, use linear execution
	if len(edges) == 0 {
		ae.log.Info(ctx, "No edges found for rule, falling back to linear execution",
			"ruleID", ruleID)
		return ae.ExecuteRuleActions(ctx, ruleID, executionContext)
	}

	// Load all actions
	actions, err := ae.workflowBus.QueryRoleActionsViewByRuleID(ctx, ruleID)
	if err != nil {
		return BatchExecutionResult{}, fmt.Errorf("failed to load actions: %w", err)
	}

	// Get rule name for result
	ruleName := ae.getRuleName(ctx, ruleID)
	executionContext.RuleName = ruleName

	ae.log.Info(ctx, "Executing actions using graph traversal",
		"ruleID", ruleID,
		"ruleName", ruleName,
		"actionCount", len(actions),
		"edgeCount", len(edges))

	// Build action map for quick lookup
	actionMap := make(map[uuid.UUID]RuleActionView)
	for _, action := range actions {
		actionMap[action.ID] = action
	}

	// Build adjacency list from edges
	outgoingEdges := make(map[uuid.UUID][]ActionEdge) // source -> edges
	var startEdges []ActionEdge

	for _, edge := range edges {
		if edge.SourceActionID == nil {
			startEdges = append(startEdges, edge)
		} else {
			outgoingEdges[*edge.SourceActionID] = append(outgoingEdges[*edge.SourceActionID], edge)
		}
	}

	// Validate we have start edges
	if len(startEdges) == 0 {
		ae.log.Warn(ctx, "No start edges found for rule, falling back to linear execution",
			"ruleID", ruleID)
		return ae.ExecuteRuleActions(ctx, ruleID, executionContext)
	}

	// Sort start edges by edge_order for deterministic execution
	sortEdgesByOrder(startEdges)

	// Execute using BFS from start edges
	actionResults := make([]ActionResult, 0)
	executed := make(map[uuid.UUID]bool)
	resultMap := make(map[uuid.UUID]ActionResult) // Track results for branch decisions
	queue := make([]uuid.UUID, 0)

	successCount := 0
	failedCount := 0
	skippedCount := 0

	// Add start edge targets to queue (in order)
	for _, edge := range startEdges {
		queue = append(queue, edge.TargetActionID)
	}

	for len(queue) > 0 {
		actionID := queue[0]
		queue = queue[1:]

		if executed[actionID] {
			continue
		}
		executed[actionID] = true

		action, exists := actionMap[actionID]
		if !exists {
			ae.log.Warn(ctx, "Action not found in action map",
				"actionID", actionID,
				"ruleID", ruleID)
			continue
		}

		if !action.IsActive {
			skippedCount++
			result := ActionResult{
				ActionID:   action.ID,
				ActionName: action.Name,
				ActionType: ae.getActionType(action),
				Status:     "skipped",
				StartedAt:  time.Now(),
			}
			actionResults = append(actionResults, result)
			resultMap[actionID] = result
			continue
		}

		// Execute the action
		result := ae.executeAction(ctx, action, executionContext)
		actionResults = append(actionResults, result)
		resultMap[actionID] = result

		switch result.Status {
		case "success":
			successCount++
		case "failed":
			failedCount++
			// Stop on critical failure if configured
			if ae.shouldStopOnFailure(action) {
				ae.log.Error(ctx, "Critical action failed, stopping graph execution",
					"actionID", action.ID,
					"actionName", action.Name)
				// Don't add more actions to queue
				continue
			}
		case "skipped":
			skippedCount++
		}

		// Determine which edges to follow based on result
		nextEdges := outgoingEdges[actionID]
		sortEdgesByOrder(nextEdges)

		for _, edge := range nextEdges {
			shouldFollow := ae.ShouldFollowEdge(edge, result)

			if shouldFollow && !executed[edge.TargetActionID] {
				queue = append(queue, edge.TargetActionID)
			}
		}
	}

	completedAt := time.Now()
	duration := completedAt.Sub(startTime)

	// Determine overall status
	status := "success"
	errorMessage := ""
	if failedCount > 0 {
		status = "failed"
		errorMessage = fmt.Sprintf("%d actions failed", failedCount)
	}

	batchResult := BatchExecutionResult{
		RuleID:             ruleID,
		RuleName:           ruleName,
		TotalActions:       len(actionResults),
		SuccessfulActions:  successCount,
		FailedActions:      failedCount,
		SkippedActions:     skippedCount,
		ActionResults:      actionResults,
		TotalExecutionTime: duration,
		StartedAt:          startTime,
		CompletedAt:        completedAt,
		Context:            executionContext,
		Status:             status,
		ErrorMessage:       errorMessage,
	}

	// Update statistics
	ae.updateStats(batchResult)

	// Add to history
	ae.mu.Lock()
	ae.history = append(ae.history, batchResult)
	if len(ae.history) > 1000 {
		ae.history = ae.history[len(ae.history)-1000:]
	}
	ae.mu.Unlock()

	ae.log.Info(ctx, "Graph execution completed",
		"ruleID", ruleID,
		"ruleName", ruleName,
		"totalExecuted", len(actionResults),
		"successful", successCount,
		"failed", failedCount,
		"skipped", skippedCount,
		"duration", duration)

	return batchResult, nil
}

// ShouldFollowEdge determines if an edge should be followed based on the action result.
// Exported for testing purposes.
func (ae *ActionExecutor) ShouldFollowEdge(edge ActionEdge, result ActionResult) bool {
	switch edge.EdgeType {
	case EdgeTypeAlways, EdgeTypeSequence:
		// Always follow these edges regardless of result
		return true
	case EdgeTypeTrueBranch:
		// Only follow if the action returned true_branch
		return result.BranchTaken == EdgeTypeTrueBranch
	case EdgeTypeFalseBranch:
		// Only follow if the action returned false_branch
		return result.BranchTaken == EdgeTypeFalseBranch
	case EdgeTypeStart:
		// Start edges shouldn't appear here (they have nil source)
		return false
	default:
		// Unknown edge type - don't follow
		ae.log.Warn(context.Background(), "Unknown edge type",
			"edgeType", edge.EdgeType,
			"edgeID", edge.ID)
		return false
	}
}

// sortEdgesByOrder sorts edges by their EdgeOrder field for deterministic execution.
func sortEdgesByOrder(edges []ActionEdge) {
	for i := 0; i < len(edges)-1; i++ {
		for j := i + 1; j < len(edges); j++ {
			if edges[i].EdgeOrder > edges[j].EdgeOrder {
				edges[i], edges[j] = edges[j], edges[i]
			}
		}
	}
}

// executeAction executes a single action
func (ae *ActionExecutor) executeAction(ctx context.Context, action RuleActionView, executionContext ActionExecutionContext) ActionResult {
	startTime := time.Now()
	actionID := action.ID

	ae.log.Info(ctx, "VERBOSE: Starting action execution",
		"action_id", actionID,
		"action_name", action.Name,
		"action_type", ae.getActionType(action),
		"rule_id", executionContext.RuleID, // Pointer - may be nil for manual executions
		"rule_name", executionContext.RuleName,
		"entity_name", executionContext.EntityName,
		"entity_id", executionContext.EntityID)

	// Track active execution
	exec := &ActionExecution{
		ActionID:  actionID,
		RuleID:    executionContext.RuleID, // Already a pointer
		StartedAt: startTime,
		Status:    "running",
		Context:   executionContext,
	}
	ae.activeExecs.Store(actionID, exec)
	defer ae.activeExecs.Delete(actionID)

	result := ActionResult{
		ActionID:   actionID,
		ActionName: action.Name,
		ActionType: ae.getActionType(action),
		Status:     "failed",
		StartedAt:  startTime,
	}

	// Validate action configuration
	validation := ae.ValidateActionConfig(action)
	if !validation.IsValid {
		result.ErrorMessage = fmt.Sprintf("Invalid action configuration: %v", validation.Errors)
		result.CompletedAt = timePtr(time.Now())
		result.Duration = time.Since(startTime)
		ae.log.Error(ctx, "VERBOSE: Action validation failed",
			"action_id", actionID,
			"action_name", action.Name,
			"validation_errors", validation.Errors)
		return result
	}

	// Get action configuration (merge template defaults with action config)
	config := ae.mergeActionConfig(action)

	// Process template variables
	templateContext := ae.buildTemplateContext(executionContext)
	processedConfig, err := ae.processTemplates(config, templateContext)
	if err != nil {
		ae.log.Error(ctx, "Failed to process templates",
			"actionID", actionID,
			"error", err)
		// Continue with original config if template processing fails
		processedConfig = config
	}

	// Get handler for action type
	actionType := ae.getActionType(action)
	handler, exists := ae.registry.Get(actionType)
	if !exists {
		result.ErrorMessage = fmt.Sprintf("No handler for action type: %s", actionType)
		result.CompletedAt = timePtr(time.Now())
		result.Duration = time.Since(startTime)
		ae.log.Error(ctx, "VERBOSE: No handler found for action type",
			"action_id", actionID,
			"action_name", action.Name,
			"action_type", actionType)
		return result
	}

	ae.log.Info(ctx, "VERBOSE: Found handler, executing action",
		"action_id", actionID,
		"action_type", actionType,
		"config_length", len(processedConfig))

	// Execute action with retry logic
	var execErr error
	var resultData interface{}

	attempts := 1
	if ae.retryEnabled {
		attempts = ae.maxRetries
	}

	for i := 0; i < attempts; i++ {
		if i > 0 {
			ae.log.Info(ctx, "Retrying action execution",
				"actionID", actionID,
				"attempt", i+1)
			time.Sleep(time.Second * time.Duration(i)) // Exponential backoff
		}

		resultData, execErr = handler.Execute(ctx, processedConfig, executionContext)
		if execErr == nil {
			ae.log.Info(ctx, "VERBOSE: Action executed successfully",
				"action_id", actionID,
				"action_type", actionType,
				"attempt", i+1)
			break
		} else {
			ae.log.Error(ctx, "VERBOSE: Action execution failed",
				"action_id", actionID,
				"action_type", actionType,
				"attempt", i+1,
				"error", execErr.Error())
		}
	}

	// Set result based on execution outcome
	if execErr != nil {
		result.Status = "failed"
		result.ErrorMessage = execErr.Error()
		ae.log.Error(ctx, "VERBOSE: Action final result - FAILED",
			"action_id", actionID,
			"action_name", action.Name,
			"action_type", actionType,
			"error", execErr.Error())
	} else {
		result.Status = "success"

		// Check if this is a condition result and extract BranchTaken
		if condResult, ok := resultData.(ConditionResult); ok {
			result.BranchTaken = condResult.BranchTaken
			result.ResultData = map[string]interface{}{
				"evaluated":    condResult.Evaluated,
				"result":       condResult.Result,
				"branch_taken": condResult.BranchTaken,
			}
		} else if resultData != nil {
			// Type assert to map[string]interface{} if possible
			if data, ok := resultData.(map[string]interface{}); ok {
				result.ResultData = data
			} else {
				// If the handler returns a different type, wrap it in a map
				result.ResultData = map[string]interface{}{
					"data": resultData,
				}
			}
		}
	}

	completedAt := time.Now()
	result.CompletedAt = &completedAt
	result.Duration = completedAt.Sub(startTime)

	return result
}

// loadRuleActions loads actions for a rule from the database
// func (ae *ActionExecutor) loadRuleActions(ctx context.Context, ruleID uuid.UUID) ([]RuleActionView, error) {
// 	query := `
// 		SELECT
// 			ra.id,
// 			ra.automation_rules_id,
// 			ra.name,
// 			ra.description,
// 			ra.action_config,
// 			ra.execution_order,
// 			ra.is_active,
// 			ra.template_id,
// 			at.name as template_name,
// 			at.action_type as template_action_type,
// 			at.default_config as template_default_config
// 		FROM workflow.rule_actions ra
// 		LEFT JOIN workflow.action_templates at ON ra.template_id = at.id
// 		WHERE ra.automation_rules_id = $1
// 		ORDER BY ra.execution_order ASC
// 	`

// 	var actions []RuleActionView
// 	if err := ae.db.SelectContext(ctx, &actions, query, ruleID); err != nil {
// 		return nil, fmt.Errorf("failed to load rule actions: %w", err)
// 	}

// 	return actions, nil
// }

// getRuleName gets the name of a rule
func (ae *ActionExecutor) getRuleName(ctx context.Context, ruleID uuid.UUID) string {
	var name string
	query := `SELECT name FROM workflow.automation_rules WHERE id = $1`

	if err := ae.db.GetContext(ctx, &name, query, ruleID); err != nil {
		ae.log.Error(ctx, "Failed to get rule name",
			"ruleID", ruleID,
			"error", err)
		return "Unknown Rule"
	}

	return name
}

// ValidateActionConfig validates an action's configuration
func (ae *ActionExecutor) ValidateActionConfig(action RuleActionView) ActionValidationResult {
	errors := make([]string, 0)
	warnings := make([]string, 0)

	// Basic validation
	actionType := ae.getActionType(action)
	if actionType == "" {
		errors = append(errors, "Action type is required")
		return ActionValidationResult{
			IsValid:  false,
			Errors:   errors,
			Warnings: warnings,
		}
	}

	// Check if handler exists
	handler, exists := ae.registry.Get(actionType)
	if !exists {
		errors = append(errors, fmt.Sprintf("Unsupported action type: %s", actionType))
		return ActionValidationResult{
			IsValid:  false,
			Errors:   errors,
			Warnings: warnings,
		}
	}

	// Get merged configuration
	config := ae.mergeActionConfig(action)

	// Validate using handler
	if err := handler.Validate(config); err != nil {
		errors = append(errors, err.Error())
	}

	return ActionValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// mergeActionConfig merges template defaults with action-specific config
func (ae *ActionExecutor) mergeActionConfig(action RuleActionView) json.RawMessage {
	var merged map[string]interface{}

	// Start with template defaults if available
	if action.TemplateDefaultConfig != nil {
		if err := json.Unmarshal(action.TemplateDefaultConfig, &merged); err != nil {
			ae.log.Error(context.Background(), "Failed to unmarshal template config",
				"actionID", action.ID,
				"error", err)
			merged = make(map[string]interface{})
		}
	} else {
		merged = make(map[string]interface{})
	}

	// Override with action-specific config
	if action.ActionConfig != nil {
		var actionConfig map[string]interface{}
		if err := json.Unmarshal(action.ActionConfig, &actionConfig); err != nil {
			ae.log.Error(context.Background(), "Failed to unmarshal action config",
				"actionID", action.ID,
				"error", err)
		} else {
			// Merge configurations
			for k, v := range actionConfig {
				merged[k] = v
			}
		}
	}

	// Convert back to JSON
	result, err := json.Marshal(merged)
	if err != nil {
		ae.log.Error(context.Background(), "Failed to marshal merged config",
			"actionID", action.ID,
			"error", err)
		return json.RawMessage("{}")
	}

	return result
}

// processTemplates processes template variables in configuration
func (ae *ActionExecutor) processTemplates(config json.RawMessage, context TemplateContext) (json.RawMessage, error) {
	// Convert config to interface{} for processing
	var configData interface{}
	if err := json.Unmarshal(config, &configData); err != nil {
		return config, fmt.Errorf("failed to unmarshal config for template processing: %w", err)
	}

	// Process templates
	result := ae.templateProc.ProcessTemplateObject(configData, context)

	// Check for errors
	if len(result.Errors) > 0 {
		return config, fmt.Errorf("template processing errors: %v", result.Errors)
	}

	// Convert back to JSON
	processed, err := json.Marshal(result.Processed)
	if err != nil {
		return config, fmt.Errorf("failed to marshal processed config: %w", err)
	}

	return processed, nil
}

// buildTemplateContext builds the template context from execution context
func (ae *ActionExecutor) buildTemplateContext(execContext ActionExecutionContext) TemplateContext {
	context := make(TemplateContext)

	// Core execution context
	context["entity_id"] = execContext.EntityID
	context["entity_name"] = execContext.EntityName
	context["event_type"] = execContext.EventType
	context["timestamp"] = execContext.Timestamp
	context["user_id"] = execContext.UserID
	// Handle pointer-based RuleID (nil for manual executions)
	if execContext.RuleID != nil {
		context["rule_id"] = *execContext.RuleID
	}
	context["rule_name"] = execContext.RuleName

	// Add raw data fields
	if execContext.RawData != nil {
		for k, v := range execContext.RawData {
			context[k] = v
		}
	}

	// Add field changes
	if execContext.FieldChanges != nil {
		context["field_changes"] = execContext.FieldChanges
	}

	return context
}

// getActionType determines the action type from the action view.
// It first checks TemplateActionType (for template-based actions),
// then falls back to extracting action_type from the ActionConfig JSON.
func (ae *ActionExecutor) getActionType(action RuleActionView) string {
	// First try template action type (for template-based actions)
	if action.TemplateActionType != "" {
		return action.TemplateActionType
	}

	// Fall back to extracting from config (for directly-configured actions)
	if len(action.ActionConfig) > 0 {
		var configMap map[string]interface{}
		if err := json.Unmarshal(action.ActionConfig, &configMap); err == nil {
			if actionType, ok := configMap["action_type"].(string); ok && actionType != "" {
				return actionType
			}
		}
	}

	return ""
}

// shouldStopOnFailure determines if execution should stop after this action fails
func (ae *ActionExecutor) shouldStopOnFailure(action RuleActionView) bool {
	// Could be configured per action or action type
	// For now, only stop on seek_approval failures
	actionType := ae.getActionType(action)
	return actionType == "seek_approval"
}

// updateStats updates execution statistics
func (ae *ActionExecutor) updateStats(result BatchExecutionResult) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	ae.stats.TotalActionsExecuted += result.TotalActions
	ae.stats.SuccessfulExecutions += result.SuccessfulActions
	ae.stats.FailedExecutions += result.FailedActions

	// Update average execution time
	if ae.stats.TotalActionsExecuted > 0 {
		totalTime := ae.stats.AverageExecutionTimeMs * float64(ae.stats.TotalActionsExecuted-result.TotalActions)
		totalTime += float64(result.TotalExecutionTime.Milliseconds())
		ae.stats.AverageExecutionTimeMs = totalTime / float64(ae.stats.TotalActionsExecuted)
	}

	now := time.Now()
	ae.stats.LastExecutedAt = &now
}

// GetStats returns current execution statistics
func (ae *ActionExecutor) GetStats() ActionExecutionStats {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	return ae.stats
}

// GetExecutionHistory returns execution history
func (ae *ActionExecutor) GetExecutionHistory(limit int) []BatchExecutionResult {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	if limit <= 0 || limit > len(ae.history) {
		limit = len(ae.history)
	}

	// Return last N items
	start := len(ae.history) - limit
	if start < 0 {
		start = 0
	}

	result := make([]BatchExecutionResult, limit)
	copy(result, ae.history[start:])
	return result
}

// ClearHistory clears execution history
func (ae *ActionExecutor) ClearHistory() {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.history = make([]BatchExecutionResult, 0)
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

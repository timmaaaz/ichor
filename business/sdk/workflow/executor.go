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

// ActionHandler defines the interface for action type handlers
type ActionHandler interface {
	Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
	Validate(config json.RawMessage) error
	GetType() string
}

// ActionExecution represents an active action execution
type ActionExecution struct {
	ActionID  uuid.UUID
	RuleID    uuid.UUID
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
	log *logger.Logger
	db  *sqlx.DB

	// Action handlers registry
	handlers map[string]ActionHandler

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
func NewActionExecutor(log *logger.Logger, db *sqlx.DB) *ActionExecutor {
	ae := &ActionExecutor{
		log:                log,
		db:                 db,
		handlers:           make(map[string]ActionHandler),
		templateProc:       NewTemplateProcessor(DefaultTemplateProcessingOptions()),
		history:            make([]BatchExecutionResult, 0),
		maxParallelActions: 10,
		retryEnabled:       true,
		maxRetries:         3,
	}

	// Register action handlers
	ae.registerHandlers()

	return ae
}

func (ae *ActionExecutor) NewSeekApprovalHandler() *SeekApprovalHandler {
	return &SeekApprovalHandler{log: ae.log, db: ae.db}
}

func (ae *ActionExecutor) NewAllocateInventoryHandler() *AllocateInventoryHandler {
	return &AllocateInventoryHandler{log: ae.log, db: ae.db}
}

func (ae *ActionExecutor) NewCreateAlertHandler() *CreateAlertHandler {
	return &CreateAlertHandler{log: ae.log, db: ae.db}
}

func (ae *ActionExecutor) NewSendEmailHandler() *SendEmailHandler {
	return &SendEmailHandler{log: ae.log, db: ae.db}
}

func (ae *ActionExecutor) NewUpdateFieldHandler() *UpdateFieldHandler {
	return &UpdateFieldHandler{log: ae.log, db: ae.db}
}

func (ae *ActionExecutor) NewSendNotificationHandler() *SendNotificationHandler {
	return &SendNotificationHandler{log: ae.log, db: ae.db}
}

// registerHandlers registers all available action handlers
func (ae *ActionExecutor) registerHandlers() {
	// Register all action type handlers
	ae.RegisterHandler(&SeekApprovalHandler{log: ae.log, db: ae.db})
	ae.RegisterHandler(&AllocateInventoryHandler{log: ae.log, db: ae.db})
	ae.RegisterHandler(&CreateAlertHandler{log: ae.log, db: ae.db})
	ae.RegisterHandler(&SendEmailHandler{log: ae.log, db: ae.db})
	ae.RegisterHandler(&UpdateFieldHandler{log: ae.log, db: ae.db})
	ae.RegisterHandler(&SendNotificationHandler{log: ae.log, db: ae.db})
}

// RegisterHandler registers an action handler
func (ae *ActionExecutor) RegisterHandler(handler ActionHandler) {
	ae.handlers[handler.GetType()] = handler
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
	actions, err := ae.loadRuleActions(ctx, ruleID)
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

// executeAction executes a single action
func (ae *ActionExecutor) executeAction(ctx context.Context, action RuleActionView, executionContext ActionExecutionContext) ActionResult {
	startTime := time.Now()
	actionID := action.ID

	// Track active execution
	exec := &ActionExecution{
		ActionID:  actionID,
		RuleID:    executionContext.RuleID,
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
	handler, exists := ae.handlers[actionType]
	if !exists {
		result.ErrorMessage = fmt.Sprintf("No handler for action type: %s", actionType)
		result.CompletedAt = timePtr(time.Now())
		result.Duration = time.Since(startTime)
		return result
	}

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
			break
		}
	}

	// Set result based on execution outcome
	if execErr != nil {
		result.Status = "failed"
		result.ErrorMessage = execErr.Error()
	} else {
		result.Status = "success"
		// To this:
		if resultData != nil {
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
func (ae *ActionExecutor) loadRuleActions(ctx context.Context, ruleID uuid.UUID) ([]RuleActionView, error) {
	query := `
		SELECT 
			ra.id,
			ra.automation_rules_id,
			ra.name,
			ra.description,
			ra.action_config,
			ra.execution_order,
			ra.is_active,
			ra.template_id,
			at.name as template_name,
			at.action_type as template_action_type,
			at.default_config as template_default_config
		FROM rule_actions ra
		LEFT JOIN action_templates at ON ra.template_id = at.id
		WHERE ra.automation_rules_id = $1
		ORDER BY ra.execution_order ASC
	`

	var actions []RuleActionView
	if err := ae.db.SelectContext(ctx, &actions, query, ruleID); err != nil {
		return nil, fmt.Errorf("failed to load rule actions: %w", err)
	}

	return actions, nil
}

// getRuleName gets the name of a rule
func (ae *ActionExecutor) getRuleName(ctx context.Context, ruleID uuid.UUID) string {
	var name string
	query := `SELECT name FROM automation_rules WHERE id = $1`

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
	handler, exists := ae.handlers[actionType]
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
	context["rule_id"] = execContext.RuleID
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

// getActionType determines the action type from the action view
func (ae *ActionExecutor) getActionType(action RuleActionView) string {
	// TODO: Check if this is necessary after type revamp
	return action.TemplateActionType
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

// Action Handler Implementations (Stubbed)

// SeekApprovalHandler handles seek_approval actions
type SeekApprovalHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

func (h *SeekApprovalHandler) GetType() string {
	return "seek_approval"
}

func (h *SeekApprovalHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Approvers    []string `json:"approvers"`
		ApprovalType string   `json:"approval_type"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.Approvers) == 0 {
		return fmt.Errorf("approvers list is required and must not be empty")
	}

	validTypes := map[string]bool{"any": true, "all": true, "majority": true}
	if !validTypes[cfg.ApprovalType] {
		return fmt.Errorf("invalid approval_type, must be: any, all, or majority")
	}

	return nil
}

func (h *SeekApprovalHandler) Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "STUB: Executing seek_approval action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	// Stub implementation
	result := map[string]interface{}{
		"approval_id":    fmt.Sprintf("approval_%d", time.Now().Unix()),
		"status":         "pending",
		"requested_at":   time.Now().Format(time.RFC3339),
		"reference_id":   context.EntityID,
		"reference_type": fmt.Sprintf("%s_%s", context.EntityName, context.EventType),
	}

	return result, nil
}

// AllocateInventoryHandler handles allocate_inventory actions
type AllocateInventoryHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

func (h *AllocateInventoryHandler) GetType() string {
	return "allocate_inventory"
}

func (h *AllocateInventoryHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		InventoryItems []struct {
			ItemID   string `json:"item_id"`
			Quantity int    `json:"quantity"`
		} `json:"inventory_items"`
		AllocationStrategy string `json:"allocation_strategy"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.InventoryItems) == 0 {
		return fmt.Errorf("inventory_items list is required and must not be empty")
	}

	validStrategies := map[string]bool{
		"fifo": true, "lifo": true, "nearest_expiry": true, "lowest_cost": true,
	}
	if !validStrategies[cfg.AllocationStrategy] {
		return fmt.Errorf("invalid allocation_strategy")
	}

	return nil
}

func (h *AllocateInventoryHandler) Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "STUB: Executing allocate_inventory action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	// Stub implementation
	result := map[string]interface{}{
		"allocation_id": fmt.Sprintf("alloc_%d", time.Now().Unix()),
		"status":        "allocated",
		"allocated_at":  time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// CreateAlertHandler handles create_alert actions
type CreateAlertHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

func (h *CreateAlertHandler) GetType() string {
	return "create_alert"
}

func (h *CreateAlertHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Message    string   `json:"message"`
		Recipients []string `json:"recipients"`
		Priority   string   `json:"priority"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.Message == "" {
		return fmt.Errorf("alert message is required")
	}

	if len(cfg.Recipients) == 0 {
		return fmt.Errorf("recipients list is required and must not be empty")
	}

	validPriorities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validPriorities[cfg.Priority] {
		return fmt.Errorf("invalid priority level")
	}

	return nil
}

func (h *CreateAlertHandler) Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "STUB: Executing create_alert action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	// Stub implementation
	result := map[string]interface{}{
		"alert_id":   fmt.Sprintf("alert_%d", time.Now().Unix()),
		"status":     "created",
		"created_at": time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// SendEmailHandler handles send_email actions
type SendEmailHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

func (h *SendEmailHandler) GetType() string {
	return "send_email"
}

func (h *SendEmailHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Recipients []string `json:"recipients"`
		Subject    string   `json:"subject"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.Recipients) == 0 {
		return fmt.Errorf("email recipients list is required and must not be empty")
	}

	if cfg.Subject == "" {
		return fmt.Errorf("email subject is required")
	}

	return nil
}

func (h *SendEmailHandler) Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "STUB: Executing send_email action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	// Stub implementation
	result := map[string]interface{}{
		"email_id": fmt.Sprintf("email_%d", time.Now().Unix()),
		"status":   "sent",
		"sent_at":  time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// UpdateFieldHandler handles update_field actions
type UpdateFieldHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

func (h *UpdateFieldHandler) GetType() string {
	return "update_field"
}

func (h *UpdateFieldHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		TargetEntity string      `json:"target_entity"`
		TargetField  string      `json:"target_field"`
		NewValue     interface{} `json:"new_value"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.TargetEntity == "" {
		return fmt.Errorf("target entity is required")
	}

	if cfg.TargetField == "" {
		return fmt.Errorf("target field is required")
	}

	if cfg.NewValue == nil {
		return fmt.Errorf("new value is required")
	}

	return nil
}

func (h *UpdateFieldHandler) Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "STUB: Executing update_field action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	// Stub implementation
	result := map[string]interface{}{
		"update_id":  fmt.Sprintf("update_%d", time.Now().Unix()),
		"status":     "updated",
		"updated_at": time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// SendNotificationHandler handles send_notification actions
type SendNotificationHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

func (h *SendNotificationHandler) GetType() string {
	return "send_notification"
}

func (h *SendNotificationHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Recipients []string `json:"recipients"`
		Channels   []struct {
			Type string `json:"type"`
		} `json:"channels"`
		Priority string `json:"priority"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.Recipients) == 0 {
		return fmt.Errorf("recipients list is required and must not be empty")
	}

	if len(cfg.Channels) == 0 {
		return fmt.Errorf("at least one notification channel is required")
	}

	validPriorities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validPriorities[cfg.Priority] {
		return fmt.Errorf("invalid priority level")
	}

	return nil
}

func (h *SendNotificationHandler) Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "STUB: Executing send_notification action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	// Stub implementation
	result := map[string]interface{}{
		"notification_id": fmt.Sprintf("notif_%d", time.Now().Unix()),
		"status":          "sent",
		"sent_at":         time.Now().Format(time.RFC3339),
	}

	return result, nil
}

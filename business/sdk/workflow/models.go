package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TriggerEvent represents an event that triggers workflow execution
type TriggerEvent struct {
	EventType    string                 `json:"event_type"` // on_create, on_update, on_delete
	EntityName   string                 `json:"entity_name"`
	EntityID     string                 `json:"entity_id,omitempty"`
	FieldChanges map[string]FieldChange `json:"field_changes,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	RawData      map[string]interface{} `json:"raw_data,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
}

// FieldChange represents a change in a field value
type FieldChange struct {
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// ExecutionBatch represents a batch of rules that can execute in parallel
type ExecutionBatch struct {
	BatchNumber       int           `json:"batch_number"`
	RuleIDs           []string      `json:"rule_ids"`
	CanRunParallel    bool          `json:"can_run_parallel"`
	EstimatedDuration time.Duration `json:"estimated_duration_ms"`
	DependencyLevel   int           `json:"dependency_level"`
}

// ExecutionPlan represents the planned execution of matched rules
type ExecutionPlan struct {
	PlanID                 string           `json:"plan_id"`
	TriggerEvent           TriggerEvent     `json:"trigger_event"`
	MatchedRuleCount       int              `json:"matched_rule_count"`
	ExecutionBatches       []ExecutionBatch `json:"execution_batches"`
	TotalBatches           int              `json:"total_batches"`
	EstimatedTotalDuration time.Duration    `json:"estimated_total_duration_ms"`
	CreatedAt              time.Time        `json:"created_at"`
}

// WorkflowExecution represents an active or completed workflow execution
type WorkflowExecution struct {
	ExecutionID   string          `json:"execution_id"`
	TriggerEvent  TriggerEvent    `json:"trigger_event"`
	ExecutionPlan ExecutionPlan   `json:"execution_plan"`
	CurrentBatch  int             `json:"current_batch"`
	Status        ExecutionStatus `json:"status"`
	StartedAt     time.Time       `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	TotalDuration *time.Duration  `json:"total_duration_ms,omitempty"`
	BatchResults  []BatchResult   `json:"batch_results"`
	Errors        []string        `json:"errors"`
}

// ExecutionStatus represents the status of a workflow execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
)

// BatchResult represents the result of executing a batch of rules
type BatchResult struct {
	BatchNumber int           `json:"batch_number"`
	RuleResults []RuleResult  `json:"rule_results"`
	BatchStatus string        `json:"batch_status"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Duration    time.Duration `json:"duration_ms"`
}

// RuleResult represents the result of executing a single rule
type RuleResult struct {
	RuleID        string         `json:"rule_id"`
	RuleName      string         `json:"rule_name"`
	Status        string         `json:"status"`
	ActionResults []ActionResult `json:"action_results"`
	StartedAt     time.Time      `json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	Duration      time.Duration  `json:"duration_ms"`
	ErrorMessage  string         `json:"error_message,omitempty"`
}

// ActionResult represents the result of executing a single action
type ActionResult struct {
	ActionID     string                 `json:"action_id"`
	ActionName   string                 `json:"action_name"`
	ActionType   string                 `json:"action_type"`
	Status       string                 `json:"status"`
	ResultData   map[string]interface{} `json:"result_data,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Duration     time.Duration          `json:"duration_ms"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// WorkflowConfig holds configuration for the workflow engine
type WorkflowConfig struct {
	MaxParallelRules      int           `json:"max_parallel_rules"`
	MaxParallelActions    int           `json:"max_parallel_actions"`
	DefaultTimeout        time.Duration `json:"default_timeout_ms"`
	RetryFailedActions    bool          `json:"retry_failed_actions"`
	MaxRetries            int           `json:"max_retries"`
	StopOnCriticalFailure bool          `json:"stop_on_critical_failure"`
}

// WorkflowStats tracks workflow execution statistics
type WorkflowStats struct {
	TotalWorkflowsProcessed int           `json:"total_workflows_processed"`
	SuccessfulWorkflows     int           `json:"successful_workflows"`
	FailedWorkflows         int           `json:"failed_workflows"`
	AverageWorkflowDuration time.Duration `json:"average_workflow_duration_ms"`
	TotalRulesExecuted      int           `json:"total_rules_executed"`
	TotalActionsExecuted    int           `json:"total_actions_executed"`
	LastExecutionAt         *time.Time    `json:"last_execution_at"`
}

// ActionExecutionContext provides context for action execution
type ActionExecutionContext struct {
	EntityID     string                 `json:"entity_id,omitempty"`
	EntityName   string                 `json:"entity_name"`
	EventType    string                 `json:"event_type"`
	FieldChanges map[string]FieldChange `json:"field_changes,omitempty"`
	RawData      map[string]interface{} `json:"raw_data,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	UserID       string                 `json:"user_id,omitempty"`
	RuleID       string                 `json:"rule_id"`
	RuleName     string                 `json:"rule_name"`
	ExecutionID  string                 `json:"execution_id,omitempty"`
}

// =============================================================================
// CRUD MODELS
// =============================================================================

// =============================================================================
// Trigger Type
// =============================================================================
// TriggerType represents types of triggers for automation rules
type TriggerType struct {
	ID          uuid.UUID
	Name        string
	Description string
}

// NewTriggerType contains information needed to create a new trigger type
type NewTriggerType struct {
	Name        string
	Description string
}

// UpdateTriggerType contains information needed to update a trigger type
type UpdateTriggerType struct {
	Name        *string
	Description *string
}

// =============================================================================
// Entity Type
// =============================================================================
// EntityType represents types of entities that can be monitored
type EntityType struct {
	ID          uuid.UUID
	Name        string
	Description string
}

// NewEntityType contains information needed to create a new entity type
type NewEntityType struct {
	Name        string
	Description string
}

// UpdateEntityType contains information needed to update an entity type
type UpdateEntityType struct {
	Name        *string
	Description *string
}

// =============================================================================
// Entity
// =============================================================================
// Entity represents a monitored database entity
type Entity struct {
	ID           uuid.UUID
	Name         string
	EntityTypeID uuid.UUID
	SchemaName   string
	IsActive     bool
	DateCreated  time.Time
}

// NewEntity contains information needed to create a new entity
type NewEntity struct {
	Name         string
	EntityTypeID uuid.UUID
	SchemaName   string
	IsActive     bool
}

// UpdateEntity contains information needed to update an entity
type UpdateEntity struct {
	Name         *string
	EntityTypeID *uuid.UUID
	SchemaName   *string
	IsActive     *bool
}

// =============================================================================
// Automation Rule
// =============================================================================
// AutomationRule represents a workflow automation rule
type AutomationRule struct {
	ID                uuid.UUID
	Name              string
	Description       string
	EntityID          uuid.UUID
	EntityTypeID      uuid.UUID
	TriggerTypeID     uuid.UUID
	TriggerConditions json.RawMessage
	IsActive          bool
	DateCreated       time.Time
	DateUpdated       time.Time
	CreatedBy         uuid.UUID
	UpdatedBy         uuid.UUID
}

// NewAutomationRule contains information needed to create a new automation rule
type NewAutomationRule struct {
	Name              string
	Description       string
	EntityID          uuid.UUID
	EntityTypeID      uuid.UUID
	TriggerTypeID     uuid.UUID
	TriggerConditions json.RawMessage
	IsActive          bool
	CreatedBy         uuid.UUID
}

// UpdateAutomationRule contains information needed to update an automation rule
type UpdateAutomationRule struct {
	Name              *string
	Description       *string
	EntityID          *uuid.UUID
	EntityTypeID      *uuid.UUID
	TriggerTypeID     *uuid.UUID
	TriggerConditions *json.RawMessage
	IsActive          *bool
	UpdatedBy         *uuid.UUID
}

// =============================================================================
// Action Template
// =============================================================================
// ActionTemplate represents a reusable action configuration template
type ActionTemplate struct {
	ID            uuid.UUID
	Name          string
	Description   string
	ActionType    string
	DefaultConfig json.RawMessage
	DateCreated   time.Time
	CreatedBy     uuid.UUID
}

// NewActionTemplate contains information needed to create a new action template
type NewActionTemplate struct {
	Name          string
	Description   string
	ActionType    string
	DefaultConfig json.RawMessage
	CreatedBy     uuid.UUID
}

// UpdateActionTemplate contains information needed to update an action template
type UpdateActionTemplate struct {
	Name          *string
	Description   *string
	ActionType    *string
	DefaultConfig *json.RawMessage
}

// =============================================================================
// Rule Action
// =============================================================================
// RuleAction represents an action within an automation rule
type RuleAction struct {
	ID               uuid.UUID
	AutomationRuleID uuid.UUID
	Name             string
	Description      string
	ActionConfig     json.RawMessage
	ExecutionOrder   int
	IsActive         bool
	TemplateID       *uuid.UUID // Nullable
}

// NewRuleAction contains information needed to create a new rule action
type NewRuleAction struct {
	AutomationRuleID uuid.UUID
	Name             string
	Description      string
	ActionConfig     json.RawMessage
	ExecutionOrder   int
	IsActive         bool
	TemplateID       *uuid.UUID
}

// UpdateRuleAction contains information needed to update a rule action
type UpdateRuleAction struct {
	Name           *string
	Description    *string
	ActionConfig   *json.RawMessage
	ExecutionOrder *int
	IsActive       *bool
	TemplateID     *uuid.UUID
}

// =============================================================================
// Rule Dependency
// =============================================================================
// RuleDependency represents a dependency between two automation rules
type RuleDependency struct {
	ParentRuleID uuid.UUID
	ChildRuleID  uuid.UUID
}

// NewRuleDependency contains information needed to create a new rule dependency
type NewRuleDependency struct {
	ParentRuleID uuid.UUID
	ChildRuleID  uuid.UUID
}

// =============================================================================
// Automation Execution
// =============================================================================
// AutomationExecution represents an execution record of an automation rule
type AutomationExecution struct {
	ID               uuid.UUID
	AutomationRuleID uuid.UUID
	EntityType       string
	TriggerData      json.RawMessage
	ActionsExecuted  json.RawMessage
	Status           ExecutionStatus
	ErrorMessage     string
	ExecutionTimeMs  int
	ExecutedAt       time.Time
}

// NewAutomationExecution contains information needed to record a new execution
type NewAutomationExecution struct {
	AutomationRuleID uuid.UUID
	EntityType       string
	TriggerData      json.RawMessage
	ActionsExecuted  json.RawMessage
	Status           ExecutionStatus
	ErrorMessage     string
	ExecutionTimeMs  int
}

// ActionExecutionStatusExecutionStatus represents the status of an automation execution
type ActionExecutionStatus string

const (
	ActionExecutionStatusSuccess ActionExecutionStatus = "success"
	ActionExecutionStatusFailed  ActionExecutionStatus = "failed"
	ActionExecutionStatusPartial ActionExecutionStatus = "partial"
)

// QueryFilter represents common query filters for listing operations
type QueryFilter struct {
	IsActive  *bool
	EntityID  *uuid.UUID
	RuleID    *uuid.UUID
	CreatedBy *uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
	OrderBy   string
	Limit     int
	Offset    int
}

// AutomationRuleView represents a flattened view of a rule with all related data
type AutomationRuleView struct {
	ID                uuid.UUID
	Name              string
	Description       string
	EntityID          *uuid.UUID
	TriggerConditions json.RawMessage
	Actions           json.RawMessage
	IsActive          bool
	DateCreated       time.Time
	DateUpdated       time.Time
	CreatedBy         uuid.UUID
	UpdatedBy         uuid.UUID
	// Trigger type information
	TriggerTypeID          *uuid.UUID
	TriggerTypeName        string
	TriggerTypeDescription string
	// Entity type information
	EntityTypeID          *uuid.UUID
	EntityTypeName        string
	EntityTypeDescription string
	// Entity information
	EntityName       string
	EntitySchemaName string
}

// RuleActionView represents a flattened view of an action with template info
type RuleActionView struct {
	ID               uuid.UUID
	AutomationRuleID *uuid.UUID
	Name             string
	Description      string
	ActionConfig     json.RawMessage
	ExecutionOrder   int
	IsActive         bool
	TemplateID       *uuid.UUID
	// Template information
	TemplateName          string
	TemplateActionType    string
	TemplateDefaultConfig json.RawMessage
}

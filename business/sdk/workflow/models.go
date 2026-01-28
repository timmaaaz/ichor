package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TODO: Sort *uuid.UUID types. Either nullable uuids must be uuid.Nil or use
// pointers but we cannot mix and match approaches.

// TriggerEvent represents an event that triggers workflow execution
type TriggerEvent struct {
	EventType    string                 `json:"event_type"` // on_create, on_update, on_delete
	EntityName   string                 `json:"entity_name"`
	EntityID     uuid.UUID              `json:"entity_id,omitempty"`
	FieldChanges map[string]FieldChange `json:"field_changes,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	RawData      map[string]any         `json:"raw_data,omitempty"`
	UserID       uuid.UUID              `json:"user_id,omitempty"`
}

// FieldChange represents a change in a field value
type FieldChange struct {
	OldValue any `json:"old_value"`
	NewValue any `json:"new_value"`
}

// ExecutionBatch represents a batch of rules that can execute in parallel
type ExecutionBatch struct {
	BatchNumber       int           `json:"batch_number"`
	RuleIDs           uuid.UUIDs    `json:"rule_ids"`
	CanRunParallel    bool          `json:"can_run_parallel"`
	EstimatedDuration time.Duration `json:"estimated_duration_ms"`
	DependencyLevel   int           `json:"dependency_level"`
}

// ExecutionPlan represents the planned execution of matched rules
type ExecutionPlan struct {
	PlanID                 uuid.UUID        `json:"plan_id"`
	TriggerEvent           TriggerEvent     `json:"trigger_event"`
	MatchedRuleCount       int              `json:"matched_rule_count"`
	ExecutionBatches       []ExecutionBatch `json:"execution_batches"`
	TotalBatches           int              `json:"total_batches"`
	EstimatedTotalDuration time.Duration    `json:"estimated_total_duration_ms"`
	CreatedAt              time.Time        `json:"created_at"`
}

// WorkflowExecution represents an active or completed workflow execution
type WorkflowExecution struct {
	ExecutionID   uuid.UUID       `json:"execution_id"`
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

// TriggerSource constants for distinguishing automated vs manual executions
const (
	TriggerSourceAutomation = "automation"
	TriggerSourceManual     = "manual"
)

// EventType constants for action execution
const (
	EventTypeOnCreate      = "on_create"
	EventTypeOnUpdate      = "on_update"
	EventTypeOnDelete      = "on_delete"
	EventTypeManualTrigger = "manual_trigger"
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
	RuleID        uuid.UUID      `json:"rule_id"`
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
	ActionID     uuid.UUID              `json:"action_id"`
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

// ActionExecutionContext provides context for action execution.
// It supports both automated workflow execution and manual action execution.
type ActionExecutionContext struct {
	EntityID      uuid.UUID              `json:"entity_id,omitempty"`
	EntityName    string                 `json:"entity_name"`
	EventType     string                 `json:"event_type"` // "on_create", "on_update", "on_delete", "manual_trigger"
	FieldChanges  map[string]FieldChange `json:"field_changes,omitempty"`
	RawData       map[string]interface{} `json:"raw_data,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	UserID        uuid.UUID              `json:"user_id,omitempty"`
	RuleID        *uuid.UUID             `json:"rule_id,omitempty"`      // Pointer: nil for manual executions
	RuleName      string                 `json:"rule_name"`
	ExecutionID   uuid.UUID              `json:"execution_id,omitempty"`
	TriggerSource string                 `json:"trigger_source"` // "automation" or "manual"
}

// =============================================================================
// CRUD MODELS
// =============================================================================

// =============================================================================
// Trigger Type
// =============================================================================
// TriggerType represents types of triggers for automation rules
type TriggerType struct {
	ID            uuid.UUID
	Name          string
	Description   string
	IsActive      bool
	DeactivatedBy uuid.UUID
}

// NewTriggerType contains information needed to create a new trigger type
type NewTriggerType struct {
	Name          string
	Description   string
	IsActive      bool
	DeactivatedBy uuid.UUID
}

// UpdateTriggerType contains information needed to update a trigger type
type UpdateTriggerType struct {
	Name        *string
	Description *string
	IsActive    *bool
	// DeactivatedBy *uuid.UUID // Specific endpoint for deactivating trigger types
}

// =============================================================================
// Entity Type
// =============================================================================
// EntityType represents types of entities that can be monitored
type EntityType struct {
	ID            uuid.UUID
	Name          string
	Description   string
	IsActive      bool
	DeactivatedBy uuid.UUID
}

// NewEntityType contains information needed to create a new entity type
type NewEntityType struct {
	Name        string
	Description string
	IsActive    bool
}

// UpdateEntityType contains information needed to update an entity type
type UpdateEntityType struct {
	Name        *string
	Description *string
	IsActive    *bool
	// DeactivatedBy *uuid.UUID // Specific endpoint for deactivating entity types
}

// =============================================================================
// Entity
// =============================================================================
// Entity represents a monitored database entity
type Entity struct {
	ID            uuid.UUID
	Name          string
	EntityTypeID  uuid.UUID
	SchemaName    string
	IsActive      bool
	CreatedDate   time.Time
	DeactivatedBy uuid.UUID
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
	TriggerConditions *json.RawMessage
	IsActive          bool
	CreatedDate       time.Time
	UpdatedDate       time.Time
	CreatedBy         uuid.UUID
	UpdatedBy         uuid.UUID
	DeactivatedBy     uuid.UUID
}

// NewAutomationRule contains information needed to create a new automation rule
type NewAutomationRule struct {
	Name              string
	Description       string
	EntityID          uuid.UUID
	EntityTypeID      uuid.UUID
	TriggerTypeID     uuid.UUID
	TriggerConditions *json.RawMessage
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
	CreatedDate   time.Time
	CreatedBy     uuid.UUID
	IsActive      bool
	DeactivatedBy uuid.UUID
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
	DeactivatedBy    uuid.UUID  // Nullable, tracks who deactivated the action
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
	ID           uuid.UUID
	ParentRuleID uuid.UUID
	ChildRuleID  uuid.UUID
}

// NewRuleDependency contains information needed to create a new rule dependency
type NewRuleDependency struct {
	ID           uuid.UUID
	ParentRuleID uuid.UUID
	ChildRuleID  uuid.UUID
}

// =============================================================================
// Automation Execution
// =============================================================================
// AutomationExecution represents an execution record of an automation rule or manual action
type AutomationExecution struct {
	ID               uuid.UUID
	AutomationRuleID *uuid.UUID // Pointer: nil for manual executions
	EntityType       string
	TriggerData      json.RawMessage
	ActionsExecuted  json.RawMessage
	Status           ExecutionStatus
	ErrorMessage     string
	ExecutionTimeMs  int
	ExecutedAt       time.Time
	TriggerSource    string     // "automation" or "manual"
	ExecutedBy       *uuid.UUID // User who triggered manual execution
	ActionType       string     // For manual executions: the action type that was executed
}

// NewAutomationExecution contains information needed to record a new execution
type NewAutomationExecution struct {
	AutomationRuleID *uuid.UUID // Pointer: nil for manual executions
	EntityType       string
	TriggerData      json.RawMessage
	ActionsExecuted  json.RawMessage
	Status           ExecutionStatus
	ErrorMessage     string
	ExecutionTimeMs  int
	TriggerSource    string     // "automation" or "manual"
	ExecutedBy       *uuid.UUID // User who triggered manual execution
	ActionType       string     // For manual executions: the action type
}

// =============================================================================
// Allocation Results
// =============================================================================

// AllocationResult represents the result of an inventory allocation action
// TODO: Figure out how to work this with the one that exists in the allocate.go
// file
type AllocationResult struct {
	ID             uuid.UUID
	IdempotencyKey string
	AllocationData []byte
	CreatedDate    time.Time
}

type NewAllocationResult struct {
	ID             uuid.UUID
	IdempotencyKey string
	AllocationData []byte
}

// ActionExecutionStatusExecutionStatus represents the status of an automation execution
type ActionExecutionStatus string

const (
	ActionExecutionStatusSuccess ActionExecutionStatus = "success"
	ActionExecutionStatusFailed  ActionExecutionStatus = "failed"
	ActionExecutionStatusPartial ActionExecutionStatus = "partial"
)

// DeliveryStatus represents the status of a notification delivery
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusBounced   DeliveryStatus = "bounced"
	DeliveryStatusRetrying  DeliveryStatus = "retrying"
)

// NotificationDelivery represents a delivery attempt for a notification
type NotificationDelivery struct {
	ID                    uuid.UUID       `json:"id"`
	NotificationID        uuid.UUID       `json:"notification_id"`
	AutomationExecutionID uuid.UUID       `json:"workflow_id,omitempty"`
	RuleID                uuid.UUID       `json:"rule_id,omitempty"`
	ActionID              uuid.UUID       `json:"action_id,omitempty"`
	RecipientID           uuid.UUID       `json:"recipient_id"`
	Channel               string          `json:"channel"`
	Status                DeliveryStatus  `json:"status"`
	Attempts              int             `json:"attempts"`
	SentAt                *time.Time      `json:"sent_at,omitempty"`
	DeliveredAt           *time.Time      `json:"delivered_at,omitempty"`
	FailedAt              *time.Time      `json:"failed_at,omitempty"`
	ErrorMessage          *string         `json:"error_message,omitempty"`
	ProviderResponse      json.RawMessage `json:"provider_response,omitempty"`
	CreatedDate           time.Time       `json:"created_date"`
	UpdatedDate           time.Time       `json:"updated_date"`
}

// TODO: Double check this, I think there's a good chance we don't want to update these
// UpdateNotificationDelivery contains information needed to update a notification delivery
type UpdateNotificationDelivery struct {
	NotificationID        *uuid.UUID
	AutomationExecutionID *uuid.UUID
	RuleID                *uuid.UUID
	ActionID              *uuid.UUID
	RecipientID           *uuid.UUID
	Channel               *string
	Status                *DeliveryStatus
	Attempts              *int
	SentAt                *time.Time
	DeliveredAt           *time.Time
	FailedAt              *time.Time
	ErrorMessage          *string
	ProviderResponse      *json.RawMessage
	UpdatedDate           *time.Time
}

// NewNotificationDelivery contains information needed to create a new notification delivery
type NewNotificationDelivery struct {
	NotificationID        uuid.UUID
	AutomationExecutionID uuid.UUID
	RuleID                uuid.UUID
	ActionID              uuid.UUID
	RecipientID           uuid.UUID
	Channel               string
	Status                DeliveryStatus
	Attempts              int
	SentAt                *time.Time
	DeliveredAt           *time.Time
	FailedAt              *time.Time
	ErrorMessage          *string
	ProviderResponse      json.RawMessage
	CreatedDate           time.Time
	UpdatedDate           time.Time
}

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
	Description       *string
	EntityID          *uuid.UUID
	TriggerConditions *json.RawMessage
	Actions           json.RawMessage
	IsActive          bool
	CreatedDate       time.Time
	UpdatedDate       time.Time
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
	ID                uuid.UUID
	AutomationRulesID *uuid.UUID
	Name              string
	Description       string
	ActionConfig      json.RawMessage
	ExecutionOrder    int
	IsActive          bool
	TemplateID        *uuid.UUID
	// Template information
	TemplateName          string
	TemplateActionType    string
	TemplateDefaultConfig json.RawMessage
	DeactivatedBy         uuid.UUID
}

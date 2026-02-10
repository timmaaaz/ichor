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
	CanvasLayout      json.RawMessage
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
	CanvasLayout      json.RawMessage
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
	CanvasLayout      *json.RawMessage
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
	IsActive         bool
	TemplateID       *uuid.UUID
}

// UpdateRuleAction contains information needed to update a rule action
type UpdateRuleAction struct {
	Name         *string
	Description  *string
	ActionConfig *json.RawMessage
	IsActive     *bool
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
// Action Edge (for workflow branching/condition nodes)
// =============================================================================

// EdgeType constants for action edges
const (
	EdgeTypeStart       = "start"        // Entry point into action graph (source is nil)
	EdgeTypeSequence    = "sequence"     // Linear progression to next action
	EdgeTypeTrueBranch  = "true_branch"  // Condition evaluated to true
	EdgeTypeFalseBranch = "false_branch" // Condition evaluated to false
	EdgeTypeAlways      = "always"       // Unconditional edge (always follow)
)

// ActionEdge represents a directed edge between actions in a workflow graph.
// Used to support branching (condition nodes) within a rule's action sequence.
type ActionEdge struct {
	ID             uuid.UUID
	RuleID         uuid.UUID
	SourceActionID *uuid.UUID // nil for start edges (entry points)
	TargetActionID uuid.UUID
	EdgeType       string // start, sequence, true_branch, false_branch, always
	EdgeOrder      int    // For deterministic traversal when multiple edges share same source
	CreatedDate    time.Time
}

// NewActionEdge contains information needed to create a new action edge
type NewActionEdge struct {
	RuleID         uuid.UUID
	SourceActionID *uuid.UUID
	TargetActionID uuid.UUID
	EdgeType       string
	EdgeOrder      int
}

// ConditionResult represents the result of evaluating a condition action.
// Returned by the evaluate_condition action handler.
type ConditionResult struct {
	Evaluated   bool   `json:"evaluated"`    // Whether evaluation was performed
	Result      bool   `json:"result"`       // The boolean result of the condition
	BranchTaken string `json:"branch_taken"` // "true_branch" or "false_branch"
}

// =============================================================================
// Automation Execution
// =============================================================================
// AutomationExecution represents an execution record of an automation rule or manual action
type AutomationExecution struct {
	ID               uuid.UUID
	AutomationRuleID *uuid.UUID // Pointer: nil for manual executions
	RuleName         string     // Rule name from LEFT JOIN (empty for manual executions)
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
	CanvasLayout      json.RawMessage
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
	IsActive          bool
	TemplateID        *uuid.UUID
	// Template information
	TemplateName          string
	TemplateActionType    string
	TemplateDefaultConfig json.RawMessage
	DeactivatedBy         uuid.UUID
}

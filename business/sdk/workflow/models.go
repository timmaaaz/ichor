package workflow

import (
	"time"
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

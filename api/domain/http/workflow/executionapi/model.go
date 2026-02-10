// Package executionapi provides HTTP layer models for workflow execution history endpoints.
// These endpoints provide read-only access to execution records for debugging and auditing.
package executionapi

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ============================================================
// Response Types
// ============================================================

// ExecutionResponse is the response for a single execution in a list.
type ExecutionResponse struct {
	ID              uuid.UUID       `json:"id"`
	AutomationRuleID *uuid.UUID     `json:"automation_rule_id,omitempty"`
	EntityType      string          `json:"entity_type"`
	Status          string          `json:"status"`
	ErrorMessage    string          `json:"error_message,omitempty"`
	ExecutionTimeMs int             `json:"execution_time_ms"`
	ExecutedAt      time.Time       `json:"executed_at"`
	TriggerSource   string          `json:"trigger_source"`
	ExecutedBy      *uuid.UUID      `json:"executed_by,omitempty"`
	ActionType      string          `json:"action_type,omitempty"`
}

// Encode implements web.Encoder for ExecutionResponse.
func (e ExecutionResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(e)
	return data, "application/json", err
}

// ExecutionDetail is the detailed response for a single execution (includes action results).
type ExecutionDetail struct {
	ID               uuid.UUID             `json:"id"`
	AutomationRuleID *uuid.UUID            `json:"automation_rule_id,omitempty"`
	RuleName         string                `json:"rule_name,omitempty"` // Human-readable rule name
	EntityType       string                `json:"entity_type"`
	TriggerData      json.RawMessage       `json:"trigger_data"`
	Status           string                `json:"status"`
	ErrorMessage     string                `json:"error_message,omitempty"`
	ExecutionTimeMs  int                   `json:"execution_time_ms"`
	ExecutedAt       time.Time             `json:"executed_at"`
	TriggerSource    string                `json:"trigger_source"`
	ExecutedBy       *uuid.UUID            `json:"executed_by,omitempty"`
	ActionType       string                `json:"action_type,omitempty"`
	ActionResults    []ActionResultDetail  `json:"action_results"`
}

// Encode implements web.Encoder for ExecutionDetail.
func (e ExecutionDetail) Encode() ([]byte, string, error) {
	data, err := json.Marshal(e)
	return data, "application/json", err
}

// ActionResultDetail represents a single action's result from execution.
type ActionResultDetail struct {
	ActionID     uuid.UUID              `json:"action_id"`
	ActionName   string                 `json:"action_name"`
	ActionType   string                 `json:"action_type"`
	Status       string                 `json:"status"`
	ResultData   map[string]interface{} `json:"result_data,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	DurationMs   int64                  `json:"duration_ms"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// ExecutionList wraps a slice of executions for JSON encoding.
type ExecutionList []ExecutionResponse

// Encode implements web.Encoder for ExecutionList.
func (l ExecutionList) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	return data, "application/json", err
}

// ============================================================
// Converter Functions
// ============================================================

// toExecutionResponse converts a business execution to an API response.
func toExecutionResponse(exec workflow.AutomationExecution) ExecutionResponse {
	return ExecutionResponse{
		ID:               exec.ID,
		AutomationRuleID: exec.AutomationRuleID,
		EntityType:       exec.EntityType,
		Status:           string(exec.Status),
		ErrorMessage:     exec.ErrorMessage,
		ExecutionTimeMs:  exec.ExecutionTimeMs,
		ExecutedAt:       exec.ExecutedAt,
		TriggerSource:    exec.TriggerSource,
		ExecutedBy:       exec.ExecutedBy,
		ActionType:       exec.ActionType,
	}
}

// toExecutionResponses converts a slice of business executions to API responses.
func toExecutionResponses(executions []workflow.AutomationExecution) ExecutionList {
	resp := make(ExecutionList, len(executions))
	for i, exec := range executions {
		resp[i] = toExecutionResponse(exec)
	}
	return resp
}

// toExecutionDetail converts a business execution to a detailed API response.
func toExecutionDetail(exec workflow.AutomationExecution) ExecutionDetail {
	return ExecutionDetail{
		ID:               exec.ID,
		AutomationRuleID: exec.AutomationRuleID,
		RuleName:         exec.RuleName,
		EntityType:       exec.EntityType,
		TriggerData:      exec.TriggerData,
		Status:           string(exec.Status),
		ErrorMessage:     exec.ErrorMessage,
		ExecutionTimeMs:  exec.ExecutionTimeMs,
		ExecutedAt:       exec.ExecutedAt,
		TriggerSource:    exec.TriggerSource,
		ExecutedBy:       exec.ExecutedBy,
		ActionType:       exec.ActionType,
		ActionResults:    parseActionResults(exec.ActionsExecuted),
	}
}

// parseActionResults extracts action results from the actions_executed JSONB column.
func parseActionResults(actionsExecuted json.RawMessage) []ActionResultDetail {
	if len(actionsExecuted) == 0 {
		return []ActionResultDetail{}
	}

	var results []ActionResultDetail
	if err := json.Unmarshal(actionsExecuted, &results); err != nil {
		// If parsing fails, return empty array (JSONB may be in unexpected format)
		return []ActionResultDetail{}
	}
	return results
}

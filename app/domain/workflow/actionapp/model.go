// Package actionapp provides the application layer for manual action execution.
// It handles request/response models, validation, and conversion between
// API and business layer types.
package actionapp

import (
	"encoding/json"
	"time"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Execute Request/Response
// =============================================================================

// ExecuteRequest represents a request to manually execute a workflow action.
type ExecuteRequest struct {
	Config     json.RawMessage        `json:"config" validate:"required"`
	EntityID   *string                `json:"entity_id,omitempty"`
	EntityName string                 `json:"entity_name,omitempty"`
	RawData    map[string]interface{} `json:"raw_data,omitempty"`
}

// Decode implements the decoder interface.
func (app *ExecuteRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app ExecuteRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// ExecuteResponse represents the result of a manual action execution.
type ExecuteResponse struct {
	ExecutionID string      `json:"execution_id"`
	Status      string      `json:"status"` // "completed", "queued", "failed"
	Result      interface{} `json:"result,omitempty"`
	TrackingURL string      `json:"tracking_url,omitempty"`
	Error       string      `json:"error,omitempty"`
}

// Encode implements the encoder interface.
func (app ExecuteResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// toAppExecuteResponse converts a workflow ExecuteResult to an app ExecuteResponse.
func toAppExecuteResponse(bus *workflow.ExecuteResult) ExecuteResponse {
	return ExecuteResponse{
		ExecutionID: bus.ExecutionID.String(),
		Status:      bus.Status,
		Result:      bus.Result,
		TrackingURL: bus.TrackingURL,
		Error:       bus.Error,
	}
}

// =============================================================================
// Available Actions (Discovery)
// =============================================================================

// AvailableAction represents metadata about an action that can be executed manually.
type AvailableAction struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	IsAsync     bool   `json:"is_async"`
}

// AvailableActions is a collection wrapper that implements the Encoder interface.
type AvailableActions []AvailableAction

// Encode implements the encoder interface.
func (app AvailableActions) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// toAppAvailableAction converts a workflow ActionInfo to an app AvailableAction.
func toAppAvailableAction(bus workflow.ActionInfo) AvailableAction {
	return AvailableAction{
		Type:        bus.Type,
		Description: bus.Description,
		IsAsync:     bus.IsAsync,
	}
}

// toAppAvailableActions converts a slice of workflow ActionInfo to AvailableActions.
func toAppAvailableActions(bus []workflow.ActionInfo) AvailableActions {
	app := make(AvailableActions, len(bus))
	for i, v := range bus {
		app[i] = toAppAvailableAction(v)
	}
	return app
}

// =============================================================================
// Execution Status (Tracking)
// =============================================================================

// ExecutionStatus represents the status of an action execution.
type ExecutionStatus struct {
	ExecutionID string      `json:"execution_id"`
	ActionType  string      `json:"action_type"`
	Status      string      `json:"status"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	StartedAt   string      `json:"started_at"`
	CompletedAt string      `json:"completed_at,omitempty"`
}

// Encode implements the encoder interface.
func (app ExecutionStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// toAppExecutionStatus converts a workflow ExecutionStatusInfo to an app ExecutionStatus.
func toAppExecutionStatus(bus *workflow.ExecutionStatusInfo) ExecutionStatus {
	app := ExecutionStatus{
		ExecutionID: bus.ExecutionID.String(),
		ActionType:  bus.ActionType,
		Status:      bus.Status,
		Result:      bus.Result,
		Error:       bus.Error,
		StartedAt:   bus.StartedAt.Format(time.RFC3339),
	}

	if bus.CompletedAt != nil {
		app.CompletedAt = bus.CompletedAt.Format(time.RFC3339)
	}

	return app
}

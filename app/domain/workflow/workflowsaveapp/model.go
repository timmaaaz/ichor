// Package workflowsaveapp provides the application layer for transactional workflow save operations.
package workflowsaveapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// SaveWorkflowRequest represents a complete workflow save request including
// the rule, all actions, and all edges.
type SaveWorkflowRequest struct {
	Name              string              `json:"name" validate:"required,min=1,max=255"`
	Description       string              `json:"description" validate:"max=1000"`
	IsActive          bool                `json:"is_active"`
	EntityID          string              `json:"entity_id" validate:"required,uuid"`
	TriggerTypeID     string              `json:"trigger_type_id" validate:"required,uuid"`
	TriggerConditions json.RawMessage `json:"trigger_conditions"`

	// Actions can be empty for draft workflows (trigger-only, no actions yet).
	// When present, edges must also be provided; see ValidateGraph.
	Actions []SaveActionRequest `json:"actions" validate:"dive"`
	Edges   []SaveEdgeRequest   `json:"edges" validate:"dive"`
	CanvasLayout      json.RawMessage     `json:"canvas_layout"`
}

// Decode implements the Decoder interface.
func (r *SaveWorkflowRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate checks the SaveWorkflowRequest for validity.
func (r SaveWorkflowRequest) Validate() error {
	if err := errs.Check(r); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// SaveActionRequest represents an action to save within a workflow.
// If ID is nil or empty, a new action will be created.
// If ID contains a UUID, the existing action will be updated.
type SaveActionRequest struct {
	ID             *string         `json:"id"`
	Name           string          `json:"name" validate:"required,min=1,max=255"`
	Description    string          `json:"description" validate:"max=1000"`
	ActionType     string          `json:"action_type" validate:"required,oneof=allocate_inventory check_inventory check_reorder_point commit_allocation create_alert create_entity delay evaluate_condition log_audit_entry lookup_entity release_reservation reserve_inventory seek_approval send_email send_notification transition_status update_field"`
	ActionConfig   json.RawMessage `json:"action_config" validate:"required"`
	IsActive       bool            `json:"is_active"`
}

// SaveEdgeRequest represents an edge between actions in the workflow graph.
// SourceActionID can be empty for start edges.
// TargetActionID uses "temp:N" format to reference new actions by array index.
type SaveEdgeRequest struct {
	SourceActionID string `json:"source_action_id"`
	TargetActionID string `json:"target_action_id" validate:"required"`
	EdgeType       string `json:"edge_type" validate:"required,oneof=start sequence always"`
	SourceOutput   string `json:"source_output,omitempty"`
	EdgeOrder      int    `json:"edge_order" validate:"min=0"`
}

// SaveWorkflowResponse represents the complete saved workflow.
type SaveWorkflowResponse struct {
	ID                string               `json:"id"`
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	IsActive          bool                 `json:"is_active"`
	EntityID          string               `json:"entity_id"`
	TriggerTypeID     string               `json:"trigger_type_id"`
	TriggerConditions json.RawMessage      `json:"trigger_conditions"`
	Actions           []SaveActionResponse `json:"actions"`
	Edges             []SaveEdgeResponse   `json:"edges"`
	CanvasLayout      json.RawMessage      `json:"canvas_layout"`
	CreatedDate       string               `json:"created_date"`
	UpdatedDate       string               `json:"updated_date"`
}

// Encode implements the Encoder interface.
func (r SaveWorkflowResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// SaveActionResponse represents a saved action in the response.
type SaveActionResponse struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	ActionType     string          `json:"action_type"`
	ActionConfig   json.RawMessage `json:"action_config"`
	IsActive       bool            `json:"is_active"`
}

// SaveEdgeResponse represents a saved edge in the response.
type SaveEdgeResponse struct {
	ID             string `json:"id"`
	SourceActionID string `json:"source_action_id"`
	TargetActionID string `json:"target_action_id"`
	EdgeType       string `json:"edge_type"`
	SourceOutput   string `json:"source_output,omitempty"`
	EdgeOrder      int    `json:"edge_order"`
}

// ValidationResult is returned for dry-run requests. It contains the
// validation outcome without committing any changes to the database.
type ValidationResult struct {
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors,omitempty"`
	ActionCount int      `json:"action_count"`
	EdgeCount   int      `json:"edge_count"`
}

// Encode implements the Encoder interface.
func (r ValidationResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

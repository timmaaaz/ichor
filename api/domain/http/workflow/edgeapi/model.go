// Package edgeapi provides HTTP handlers for workflow action edge CRUD operations.
// Action edges define the directed graph structure for workflow branching,
// enabling condition nodes and complex action sequences within automation rules.
package edgeapi

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ============================================================
// Request Types
// ============================================================

// CreateEdgeRequest is the request body for creating an action edge.
type CreateEdgeRequest struct {
	SourceActionID *uuid.UUID `json:"source_action_id,omitempty"` // nil for start edges
	TargetActionID uuid.UUID  `json:"target_action_id"`
	EdgeType       string     `json:"edge_type"` // start, sequence, true_branch, false_branch, always
	EdgeOrder      int        `json:"edge_order"`
}

// Decode implements the web.Decoder interface.
func (r *CreateEdgeRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate validates the create edge request.
func (r CreateEdgeRequest) Validate() *ValidationErrors {
	var errs []ValidationError

	// Validate edge type
	validTypes := map[string]bool{
		workflow.EdgeTypeStart:       true,
		workflow.EdgeTypeSequence:    true,
		workflow.EdgeTypeTrueBranch:  true,
		workflow.EdgeTypeFalseBranch: true,
		workflow.EdgeTypeAlways:      true,
	}
	if !validTypes[r.EdgeType] {
		errs = append(errs, ValidationError{
			Field:   "edge_type",
			Message: "edge_type must be one of: start, sequence, true_branch, false_branch, always",
		})
	}

	// Start edges must not have a source action
	if r.EdgeType == workflow.EdgeTypeStart && r.SourceActionID != nil {
		errs = append(errs, ValidationError{
			Field:   "source_action_id",
			Message: "start edges must not have a source_action_id",
		})
	}

	// Non-start edges must have a source action
	if r.EdgeType != workflow.EdgeTypeStart && r.SourceActionID == nil {
		errs = append(errs, ValidationError{
			Field:   "source_action_id",
			Message: "non-start edges must have a source_action_id",
		})
	}

	// Target action is always required
	if r.TargetActionID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field:   "target_action_id",
			Message: "target_action_id is required",
		})
	}

	if len(errs) > 0 {
		return &ValidationErrors{Errors: errs}
	}
	return nil
}

// ============================================================
// Response Types
// ============================================================

// EdgeResponse is the response for a single action edge.
type EdgeResponse struct {
	ID             uuid.UUID  `json:"id"`
	RuleID         uuid.UUID  `json:"rule_id"`
	SourceActionID *uuid.UUID `json:"source_action_id,omitempty"`
	TargetActionID uuid.UUID  `json:"target_action_id"`
	EdgeType       string     `json:"edge_type"`
	EdgeOrder      int        `json:"edge_order"`
	CreatedDate    time.Time  `json:"created_date"`
}

// Encode implements web.Encoder for EdgeResponse.
func (r EdgeResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// EdgeList wraps a slice of edges for JSON encoding.
type EdgeList []EdgeResponse

// Encode implements web.Encoder for EdgeList.
func (l EdgeList) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	return data, "application/json", err
}

// ============================================================
// Validation Types
// ============================================================

// ValidationError represents a single validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors wraps multiple validation errors.
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface.
func (v ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}
	return v.Errors[0].Message
}

// ============================================================
// Converters
// ============================================================

// toEdgeResponse converts a business layer ActionEdge to an API response.
func toEdgeResponse(edge workflow.ActionEdge) EdgeResponse {
	return EdgeResponse{
		ID:             edge.ID,
		RuleID:         edge.RuleID,
		SourceActionID: edge.SourceActionID,
		TargetActionID: edge.TargetActionID,
		EdgeType:       edge.EdgeType,
		EdgeOrder:      edge.EdgeOrder,
		CreatedDate:    edge.CreatedDate,
	}
}

// toEdgeResponses converts a slice of business layer ActionEdges to API responses.
func toEdgeResponses(edges []workflow.ActionEdge) EdgeList {
	resp := make(EdgeList, len(edges))
	for i, edge := range edges {
		resp[i] = toEdgeResponse(edge)
	}
	return resp
}

// toNewActionEdge converts an API request to a business layer NewActionEdge.
func toNewActionEdge(ruleID uuid.UUID, req CreateEdgeRequest) workflow.NewActionEdge {
	return workflow.NewActionEdge{
		RuleID:         ruleID,
		SourceActionID: req.SourceActionID,
		TargetActionID: req.TargetActionID,
		EdgeType:       req.EdgeType,
		EdgeOrder:      req.EdgeOrder,
	}
}

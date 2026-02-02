// Package ruleapi provides HTTP layer models and utilities for workflow automation
// rule CRUD operations. This package follows the HTTP → Business pattern where
// HTTP handlers call the business layer (workflow.Business) directly without an
// intermediate app layer.
//
// The package provides:
//   - Request/response models (CreateRuleRequest, RuleResponse)
//   - Filter parsing from query parameters to workflow.AutomationRuleFilter
//   - Validation utilities for request bodies
//   - Model converters between HTTP and business layer types
//
// All response types implement web.Encoder for JSON serialization.
// All request types implement web.Decoder for JSON deserialization.
package ruleapi

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// Request Types
// ============================================================

// CreateRuleRequest is the request body for creating a rule.
type CreateRuleRequest struct {
	Name              string              `json:"name"`
	Description       string              `json:"description"`
	EntityID          uuid.UUID           `json:"entity_id"`
	EntityTypeID      uuid.UUID           `json:"entity_type_id"`
	TriggerTypeID     uuid.UUID           `json:"trigger_type_id"`
	TriggerConditions json.RawMessage     `json:"trigger_conditions"`
	IsActive          bool                `json:"is_active"`
	Actions           []CreateActionInput `json:"actions,omitempty"`
}

// Decode implements the web.Decoder interface.
func (r *CreateRuleRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// CreateActionInput is an action embedded in a CreateRuleRequest.
type CreateActionInput struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	ActionConfig   json.RawMessage `json:"action_config"`
	ExecutionOrder int             `json:"execution_order"`
	IsActive       bool            `json:"is_active"`
	TemplateID     *uuid.UUID      `json:"template_id,omitempty"`
}

// UpdateRuleRequest is the request body for updating a rule (partial update).
type UpdateRuleRequest struct {
	Name              *string          `json:"name,omitempty"`
	Description       *string          `json:"description,omitempty"`
	EntityID          *uuid.UUID       `json:"entity_id,omitempty"`
	EntityTypeID      *uuid.UUID       `json:"entity_type_id,omitempty"`
	TriggerTypeID     *uuid.UUID       `json:"trigger_type_id,omitempty"`
	TriggerConditions *json.RawMessage `json:"trigger_conditions,omitempty"`
	IsActive          *bool            `json:"is_active,omitempty"`
}

// Decode implements the web.Decoder interface.
func (r *UpdateRuleRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// ToggleActiveRequest is the request body for toggling active status.
type ToggleActiveRequest struct {
	IsActive bool `json:"is_active"`
}

// Decode implements the web.Decoder interface.
func (r *ToggleActiveRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// ============================================================
// Response Types
// ============================================================

// RuleResponse is the response for a single rule.
// Maps from workflow.AutomationRuleView
type RuleResponse struct {
	ID                uuid.UUID        `json:"id"`
	Name              string           `json:"name"`
	Description       string           `json:"description"`
	EntityID          uuid.UUID        `json:"entity_id"`
	EntityName        string           `json:"entity_name"`
	EntitySchemaName  string           `json:"entity_schema_name"`
	EntityTypeID      uuid.UUID        `json:"entity_type_id"`
	EntityTypeName    string           `json:"entity_type_name"`
	TriggerTypeID     uuid.UUID        `json:"trigger_type_id"`
	TriggerTypeName   string           `json:"trigger_type_name"`
	TriggerConditions json.RawMessage  `json:"trigger_conditions"`
	IsActive          bool             `json:"is_active"`
	CreatedBy         uuid.UUID        `json:"created_by"`
	UpdatedBy         uuid.UUID        `json:"updated_by"`
	CreatedDate       time.Time        `json:"created_date"`
	UpdatedDate       time.Time        `json:"updated_date"`
	Actions           []ActionResponse `json:"actions,omitempty"`
}

// Encode implements web.Encoder for RuleResponse.
func (r RuleResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// ActionResponse is the response for a single action.
// Maps from workflow.RuleActionView
type ActionResponse struct {
	ID             uuid.UUID       `json:"id"`
	RuleID         uuid.UUID       `json:"rule_id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	ActionConfig   json.RawMessage `json:"action_config"`
	ExecutionOrder int             `json:"execution_order"`
	IsActive       bool            `json:"is_active"`
	TemplateID     *uuid.UUID      `json:"template_id,omitempty"`
	TemplateName   string          `json:"template_name,omitempty"`
}

// Encode implements web.Encoder for ActionResponse.
func (a ActionResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(a)
	return data, "application/json", err
}

// ActionList wraps a slice of actions for JSON encoding.
type ActionList []ActionResponse

// Encode implements web.Encoder for ActionList.
func (l ActionList) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	return data, "application/json", err
}

// RuleList wraps a slice of rules for JSON encoding.
type RuleList []RuleResponse

// Encode implements web.Encoder for RuleList.
func (l RuleList) Encode() ([]byte, string, error) {
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

// HTTPStatus implements the httpStatus interface to return 400 Bad Request for validation errors.
func (v ValidationErrors) HTTPStatus() int {
	return 400 // http.StatusBadRequest
}

// Encode implements web.Encoder for ValidationErrors.
func (v ValidationErrors) Encode() ([]byte, string, error) {
	data, err := json.Marshal(v)
	return data, "application/json", err
}

package ruleapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ============================================================
// Action Request Types
// ============================================================

// CreateActionRequest is the request body for adding an action to a rule.
type CreateActionRequest struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	ActionConfig   json.RawMessage `json:"action_config"`
	ExecutionOrder int             `json:"execution_order"`
	IsActive       bool            `json:"is_active"`
	TemplateID     *uuid.UUID      `json:"template_id,omitempty"`
}

// Decode implements web.Decoder for CreateActionRequest.
func (r *CreateActionRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// UpdateActionRequest is the request body for updating an action (partial update).
type UpdateActionRequest struct {
	Name           *string          `json:"name,omitempty"`
	Description    *string          `json:"description,omitempty"`
	ActionConfig   *json.RawMessage `json:"action_config,omitempty"`
	ExecutionOrder *int             `json:"execution_order,omitempty"`
	IsActive       *bool            `json:"is_active,omitempty"`
	TemplateID     *uuid.UUID       `json:"template_id,omitempty"`
}

// Decode implements web.Decoder for UpdateActionRequest.
func (r *UpdateActionRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// ============================================================
// Validation Response Types
// ============================================================

// ValidateRuleResponse is the response from the validation endpoint.
type ValidateRuleResponse struct {
	Valid  bool              `json:"valid"`
	RuleID uuid.UUID         `json:"rule_id"`
	Issues []ValidationIssue `json:"issues,omitempty"`
}

// Encode implements web.Encoder for ValidateRuleResponse.
func (v ValidateRuleResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(v)
	return data, "application/json", err
}

// ValidationIssue represents a single validation issue.
type ValidationIssue struct {
	Level   string `json:"level"`   // "error", "warning", "info"
	Field   string `json:"field"`   // JSON path to the problematic field
	Message string `json:"message"` // Human-readable description
}

// ============================================================
// Action Validation Functions
// ============================================================

// ValidateCreateAction validates a CreateActionRequest.
func ValidateCreateAction(req CreateActionRequest) *ValidationErrors {
	var errors []ValidationError

	// Trim whitespace before validation
	name := strings.TrimSpace(req.Name)
	if name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	}
	if len(name) > MaxNameLength {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("name must be %d characters or less", MaxNameLength),
		})
	}
	// Check if action_config is missing, empty, or JSON null
	if len(req.ActionConfig) == 0 || string(req.ActionConfig) == "null" {
		errors = append(errors, ValidationError{
			Field:   "action_config",
			Message: "action_config is required",
		})
	}
	// Validate action config is valid JSON (only if not empty/null)
	if len(req.ActionConfig) > 0 && string(req.ActionConfig) != "null" {
		var ac interface{}
		if err := json.Unmarshal(req.ActionConfig, &ac); err != nil {
			errors = append(errors, ValidationError{
				Field:   "action_config",
				Message: "action_config must be valid JSON",
			})
		}
	}
	if req.ExecutionOrder < 0 {
		errors = append(errors, ValidationError{
			Field:   "execution_order",
			Message: "execution_order must be non-negative",
		})
	}

	if len(errors) > 0 {
		return &ValidationErrors{Errors: errors}
	}
	return nil
}

// ValidateUpdateAction validates an UpdateActionRequest.
func ValidateUpdateAction(req UpdateActionRequest) *ValidationErrors {
	var errors []ValidationError

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) > MaxNameLength {
			errors = append(errors, ValidationError{
				Field:   "name",
				Message: fmt.Sprintf("name must be %d characters or less", MaxNameLength),
			})
		}
	}
	// Validate action config is valid JSON if provided
	if req.ActionConfig != nil && len(*req.ActionConfig) > 0 {
		var ac interface{}
		if err := json.Unmarshal(*req.ActionConfig, &ac); err != nil {
			errors = append(errors, ValidationError{
				Field:   "action_config",
				Message: "action_config must be valid JSON",
			})
		}
	}
	if req.ExecutionOrder != nil && *req.ExecutionOrder < 0 {
		errors = append(errors, ValidationError{
			Field:   "execution_order",
			Message: "execution_order must be non-negative",
		})
	}

	if len(errors) > 0 {
		return &ValidationErrors{Errors: errors}
	}
	return nil
}

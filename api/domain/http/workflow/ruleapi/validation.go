package ruleapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// MaxNameLength is the maximum allowed length for rule names.
const MaxNameLength = 255

// ValidateCreateRule validates a CreateRuleRequest.
func ValidateCreateRule(req CreateRuleRequest) *ValidationErrors {
	var errors []ValidationError

	// Trim whitespace before validation
	name := strings.TrimSpace(req.Name)
	if name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(name) > MaxNameLength {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("name must be %d characters or less", MaxNameLength),
		})
	}
	if req.EntityID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "entity_id",
			Message: "entity_id is required",
		})
	}
	if req.EntityTypeID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "entity_type_id",
			Message: "entity_type_id is required",
		})
	}
	if req.TriggerTypeID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "trigger_type_id",
			Message: "trigger_type_id is required",
		})
	}

	// Validate trigger conditions JSON if provided
	if len(req.TriggerConditions) > 0 {
		var tc interface{}
		if err := json.Unmarshal(req.TriggerConditions, &tc); err != nil {
			errors = append(errors, ValidationError{
				Field:   "trigger_conditions",
				Message: "trigger_conditions must be valid JSON",
			})
		}
	}

	// Validate embedded actions
	for i, action := range req.Actions {
		if action.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("actions[%d].name", i),
				Message: "action name is required",
			})
		}
		if len(action.ActionConfig) == 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("actions[%d].action_config", i),
				Message: "action_config is required",
			})
		}
		// Validate action config JSON
		if len(action.ActionConfig) > 0 {
			var ac interface{}
			if err := json.Unmarshal(action.ActionConfig, &ac); err != nil {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("actions[%d].action_config", i),
					Message: "action_config must be valid JSON",
				})
			}
		}
		// Validate execution order is non-negative
		if action.ExecutionOrder < 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("actions[%d].execution_order", i),
				Message: "execution_order must be non-negative",
			})
		}
	}

	if len(errors) > 0 {
		return &ValidationErrors{Errors: errors}
	}
	return nil
}

// ValidateUpdateRule validates an UpdateRuleRequest.
func ValidateUpdateRule(req UpdateRuleRequest) *ValidationErrors {
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

	// Validate trigger conditions JSON if provided
	if req.TriggerConditions != nil && len(*req.TriggerConditions) > 0 {
		var tc interface{}
		if err := json.Unmarshal(*req.TriggerConditions, &tc); err != nil {
			errors = append(errors, ValidationError{
				Field:   "trigger_conditions",
				Message: "trigger_conditions must be valid JSON",
			})
		}
	}

	if len(errors) > 0 {
		return &ValidationErrors{Errors: errors}
	}
	return nil
}

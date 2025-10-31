package formdataapp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
)

// FormValidationError represents a validation error for a specific entity in a form.
type FormValidationError struct {
	EntityName      string   `json:"entity_name"`
	Operation       string   `json:"operation"`
	MissingFields   []string `json:"missing_fields"`
	AvailableFields []string `json:"available_fields,omitempty"`
}

// FormValidationResult represents the result of validating a form configuration.
type FormValidationResult struct {
	Valid  bool                  `json:"valid"`
	Errors []FormValidationError `json:"errors,omitempty"`
}

// Encode implements the encoder interface for FormValidationResult.
func (r FormValidationResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// FormValidationRequest represents a request to validate a form configuration.
type FormValidationRequest struct {
	Operations map[string]formdataregistry.EntityOperation `json:"operations"` // entity name -> operation type
}

// Decode implements the decoder interface for FormValidationRequest.
func (r *FormValidationRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// ValidateForm validates that a form has all required fields for the specified operations.
//
// This method checks that for each entity and operation type in the request,
// the form contains all fields marked as required in the entity's app model.
//
// Example request:
//
//	{
//	  "operations": {
//	    "users": "create",
//	    "assets": "create"
//	  }
//	}
//
// Returns a ValidationResult indicating whether the form is valid and listing
// any missing required fields per entity.
func (a *App) ValidateForm(
	ctx context.Context,
	formID uuid.UUID,
	req FormValidationRequest,
) (FormValidationResult, error) {
	// Load form fields
	fields, err := a.formFieldBus.QueryByFormID(ctx, formID)
	if err != nil {
		return FormValidationResult{}, errs.Newf(errs.Internal, "load form fields: %s", err)
	}

	// Build a map of entity names to their field names
	// We need to match the entity_id in fields to the entity names in the request
	// For now, we'll build a map of all field names and check against all entities
	// This is a limitation - ideally we'd query workflow.entities to map IDs to names

	allFieldNames := make([]string, len(fields))
	for i, field := range fields {
		allFieldNames[i] = field.Name
	}

	// Validate each entity operation
	var validationErrors []FormValidationError

	for entityName, operation := range req.Operations {
		// Get entity registration
		reg, err := a.registry.Get(entityName)
		if err != nil {
			validationErrors = append(validationErrors, FormValidationError{
				EntityName:    entityName,
				Operation:     operation.String(),
				MissingFields: []string{fmt.Sprintf("entity '%s' not registered in system", entityName)},
			})
			continue
		}

		// Get required fields based on operation
		var model interface{}
		switch operation {
		case formdataregistry.OperationCreate:
			model = reg.CreateModel
		case formdataregistry.OperationUpdate:
			model = reg.UpdateModel
		default:
			return FormValidationResult{}, errs.Newf(errs.InvalidArgument, "invalid operation: %s", operation)
		}

		if model == nil {
			// No model registered, skip validation
			continue
		}

		// Extract required fields from model using reflection
		requiredFields := formdataregistry.GetRequiredFields(model)

		// Check which required fields are missing
		missingFields := findMissingFields(requiredFields, allFieldNames)

		if len(missingFields) > 0 {
			validationErrors = append(validationErrors, FormValidationError{
				EntityName:      entityName,
				Operation:       operation.String(),
				MissingFields:   missingFields,
				AvailableFields: allFieldNames,
			})
		}
	}

	if len(validationErrors) > 0 {
		return FormValidationResult{
			Valid:  false,
			Errors: validationErrors,
		}, nil
	}

	return FormValidationResult{
		Valid:  true,
		Errors: nil,
	}, nil
}

// validateFormRequiredFields checks that a form has all required fields before executing operations.
// This is called automatically before upserting form data.
func (a *App) validateFormRequiredFields(
	ctx context.Context,
	formID uuid.UUID,
	req FormDataRequest,
) error {
	// Build validation request from operations
	validationReq := FormValidationRequest{
		Operations: make(map[string]formdataregistry.EntityOperation),
	}

	for entityName, meta := range req.Operations {
		validationReq.Operations[entityName] = meta.Operation
	}

	// Validate
	result, err := a.ValidateForm(ctx, formID, validationReq)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !result.Valid {
		return errs.Newf(errs.InvalidArgument, "form validation failed: %+v", result.Errors)
	}

	return nil
}

// findMissingFields returns the fields that are required but not present.
func findMissingFields(requiredFields []string, availableFields []string) []string {
	// Create a map of available fields for O(1) lookup
	available := make(map[string]bool)
	for _, field := range availableFields {
		available[field] = true
	}

	// Find missing required fields
	var missing []string
	for _, required := range requiredFields {
		if !available[required] {
			missing = append(missing, required)
		}
	}

	return missing
}
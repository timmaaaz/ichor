package formdataapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
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

// Validate checks that the request is valid.
func (r FormValidationRequest) Validate() error {
	if len(r.Operations) == 0 {
		return errs.Newf(errs.InvalidArgument, "validate: operations is required")
	}

	for entityName, operation := range r.Operations {
		if !operation.IsValid() {
			return errs.Newf(errs.InvalidArgument, "entity %s has invalid operation: %s", entityName, operation)
		}
	}

	return nil
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
	// Check form exists
	_, err := a.formBus.QueryByID(ctx, formID)
	if err != nil {
		return FormValidationResult{}, errs.New(errs.NotFound, err)
	}

	// Load form fields
	fields, err := a.formFieldBus.QueryByFormID(ctx, formID)
	if err != nil {
		return FormValidationResult{}, errs.Newf(errs.Internal, "load form fields: %s", err)
	}

	// Build a map of entity names to their field names
	// For lineitems field type, we need to extract nested field names from the config
	// Entity names use schema.table format (e.g., "sales.orders", "sales.order_line_items")
	entityFieldsMap := make(map[string][]string)

	// Also track auto-populated fields (those with {{$me}}, {{$now}} defaults)
	autoPopulatedFields := make(map[string]bool)

	for _, field := range fields {
		// Use schema.table format as entity key (matches registry entity names like "sales.customers")
		entityKey := field.EntitySchema + "." + field.EntityTable

		// Check if this field has auto-populated defaults
		if isAutoPopulatedField(field) {
			autoPopulatedFields[field.Name] = true
		}

		// Handle lineitems field type specially
		if field.FieldType == "lineitems" {
			// Parse the LineItemsFieldConfig from the Config JSON
			var lineItemsConfig struct {
				ParentField string `json:"parent_field"`
				Fields      []struct {
					Name string `json:"name"`
				} `json:"fields"`
			}
			if err := json.Unmarshal(field.Config, &lineItemsConfig); err == nil {
				// Add the parent field (FK that references parent entity)
				// This field is auto-populated via template variables, but still needs to be in available fields
				if lineItemsConfig.ParentField != "" {
					entityFieldsMap[entityKey] = append(entityFieldsMap[entityKey], lineItemsConfig.ParentField)
				}

				// Add all nested field names to this entity's available fields
				for _, nestedField := range lineItemsConfig.Fields {
					entityFieldsMap[entityKey] = append(entityFieldsMap[entityKey], nestedField.Name)
				}
			}
		} else {
			// Regular field - add to entity's field list
			entityFieldsMap[entityKey] = append(entityFieldsMap[entityKey], field.Name)
		}
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

		// Get available fields for this specific entity
		entityFields, exists := entityFieldsMap[entityName]
		if !exists {
			// Entity has no form fields configured
			validationErrors = append(validationErrors, FormValidationError{
				EntityName:      entityName,
				Operation:       operation.String(),
				MissingFields:   requiredFields,
				AvailableFields: []string{},
			})
			continue
		}

		// Check which required fields are missing
		missingFields := findMissingFields(requiredFields, entityFields, autoPopulatedFields)

		if len(missingFields) > 0 {
			validationErrors = append(validationErrors, FormValidationError{
				EntityName:      entityName,
				Operation:       operation.String(),
				MissingFields:   missingFields,
				AvailableFields: entityFields,
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

// isAutoPopulatedField checks if a field has a magic default configured ({{$me}}, {{$now}}).
// Fields with these defaults are auto-populated by the backend and should not be required
// in form configurations.
func isAutoPopulatedField(field formfieldbus.FormField) bool {
	var cfg formfieldbus.FieldDefaultConfig
	if err := json.Unmarshal(field.Config, &cfg); err != nil {
		return false
	}

	hasBuiltin := func(s string) bool {
		return strings.Contains(s, "{{$me}}") ||
			strings.Contains(s, "{{$now}}") ||
			strings.Contains(s, "{{$")
	}

	return hasBuiltin(cfg.DefaultValue) ||
		hasBuiltin(cfg.DefaultValueCreate) ||
		hasBuiltin(cfg.DefaultValueUpdate)
}

// findMissingFields returns the fields that are required but not present.
// Auto-populated fields (those with {{$me}}, {{$now}} defaults) are excluded
// from the check as they are system-provided and not expected in form configurations.
func findMissingFields(requiredFields []string, availableFields []string, autoPopulatedFields map[string]bool) []string {
	// Create a map of available fields for O(1) lookup
	available := make(map[string]bool)
	for _, field := range availableFields {
		available[field] = true
	}

	// Find missing required fields (excluding auto-populated fields)
	var missing []string
	for _, required := range requiredFields {
		// Skip auto-populated fields - they are system-provided via {{$me}}, {{$now}}
		if autoPopulatedFields[required] {
			continue
		}
		if !available[required] {
			missing = append(missing, required)
		}
	}

	return missing
}
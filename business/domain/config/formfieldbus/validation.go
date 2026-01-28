package formfieldbus

import (
	"encoding/json"
	"fmt"
	"strings"
)

// =============================================================================
// Validation Result Types
// =============================================================================

// FormValidationResult contains all validation errors and warnings
type FormValidationResult struct {
	Errors   []FormValidationError
	Warnings []FormValidationWarning
}

// FormValidationError represents a validation error
type FormValidationError struct {
	Field   string // e.g., "fields[0].name" or "entity:sales.orders"
	Message string
	Code    string // e.g., "MISSING_ID_FIELD", "INVALID_FIELD_TYPE"
}

// FormValidationWarning represents a non-fatal warning
type FormValidationWarning struct {
	Field   string
	Message string
}

// HasErrors returns true if there are any validation errors
func (r *FormValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// AddError adds a validation error to the result
func (r *FormValidationResult) AddError(field, message, code string) {
	r.Errors = append(r.Errors, FormValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// AddWarning adds a validation warning to the result
func (r *FormValidationResult) AddWarning(field, message string) {
	r.Warnings = append(r.Warnings, FormValidationWarning{
		Field:   field,
		Message: message,
	})
}

// Error implements the error interface, returning a formatted error message
func (r *FormValidationResult) Error() string {
	if !r.HasErrors() {
		return ""
	}
	var msgs []string
	for _, err := range r.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

// =============================================================================
// Form Validation Functions
// =============================================================================

// FormValidationOptions configures the validation behavior
type FormValidationOptions struct {
	// SupportsUpdate indicates this form can edit existing records
	// When true, all entities MUST have a hidden "id" field
	SupportsUpdate bool

	// FormName is used for error messages
	FormName string
}

// ValidateFormFields validates a collection of form fields
func ValidateFormFields(fields []NewFormField, opts FormValidationOptions) *FormValidationResult {
	result := &FormValidationResult{}

	// Group fields by entity
	entitiesByKey := make(map[string][]NewFormField)
	for _, f := range fields {
		key := fmt.Sprintf("%s.%s", f.EntitySchema, f.EntityTable)
		entitiesByKey[key] = append(entitiesByKey[key], f)
	}

	// Validate each entity has required fields
	for entityKey, entityFields := range entitiesByKey {
		validateEntity(result, entityKey, entityFields, opts)
	}

	// Validate individual fields
	for i, f := range fields {
		validateField(result, f, fmt.Sprintf("fields[%d]", i), opts)
	}

	return result
}

// validateEntity checks that an entity has all required fields
func validateEntity(result *FormValidationResult, entityKey string, fields []NewFormField, opts FormValidationOptions) {
	prefix := fmt.Sprintf("entity:%s", entityKey)

	// Check for id field if form supports updates
	if opts.SupportsUpdate {
		hasIDField := false
		isLineItemsOnlyEntity := true // Assume true, then check

		for _, f := range fields {
			// If there are non-lineitems fields for this entity, it needs a top-level id
			if f.FieldType != "lineitems" {
				isLineItemsOnlyEntity = false
			}

			if f.Name == "id" {
				hasIDField = true
				// Verify id field is hidden
				var config map[string]interface{}
				if err := json.Unmarshal(f.Config, &config); err == nil {
					if hidden, ok := config["hidden"].(bool); !ok || !hidden {
						result.AddWarning(
							prefix,
							fmt.Sprintf("id field for %s should have hidden:true in config", entityKey),
						)
					}
				}
				break
			}
		}

		// If the entity only has lineitems fields, the id check is done inside
		// validateLineItemsField, so skip the entity-level check
		if !hasIDField && !isLineItemsOnlyEntity {
			result.AddError(
				prefix,
				fmt.Sprintf("entity %s is missing required 'id' field for update operations. "+
					"Add a hidden id field: {Name: \"id\", FieldType: \"hidden\", Config: `{\"hidden\": true}`}",
					entityKey),
				"MISSING_ID_FIELD",
			)
		}
	}

	// NOTE: Audit field validation (created_by, created_date, updated_by, updated_date)
	// is intentionally NOT done here. Structural validation cannot know if the table
	// actually has these columns. This check is deferred to deep/schema validation
	// (Part 3) where we can introspect the database to verify column existence.
}

// validateField validates a single form field
func validateField(result *FormValidationResult, f NewFormField, prefix string, opts FormValidationOptions) {
	// Validate field name
	if f.Name == "" {
		result.AddError(prefix+".name", "field name is required", "REQUIRED")
	}

	// Validate field type
	if f.FieldType == "" {
		result.AddError(prefix+".field_type", "field_type is required", "REQUIRED")
	} else if !AllowedFieldTypes[f.FieldType] {
		result.AddError(prefix+".field_type", fmt.Sprintf("invalid field type: %s", f.FieldType), "INVALID_FIELD_TYPE")
	}

	// Validate entity info
	if f.EntitySchema == "" {
		result.AddError(prefix+".entity_schema", "entity_schema is required", "REQUIRED")
	}
	if f.EntityTable == "" {
		result.AddError(prefix+".entity_table", "entity_table is required", "REQUIRED")
	}

	// Type-specific validation
	switch f.FieldType {
	case "lineitems":
		validateLineItemsField(result, f, prefix, opts)
	case "smart-combobox", "dropdown":
		validateDropdownField(result, f, prefix)
	case "enum":
		validateEnumField(result, f, prefix)
	}
}

// validateLineItemsField validates a lineitems field configuration
func validateLineItemsField(result *FormValidationResult, f NewFormField, prefix string, opts FormValidationOptions) {
	var config LineItemsFieldConfig
	if err := json.Unmarshal(f.Config, &config); err != nil {
		result.AddError(prefix+".config", fmt.Sprintf("invalid lineitems config: %v", err), "INVALID_CONFIG")
		return
	}

	// Required fields
	if config.Entity == "" {
		result.AddError(prefix+".config.entity", "entity is required for lineitems", "REQUIRED")
	}
	if config.ParentField == "" {
		result.AddError(prefix+".config.parent_field", "parent_field is required for lineitems", "REQUIRED")
	}
	if config.ExecutionOrder <= 0 {
		result.AddError(prefix+".config.execution_order", "execution_order must be > 0 for lineitems", "INVALID_VALUE")
	}

	// Check for id field in line item fields if form supports updates
	if opts.SupportsUpdate {
		hasIDField := false
		for _, lif := range config.Fields {
			if lif.Name == "id" {
				hasIDField = true
				break
			}
		}

		if !hasIDField {
			result.AddError(
				prefix+".config.fields",
				fmt.Sprintf("lineitems for %s is missing 'id' field. "+
					"Add: {Name: \"id\", Type: \"hidden\", Hidden: true}",
					config.Entity),
				"MISSING_ID_FIELD",
			)
		}
	}

	// Validate each line item field
	for i, lif := range config.Fields {
		lifPrefix := fmt.Sprintf("%s.config.fields[%d]", prefix, i)

		if lif.Name == "" {
			result.AddError(lifPrefix+".name", "field name is required", "REQUIRED")
		}

		if lif.Type == "" {
			result.AddError(lifPrefix+".type", "field type is required", "REQUIRED")
		} else if !AllowedLineItemFieldTypes[lif.Type] {
			result.AddError(lifPrefix+".type", fmt.Sprintf("invalid line item field type: %s", lif.Type), "INVALID_FIELD_TYPE")
		}

		// Validate dropdown config for dropdown fields
		if lif.Type == "dropdown" && lif.DropdownConfig == nil {
			result.AddError(lifPrefix+".dropdown_config", "dropdown_config required for dropdown fields", "REQUIRED")
		}

		if lif.DropdownConfig != nil {
			if lif.DropdownConfig.LabelColumn == "" {
				result.AddError(lifPrefix+".dropdown_config.label_column", "label_column is required", "REQUIRED")
			}
			if lif.DropdownConfig.ValueColumn == "" {
				result.AddError(lifPrefix+".dropdown_config.value_column", "value_column is required", "REQUIRED")
			}
			if lif.DropdownConfig.Entity == "" && lif.DropdownConfig.TableConfigName == "" {
				result.AddError(lifPrefix+".dropdown_config", "either entity or table_config_name is required", "REQUIRED")
			}
		}

		// Validate enum config
		if lif.Type == "enum" && lif.EnumName == "" {
			result.AddError(lifPrefix+".enum_name", "enum_name required for enum fields", "REQUIRED")
		}
	}
}

// validateDropdownField validates smart-combobox and dropdown fields
func validateDropdownField(result *FormValidationResult, f NewFormField, prefix string) {
	var config map[string]interface{}
	if err := json.Unmarshal(f.Config, &config); err != nil {
		result.AddError(prefix+".config", fmt.Sprintf("invalid dropdown config: %v", err), "INVALID_CONFIG")
		return
	}

	// Check for entity or table_config_name
	entity, hasEntity := config["entity"].(string)
	tableConfigName, hasTableConfig := config["table_config_name"].(string)

	if !hasEntity && !hasTableConfig {
		result.AddError(prefix+".config", "dropdown requires 'entity' or 'table_config_name'", "REQUIRED")
	}

	if hasEntity && entity == "" {
		result.AddError(prefix+".config.entity", "entity cannot be empty", "INVALID_VALUE")
	}

	if hasTableConfig && tableConfigName == "" {
		result.AddError(prefix+".config.table_config_name", "table_config_name cannot be empty", "INVALID_VALUE")
	}
}

// validateEnumField validates enum fields
func validateEnumField(result *FormValidationResult, f NewFormField, prefix string) {
	var config map[string]interface{}
	if err := json.Unmarshal(f.Config, &config); err != nil {
		result.AddError(prefix+".config", fmt.Sprintf("invalid enum config: %v", err), "INVALID_CONFIG")
		return
	}

	enumName, hasEnum := config["enum_name"].(string)
	if !hasEnum || enumName == "" {
		result.AddError(prefix+".config.enum_name", "enum_name is required for enum fields", "REQUIRED")
	}
}

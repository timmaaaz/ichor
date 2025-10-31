package formdataregistry

import (
	"reflect"
	"strings"
)

// GetRequiredFields uses reflection to extract field names that have validate:"required" tags.
// It returns the JSON field names (from json:"field_name" tags) for all required fields.
//
// Example:
//
//	type NewAsset struct {
//	    ValidAssetID     string `json:"valid_asset_id" validate:"required"`
//	    SerialNumber     string `json:"serial_number" validate:"required"`
//	    AssetConditionID string `json:"asset_condition_id" validate:"required"`
//	    LastMaintenance  string `json:"last_maintenance"`
//	}
//
//	fields := GetRequiredFields(NewAsset{})
//	// Returns: ["valid_asset_id", "serial_number", "asset_condition_id"]
func GetRequiredFields(model interface{}) []string {
	if model == nil {
		return nil
	}

	var requiredFields []string

	// Get the type of the model
	t := reflect.TypeOf(model)

	// If it's a pointer, get the underlying type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only process structs
	if t.Kind() != reflect.Struct {
		return nil
	}

	// Iterate through all fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get the validate tag
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// Check if "required" is in the validation tags
		// Tags can be like: "required", "required,email", "omitempty,required", etc.
		if !isRequired(validateTag) {
			continue
		}

		// Get the JSON field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			// If no json tag, skip this field (shouldn't happen in practice)
			continue
		}

		// Parse the json tag (format: "field_name,omitempty" or just "field_name")
		jsonFieldName := parseJSONTag(jsonTag)
		if jsonFieldName == "" || jsonFieldName == "-" {
			// Skip fields with no name or explicitly excluded
			continue
		}

		requiredFields = append(requiredFields, jsonFieldName)
	}

	return requiredFields
}

// isRequired checks if the validate tag contains "required"
func isRequired(validateTag string) bool {
	// Split by comma to handle multiple validation rules
	parts := strings.Split(validateTag, ",")
	for _, part := range parts {
		if strings.TrimSpace(part) == "required" {
			return true
		}
	}
	return false
}

// parseJSONTag extracts the field name from a json struct tag
// Examples:
//   - "field_name" -> "field_name"
//   - "field_name,omitempty" -> "field_name"
//   - "-" -> "-"
func parseJSONTag(jsonTag string) string {
	// Split by comma to get the field name (first part)
	parts := strings.Split(jsonTag, ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}
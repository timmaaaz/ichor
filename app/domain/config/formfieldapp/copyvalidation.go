package formfieldapp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

// validateCopyFromField checks that copy_from_field references a valid sibling field
// in the same form and entity combination. Called at form field Create and Update time.
//
// For regular fields: the referenced field must exist as a sibling in the same
// form_id and entity_schema.entity_table.
//
// For lineitems fields: each sub-field's copy_from_field is checked against
// other sub-field names within the same lineitems config block.
func validateCopyFromField(
	ctx context.Context,
	bus *formfieldbus.Business,
	formID uuid.UUID,
	entitySchema, entityTable, fieldName, fieldType string,
	config json.RawMessage,
) error {
	if len(config) == 0 {
		return nil
	}

	copyFromField, copyRefs, allNames, err := extractCopyFromRefs(fieldType, config)
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "parse config for copy_from_field: %s", err)
	}

	if copyFromField == "" && len(copyRefs) == 0 {
		return nil
	}

	// Validate top-level copy_from_field against sibling fields in the form.
	if copyFromField != "" {
		if copyFromField == fieldName {
			return errs.Newf(errs.InvalidArgument,
				"copy_from_field: field %q cannot reference itself", fieldName)
		}

		siblings, err := bus.QueryByFormID(ctx, formID)
		if err != nil {
			return errs.Newf(errs.Internal, "query sibling fields: %s", err)
		}

		if !siblingFieldExists(copyFromField, entitySchema, entityTable, siblings) {
			return errs.Newf(errs.InvalidArgument,
				"copy_from_field: field %q references %q which does not exist in entity %s.%s",
				fieldName, copyFromField, entitySchema, entityTable)
		}
	}

	// Validate line item sub-field copy_from_field references within the block.
	for subField, copyRef := range copyRefs {
		if copyRef == subField {
			return errs.Newf(errs.InvalidArgument,
				"copy_from_field: line item field %q cannot reference itself", subField)
		}
		if _, exists := allNames[copyRef]; !exists {
			return errs.Newf(errs.InvalidArgument,
				"copy_from_field: line item field %q references %q which is not a field in this lineitems block",
				subField, copyRef)
		}
	}

	return nil
}

// extractCopyFromRefs parses the config JSON and returns:
//   - top-level copy_from_field string (for non-lineitems fields)
//   - copyRefs: map of sub-field name → copy_from_field value (only entries with non-empty CopyFromField)
//   - allNames: set of all sub-field names (for existence checking)
func extractCopyFromRefs(fieldType string, config json.RawMessage) (string, map[string]string, map[string]struct{}, error) {
	if fieldType == "lineitems" {
		var cfg formfieldbus.LineItemsFieldConfig
		if err := json.Unmarshal(config, &cfg); err != nil {
			return "", nil, nil, fmt.Errorf("unmarshal lineitems config: %w", err)
		}

		// Build set of all sub-field names and map of only those with copy_from_field.
		allNames := make(map[string]struct{}, len(cfg.Fields))
		copyRefs := make(map[string]string)
		for _, f := range cfg.Fields {
			allNames[f.Name] = struct{}{}
			if f.CopyFromField != "" {
				copyRefs[f.Name] = f.CopyFromField
			}
		}

		return "", copyRefs, allNames, nil
	}

	var cfg formfieldbus.FieldDefaultConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return "", nil, nil, fmt.Errorf("unmarshal field config: %w", err)
	}
	return cfg.CopyFromField, nil, nil, nil
}

// siblingFieldExists checks if a field with the given name exists among siblings
// in the same entity_schema.entity_table.
func siblingFieldExists(targetName, entitySchema, entityTable string, siblings []formfieldbus.FormField) bool {
	for _, s := range siblings {
		if s.EntitySchema == entitySchema &&
			s.EntityTable == entityTable &&
			s.Name == targetName {
			return true
		}
	}
	return false
}

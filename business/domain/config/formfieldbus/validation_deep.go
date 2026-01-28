package formfieldbus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/introspectionbus"
)

// =============================================================================
// Deep Schema Validation
// =============================================================================

// ValidateFormFieldsAgainstSchema validates form fields against the live database schema.
// This catches issues that structural validation cannot detect:
// - Non-existent tables/entities
// - Non-existent columns
// - Invalid dropdown/enum references
func ValidateFormFieldsAgainstSchema(
	ctx context.Context,
	introspection *introspectionbus.Business,
	fields []NewFormField,
	opts FormValidationOptions,
) *FormValidationResult {
	result := &FormValidationResult{}

	// Cache schema lookups to avoid repeated queries
	tableCache := make(map[string][]introspectionbus.Column) // "schema.table" -> columns
	tableExistsCache := make(map[string]bool)                // "schema.table" -> exists

	for i, f := range fields {
		prefix := fmt.Sprintf("fields[%d]", i)
		entityKey := f.EntitySchema + "." + f.EntityTable

		// 1. Check table exists
		columns, exists := getOrQueryColumns(ctx, introspection, f.EntitySchema, f.EntityTable, tableCache, tableExistsCache)
		if !exists {
			result.AddError(prefix,
				fmt.Sprintf("table %s does not exist in database", entityKey),
				"TABLE_NOT_FOUND")
			continue
		}

		// 2. Check field exists as column (skip virtual fields like 'lineitems')
		if f.FieldType != "lineitems" {
			if !columnExists(columns, f.Name) {
				result.AddError(prefix,
					fmt.Sprintf("column %q not found in table %s", f.Name, entityKey),
					"COLUMN_NOT_FOUND")
			}
		}

		// 3. Type-specific deep validation
		switch f.FieldType {
		case "lineitems":
			validateLineItemsAgainstSchema(ctx, introspection, f, prefix, result, tableCache, tableExistsCache)
		case "dropdown", "smart-combobox":
			validateDropdownAgainstSchema(ctx, introspection, f, prefix, result, tableCache, tableExistsCache)
		case "enum":
			validateEnumAgainstSchema(ctx, introspection, f, prefix, result)
		}
	}

	return result
}

// =============================================================================
// Helper Functions
// =============================================================================

// getOrQueryColumns returns columns for a table, using cache if available.
// Returns the columns and a boolean indicating if the table exists.
func getOrQueryColumns(
	ctx context.Context,
	introspection *introspectionbus.Business,
	schema, table string,
	tableCache map[string][]introspectionbus.Column,
	tableExistsCache map[string]bool,
) ([]introspectionbus.Column, bool) {
	key := schema + "." + table

	// Check cache first
	if exists, ok := tableExistsCache[key]; ok {
		if !exists {
			return nil, false
		}
		return tableCache[key], true
	}

	// Query the database
	columns, err := introspection.QueryColumns(ctx, schema, table)
	if err != nil {
		// If there's an error, the table likely doesn't exist
		tableExistsCache[key] = false
		return nil, false
	}

	// Empty columns also means table doesn't exist (or has no columns, which is effectively the same)
	if len(columns) == 0 {
		tableExistsCache[key] = false
		return nil, false
	}

	// Cache and return
	tableCache[key] = columns
	tableExistsCache[key] = true
	return columns, true
}

// columnExists checks if a column exists in the list of columns
func columnExists(columns []introspectionbus.Column, name string) bool {
	for _, col := range columns {
		if col.Name == name {
			return true
		}
	}
	return false
}

// =============================================================================
// Type-Specific Validators
// =============================================================================

// validateLineItemsAgainstSchema validates line items configuration against the database schema
func validateLineItemsAgainstSchema(
	ctx context.Context,
	introspection *introspectionbus.Business,
	f NewFormField,
	prefix string,
	result *FormValidationResult,
	tableCache map[string][]introspectionbus.Column,
	tableExistsCache map[string]bool,
) {
	var config LineItemsFieldConfig
	if err := json.Unmarshal(f.Config, &config); err != nil {
		// Structural validation already caught this
		return
	}

	// Parse entity (format: "schema.table")
	parts := strings.Split(config.Entity, ".")
	if len(parts) != 2 {
		result.AddError(prefix+".config.entity",
			fmt.Sprintf("invalid entity format %q, expected 'schema.table'", config.Entity),
			"INVALID_ENTITY_FORMAT")
		return
	}
	schema, table := parts[0], parts[1]

	// 1. Check that the line items table exists
	columns, exists := getOrQueryColumns(ctx, introspection, schema, table, tableCache, tableExistsCache)
	if !exists {
		result.AddError(prefix+".config.entity",
			fmt.Sprintf("line items table %s does not exist in database", config.Entity),
			"TABLE_NOT_FOUND")
		return
	}

	// 2. Check that the parent field (FK column) exists in the line items table
	if !columnExists(columns, config.ParentField) {
		result.AddError(prefix+".config.parent_field",
			fmt.Sprintf("FK column %q not found in table %s", config.ParentField, config.Entity),
			"FK_COLUMN_NOT_FOUND")
	}

	// 3. Check each field in the line items configuration
	for i, lif := range config.Fields {
		lifPrefix := fmt.Sprintf("%s.config.fields[%d]", prefix, i)

		// Skip hidden computed fields that might not be in DB
		if lif.Hidden && lif.Name == "order_id" {
			// Parent FK field - already validated above
			continue
		}

		// Check column exists
		if !columnExists(columns, lif.Name) {
			result.AddError(lifPrefix,
				fmt.Sprintf("column %q not found in table %s", lif.Name, config.Entity),
				"COLUMN_NOT_FOUND")
		}

		// Validate dropdown config if present
		if lif.DropdownConfig != nil {
			validateLineItemDropdownAgainstSchema(ctx, introspection, lif.DropdownConfig, lifPrefix, result, tableCache, tableExistsCache)
		}

		// Validate enum if present
		if lif.Type == "enum" && lif.EnumName != "" {
			validateEnumTypeAgainstSchema(ctx, introspection, lif.EnumName, lifPrefix, result)
		}
	}
}

// validateLineItemDropdownAgainstSchema validates a line item dropdown configuration
func validateLineItemDropdownAgainstSchema(
	ctx context.Context,
	introspection *introspectionbus.Business,
	config *DropdownConfig,
	prefix string,
	result *FormValidationResult,
	tableCache map[string][]introspectionbus.Column,
	tableExistsCache map[string]bool,
) {
	// If using entity reference (not table_config_name), validate it exists
	if config.Entity != "" {
		parts := strings.Split(config.Entity, ".")
		if len(parts) != 2 {
			result.AddError(prefix+".dropdown_config.entity",
				fmt.Sprintf("invalid entity format %q, expected 'schema.table'", config.Entity),
				"INVALID_ENTITY_FORMAT")
			return
		}
		schema, table := parts[0], parts[1]

		columns, exists := getOrQueryColumns(ctx, introspection, schema, table, tableCache, tableExistsCache)
		if !exists {
			result.AddError(prefix+".dropdown_config.entity",
				fmt.Sprintf("dropdown entity %s does not exist in database", config.Entity),
				"DROPDOWN_TABLE_NOT_FOUND")
			return
		}

		// Validate label and value columns exist
		if config.LabelColumn != "" && !columnExists(columns, config.LabelColumn) {
			result.AddError(prefix+".dropdown_config.label_column",
				fmt.Sprintf("label column %q not found in table %s", config.LabelColumn, config.Entity),
				"DROPDOWN_COLUMN_NOT_FOUND")
		}
		if config.ValueColumn != "" && !columnExists(columns, config.ValueColumn) {
			result.AddError(prefix+".dropdown_config.value_column",
				fmt.Sprintf("value column %q not found in table %s", config.ValueColumn, config.Entity),
				"DROPDOWN_COLUMN_NOT_FOUND")
		}
	}
	// If using table_config_name, we can't validate against schema directly
	// (would need to query the config table, which is a different concern)
}

// validateDropdownAgainstSchema validates a dropdown/combobox field against the database schema
func validateDropdownAgainstSchema(
	ctx context.Context,
	introspection *introspectionbus.Business,
	f NewFormField,
	prefix string,
	result *FormValidationResult,
	tableCache map[string][]introspectionbus.Column,
	tableExistsCache map[string]bool,
) {
	var config map[string]interface{}
	if err := json.Unmarshal(f.Config, &config); err != nil {
		// Structural validation already caught this
		return
	}

	// If using entity reference, validate it
	if entity, ok := config["entity"].(string); ok && entity != "" {
		parts := strings.Split(entity, ".")
		if len(parts) != 2 {
			result.AddError(prefix+".config.entity",
				fmt.Sprintf("invalid entity format %q, expected 'schema.table'", entity),
				"INVALID_ENTITY_FORMAT")
			return
		}
		schema, table := parts[0], parts[1]

		columns, exists := getOrQueryColumns(ctx, introspection, schema, table, tableCache, tableExistsCache)
		if !exists {
			result.AddError(prefix+".config.entity",
				fmt.Sprintf("dropdown entity %s does not exist in database", entity),
				"DROPDOWN_TABLE_NOT_FOUND")
			return
		}

		// Validate label and value columns
		if labelCol, ok := config["label_column"].(string); ok && labelCol != "" {
			if !columnExists(columns, labelCol) {
				result.AddError(prefix+".config.label_column",
					fmt.Sprintf("label column %q not found in table %s", labelCol, entity),
					"DROPDOWN_COLUMN_NOT_FOUND")
			}
		}
		if valueCol, ok := config["value_column"].(string); ok && valueCol != "" {
			if !columnExists(columns, valueCol) {
				result.AddError(prefix+".config.value_column",
					fmt.Sprintf("value column %q not found in table %s", valueCol, entity),
					"DROPDOWN_COLUMN_NOT_FOUND")
			}
		}
		// Also check display_field which is used by smart-combobox
		if displayField, ok := config["display_field"].(string); ok && displayField != "" {
			if !columnExists(columns, displayField) {
				result.AddError(prefix+".config.display_field",
					fmt.Sprintf("display column %q not found in table %s", displayField, entity),
					"DROPDOWN_COLUMN_NOT_FOUND")
			}
		}
	}
}

// validateEnumAgainstSchema validates an enum field against the database schema
func validateEnumAgainstSchema(
	ctx context.Context,
	introspection *introspectionbus.Business,
	f NewFormField,
	prefix string,
	result *FormValidationResult,
) {
	var config map[string]interface{}
	if err := json.Unmarshal(f.Config, &config); err != nil {
		// Structural validation already caught this
		return
	}

	enumName, ok := config["enum_name"].(string)
	if !ok || enumName == "" {
		// Structural validation already caught this
		return
	}

	validateEnumTypeAgainstSchema(ctx, introspection, enumName, prefix, result)
}

// validateEnumTypeAgainstSchema checks if an enum type exists in the database
func validateEnumTypeAgainstSchema(
	ctx context.Context,
	introspection *introspectionbus.Business,
	enumName string,
	prefix string,
	result *FormValidationResult,
) {
	// Parse enum name (format: "schema.enum_name")
	parts := strings.Split(enumName, ".")
	if len(parts) != 2 {
		result.AddError(prefix+".config.enum_name",
			fmt.Sprintf("invalid enum name format %q, expected 'schema.enum_name'", enumName),
			"INVALID_ENUM_FORMAT")
		return
	}
	schema, name := parts[0], parts[1]

	// Query enum types for the schema
	enums, err := introspection.QueryEnumTypes(ctx, schema)
	if err != nil {
		result.AddError(prefix+".config.enum_name",
			fmt.Sprintf("error querying enum types for schema %s: %v", schema, err),
			"ENUM_QUERY_ERROR")
		return
	}

	// Check if the enum exists
	found := false
	for _, e := range enums {
		if e.Name == name {
			found = true
			break
		}
	}

	if !found {
		result.AddError(prefix+".config.enum_name",
			fmt.Sprintf("enum type %s does not exist in database", enumName),
			"ENUM_NOT_FOUND")
	}
}

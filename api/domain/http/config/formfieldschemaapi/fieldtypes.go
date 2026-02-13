package formfieldschemaapi

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
)

//go:embed schemas/*.json
var schemaFS embed.FS

// fieldTypeMeta holds static metadata for a field type.
type fieldTypeMeta struct {
	Name        string
	Description string
}

// fieldTypeMetadata contains metadata for all supported form field types.
var fieldTypeMetadata = map[string]fieldTypeMeta{
	"boolean": {
		Name:        "Boolean",
		Description: "Checkbox or toggle for true/false values",
	},
	"currency": {
		Name:        "Currency",
		Description: "Monetary value input with configurable precision",
	},
	"date": {
		Name:        "Date",
		Description: "Date picker with optional min/max and future constraints",
	},
	"datetime": {
		Name:        "Date/Time",
		Description: "Combined date and time picker",
	},
	"dropdown": {
		Name:        "Dropdown",
		Description: "Select dropdown populated from an entity data source with auto-populate support",
	},
	"email": {
		Name:        "Email",
		Description: "Email address input with format validation",
	},
	"enum": {
		Name:        "Enum",
		Description: "Dropdown populated from a PostgreSQL ENUM type",
	},
	"hidden": {
		Name:        "Hidden",
		Description: "Hidden field for passing values not shown to the user (e.g., parent IDs)",
	},
	"lineitems": {
		Name:        "Line Items",
		Description: "Multi-row inline collection for child entity records (e.g., order line items)",
	},
	"number": {
		Name:        "Number",
		Description: "Numeric input with optional min, max, step, and precision",
	},
	"percent": {
		Name:        "Percent",
		Description: "Percentage input with configurable precision and range",
	},
	"smart-combobox": {
		Name:        "Smart Combobox",
		Description: "Autocomplete dropdown with search, entity data source, and auto-populate support",
	},
	"tel": {
		Name:        "Telephone",
		Description: "Telephone number input",
	},
	"text": {
		Name:        "Text",
		Description: "Single-line text input with optional length constraints and conditional behavior",
	},
	"textarea": {
		Name:        "Textarea",
		Description: "Multi-line text input with optional length constraints",
	},
	"time": {
		Name:        "Time",
		Description: "Time-only picker",
	},
}

// fieldTypeSchemas holds the loaded JSON schemas keyed by field type.
var fieldTypeSchemas map[string]json.RawMessage

func init() {
	fieldTypeSchemas = make(map[string]json.RawMessage, len(fieldTypeMetadata))

	for typeName := range fieldTypeMetadata {
		schemaPath := fmt.Sprintf("schemas/%s.json", typeName)
		schemaBytes, err := schemaFS.ReadFile(schemaPath)
		if err != nil {
			panic(fmt.Sprintf("failed to load schema for field type %s: %v", typeName, err))
		}

		var dummy interface{}
		if err := json.Unmarshal(schemaBytes, &dummy); err != nil {
			panic(fmt.Sprintf("invalid JSON schema for field type %s: %v", typeName, err))
		}

		fieldTypeSchemas[typeName] = schemaBytes
	}
}

// GetFieldTypes returns all field types in alphabetical order.
func GetFieldTypes() []FieldTypeInfo {
	typeNames := make([]string, 0, len(fieldTypeMetadata))
	for name := range fieldTypeMetadata {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)

	types := make([]FieldTypeInfo, 0, len(typeNames))
	for _, name := range typeNames {
		meta := fieldTypeMetadata[name]
		types = append(types, FieldTypeInfo{
			Type:         name,
			Name:         meta.Name,
			Description:  meta.Description,
			ConfigSchema: fieldTypeSchemas[name],
		})
	}
	return types
}

// getFieldTypeSchema returns the info for a specific field type.
func getFieldTypeSchema(fieldType string) (FieldTypeInfo, bool) {
	meta, found := fieldTypeMetadata[fieldType]
	if !found {
		return FieldTypeInfo{}, false
	}

	schema, hasSchema := fieldTypeSchemas[fieldType]
	if !hasSchema {
		return FieldTypeInfo{}, false
	}

	return FieldTypeInfo{
		Type:         fieldType,
		Name:         meta.Name,
		Description:  meta.Description,
		ConfigSchema: schema,
	}, true
}

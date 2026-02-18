package formfieldbus

import (
	"encoding/json"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// FormField represents a field configuration within a form.
type FormField struct {
	ID           uuid.UUID       `json:"id"`
	FormID       uuid.UUID       `json:"form_id"`
	EntityID     uuid.UUID       `json:"entity_id"`
	EntitySchema string          `json:"entity_schema"`
	EntityTable  string          `json:"entity_table"`
	Name         string          `json:"name"`
	Label        string          `json:"label"`
	FieldType    string          `json:"field_type"`
	FieldOrder   int             `json:"field_order"`
	Required     bool            `json:"required"`
	Config       json.RawMessage `json:"config"`
}

// NewFormField contains the information needed to create a new form field.
type NewFormField struct {
	FormID       uuid.UUID       `json:"form_id"`
	EntityID     uuid.UUID       `json:"entity_id"`
	EntitySchema string          `json:"entity_schema"`
	EntityTable  string          `json:"entity_table"`
	Name         string          `json:"name"`
	Label        string          `json:"label"`
	FieldType    string          `json:"field_type"`
	FieldOrder   int             `json:"field_order"`
	Required     bool            `json:"required"`
	Config       json.RawMessage `json:"config"`
}

// UpdateFormField contains the information needed to update a form field.
type UpdateFormField struct {
	FormID       *uuid.UUID       `json:"form_id,omitempty"`
	EntityID     *uuid.UUID       `json:"entity_id,omitempty"`
	EntitySchema *string          `json:"entity_schema,omitempty"`
	EntityTable  *string          `json:"entity_table,omitempty"`
	Name         *string          `json:"name,omitempty"`
	Label        *string          `json:"label,omitempty"`
	FieldType    *string          `json:"field_type,omitempty"`
	FieldOrder   *int             `json:"field_order,omitempty"`
	Required     *bool            `json:"required,omitempty"`
	Config       *json.RawMessage `json:"config,omitempty"`
}

// =============================================================================
// FIELD CONFIGURATION TYPES
// =============================================================================
// These types define the structure of the Config JSON for different field types.
// They provide type safety when building form field configurations.

// FieldDefaultConfig defines auto-population behavior for a field.
// This is used for audit fields like created_by, updated_by, created_date, updated_date
// which can be automatically populated using magic variables like {{$me}} and {{$now}}.
type FieldDefaultConfig struct {
	DefaultValue       string `json:"default_value,omitempty"`        // e.g., "{{$me}}" - applied to both create and update
	DefaultValueCreate string `json:"default_value_create,omitempty"` // e.g., "{{$me}}" - only for create operations
	DefaultValueUpdate string `json:"default_value_update,omitempty"` // e.g., "{{$me}}" - only for update operations
	Hidden             bool   `json:"hidden,omitempty"`               // If true, field is not rendered in UI
	CopyFromField      string `json:"copy_from_field,omitempty"`      // Copy value from this sibling field when target is absent
}

// AutoPopulateMapping defines how to populate a target field from a dropdown selection.
// When a user selects an option from a dropdown, additional fields can be auto-populated
// with values from the selected record based on these mappings.
type AutoPopulateMapping struct {
	SourceColumn string `json:"source_column"` // Column from dropdown record to copy (e.g., "selling_price")
	TargetField  string `json:"target_field"`  // Field to populate in the form/line item (e.g., "unit_price")
}

// DropdownConfig defines configuration for dropdown fields.
type DropdownConfig struct {
	Entity          string                `json:"entity,omitempty"`            // Format: "schema.table" (e.g., "products.products") - direct table query
	TableConfigName string                `json:"table_config_name,omitempty"` // Table config name for joined queries (preferred for complex lookups)
	LabelColumn     string                `json:"label_column"`
	ValueColumn     string                `json:"value_column"`
	DisplayColumns  []string              `json:"additional_display_columns,omitempty"`
	AutoPopulate    []AutoPopulateMapping `json:"auto_populate,omitempty"` // Auto-populate mappings for target fields
}

// ValidationConfig defines validation rules for fields.
type ValidationConfig struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
	// Date constraints
	MinDate      string `json:"min_date,omitempty"`       // "today", "{{field_name}}", or ISO date
	MaxDate      string `json:"max_date,omitempty"`       // "today", "{{field_name}}", or ISO date
	MustBeFuture bool   `json:"must_be_future,omitempty"` // Shorthand for min_date: "today"
}

// =============================================================================
// CONDITIONAL FIELD CONFIGURATION TYPES
// =============================================================================
// These types enable dynamic field behavior based on related field values.

// FieldOverrideConfig defines field property overrides for conditional field transformation.
// These properties override the field's base configuration when the depends_on condition matches.
type FieldOverrideConfig struct {
	Type       string            `json:"type,omitempty"`       // Override field type (e.g., "currency", "percent")
	Label      string            `json:"label,omitempty"`      // Override field label
	Validation *ValidationConfig `json:"validation,omitempty"` // Override validation rules
}

// DependsOnConfig defines conditional field behavior based on another field's value.
// This enables dynamic field transformation at runtime without frontend hardcoding.
//
// Example usage for discount field that changes based on discount_type:
//
//	DependsOn: &DependsOnConfig{
//	    Field: "discount_type",
//	    ValueMappings: map[string]FieldOverrideConfig{
//	        "flat":    {Type: "currency", Label: "Discount ($)"},
//	        "percent": {Type: "percent", Label: "Discount (%)", Validation: &ValidationConfig{Max: intPtr(100)}},
//	    },
//	    Default: FieldOverrideConfig{Type: "currency"},
//	}
type DependsOnConfig struct {
	Field         string                         `json:"field"`             // Field name to watch (e.g., "discount_type")
	ValueMappings map[string]FieldOverrideConfig `json:"value_mappings"`    // Map field values to override configs
	Default       FieldOverrideConfig            `json:"default,omitempty"` // Default overrides when no mapping matches
}

// =============================================================================
// FORM FIELD CONFIGURATION TYPES
// =============================================================================
// These types provide typed configuration for regular form fields.

// InlineCreateConfig defines configuration for inline entity creation.
type InlineCreateConfig struct {
	Enabled    bool   `json:"enabled"`
	FormName   string `json:"form_name,omitempty"`
	ButtonText string `json:"button_text,omitempty"`
}

// FormFieldConfig defines the typed configuration for regular form fields.
// This provides type safety when building form field configurations and
// includes support for conditional field behavior via DependsOn.
type FormFieldConfig struct {
	// Default value configuration
	DefaultValue       string `json:"default_value,omitempty"`
	DefaultValueCreate string `json:"default_value_create,omitempty"`
	DefaultValueUpdate string `json:"default_value_update,omitempty"`
	Hidden             bool   `json:"hidden,omitempty"`

	// Dropdown/combobox configuration
	Entity       string              `json:"entity,omitempty"`
	DisplayField string              `json:"display_field,omitempty"`
	LabelColumn  string              `json:"label_column,omitempty"`
	ValueColumn  string              `json:"value_column,omitempty"`
	InlineCreate *InlineCreateConfig `json:"inline_create,omitempty"`

	// Enum configuration
	EnumName string `json:"enum_name,omitempty"`

	// Validation
	Min       *int    `json:"min,omitempty"`
	Max       *int    `json:"max,omitempty"`
	Step      float64 `json:"step,omitempty"`
	Precision *int    `json:"precision,omitempty"`

	// Multi-entity form configuration
	ExecutionOrder int    `json:"execution_order,omitempty"`
	ParentEntity   string `json:"parent_entity,omitempty"`
	ParentField    string `json:"parent_field,omitempty"`

	// Conditional field behavior
	DependsOn *DependsOnConfig `json:"depends_on,omitempty"`

	// CopyFromField copies the value from a sibling field when this field is absent from submitted data.
	CopyFromField string `json:"copy_from_field,omitempty"`
}

// ToJSON marshals the config to json.RawMessage for use in FormField.Config.
func (c FormFieldConfig) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// =============================================================================
// LINE ITEM FIELD TYPES
// =============================================================================

// LineItemField defines a field within a lineitems configuration.
type LineItemField struct {
	Name           string            `json:"name"`
	Label          string            `json:"label"`
	Type           string            `json:"type"` // text, number, dropdown, enum, date, currency, percent, hidden
	Required       bool              `json:"required"`
	DropdownConfig *DropdownConfig   `json:"dropdown_config,omitempty"` // For type: "dropdown" (FK lookups)
	Validation     *ValidationConfig `json:"validation,omitempty"`
	// Default value configuration - same fields as FieldDefaultConfig for consistency
	Hidden             bool   `json:"hidden,omitempty"`               // If true, field is not rendered in UI
	DefaultValue       string `json:"default_value,omitempty"`        // Applied to both create and update
	DefaultValueCreate string `json:"default_value_create,omitempty"` // Only for create operations
	DefaultValueUpdate string `json:"default_value_update,omitempty"` // Only for update operations
	// EnumName references a PostgreSQL ENUM type for dropdown options.
	// Required when Type is "enum". Format: "schema.enum_name" (e.g., "sales.discount_type")
	EnumName string `json:"enum_name,omitempty"`
	// DependsOn enables conditional field behavior based on another field's value.
	// When set, the field's type/label/validation can change at runtime based on the
	// value of another field in the same line item.
	DependsOn *DependsOnConfig `json:"depends_on,omitempty"`
	// CopyFromField copies the value from another field in the same line item when this field is absent.
	CopyFromField string `json:"copy_from_field,omitempty"`
}

// LineItemsFieldConfig defines the configuration for a lineitems field type.
// This is stored as JSON in the Config field of FormField/NewFormField.
type LineItemsFieldConfig struct {
	ExecutionOrder    int             `json:"execution_order"`
	Entity            string          `json:"entity"`
	ParentField       string          `json:"parent_field"`
	Fields            []LineItemField `json:"fields"`
	ItemLabel         string          `json:"item_label"`
	SingularItemLabel string          `json:"singular_item_label"`
	MinItems          int             `json:"min_items"`
	MaxItems          int             `json:"max_items"`
	FullWidth         bool            `json:"full_width,omitempty"`
	Columns           int             `json:"columns,omitempty"` // 1-6, default 2
}

// ToJSON marshals the config to json.RawMessage for use in FormField.Config.
func (c LineItemsFieldConfig) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

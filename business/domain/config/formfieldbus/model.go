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
}

// DropdownConfig defines configuration for dropdown fields.
type DropdownConfig struct {
	Entity         string   `json:"entity"`                               // Format: "schema.table" (e.g., "products.products")
	LabelColumn    string   `json:"label_column"`
	ValueColumn    string   `json:"value_column"`
	DisplayColumns []string `json:"additional_display_columns,omitempty"`
}

// ValidationConfig defines validation rules for fields.
type ValidationConfig struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// LineItemField defines a field within a lineitems configuration.
type LineItemField struct {
	Name           string            `json:"name"`
	Label          string            `json:"label"`
	Type           string            `json:"type"`
	Required       bool              `json:"required"`
	DropdownConfig *DropdownConfig   `json:"dropdown_config,omitempty"`
	Validation     *ValidationConfig `json:"validation,omitempty"`
	// Default value configuration - same fields as FieldDefaultConfig for consistency
	Hidden             bool   `json:"hidden,omitempty"`               // If true, field is not rendered in UI
	DefaultValue       string `json:"default_value,omitempty"`        // Applied to both create and update
	DefaultValueCreate string `json:"default_value_create,omitempty"` // Only for create operations
	DefaultValueUpdate string `json:"default_value_update,omitempty"` // Only for update operations
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

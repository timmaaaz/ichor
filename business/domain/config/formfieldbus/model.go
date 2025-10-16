package formfieldbus

import (
	"encoding/json"

	"github.com/google/uuid"
)

// FormField represents a field configuration within a form.
type FormField struct {
	ID         uuid.UUID
	FormID     uuid.UUID
	Name       string
	Label      string
	FieldType  string
	FieldOrder int
	Required   bool
	Config     json.RawMessage
}

// NewFormField contains the information needed to create a new form field.
type NewFormField struct {
	FormID     uuid.UUID
	Name       string
	Label      string
	FieldType  string
	FieldOrder int
	Required   bool
	Config     json.RawMessage
}

// UpdateFormField contains the information needed to update a form field.
type UpdateFormField struct {
	FormID     *uuid.UUID
	Name       *string
	Label      *string
	FieldType  *string
	FieldOrder *int
	Required   *bool
	Config     *json.RawMessage
}

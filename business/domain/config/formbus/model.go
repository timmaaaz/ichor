package formbus

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// Form represents a form configuration in the system.
type Form struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	IsReferenceData   bool      `json:"is_reference_data"`
	AllowInlineCreate bool      `json:"allow_inline_create"`
}

// NewForm contains the information needed to create a new form.
type NewForm struct {
	Name              string `json:"name"`
	IsReferenceData   bool   `json:"is_reference_data"`
	AllowInlineCreate bool   `json:"allow_inline_create"`
}

// UpdateForm contains the information needed to update a form.
type UpdateForm struct {
	Name              *string `json:"name,omitempty"`
	IsReferenceData   *bool   `json:"is_reference_data,omitempty"`
	AllowInlineCreate *bool   `json:"allow_inline_create,omitempty"`
}

// FormWithFields represents a form with its associated fields for export/import.
type FormWithFields struct {
	Form   Form                      `json:"form"`
	Fields []formfieldbus.FormField `json:"fields"`
}

// ImportStats represents statistics from an import operation.
type ImportStats struct {
	ImportedCount int `json:"imported_count"`
	SkippedCount  int `json:"skipped_count"`
	UpdatedCount  int `json:"updated_count"`
}

package formbus

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

// Form represents a form configuration in the system.
type Form struct {
	ID                 uuid.UUID
	Name               string
	IsReferenceData    bool
	AllowInlineCreate  bool
}

// NewForm contains the information needed to create a new form.
type NewForm struct {
	Name               string
	IsReferenceData    bool
	AllowInlineCreate  bool
}

// UpdateForm contains the information needed to update a form.
type UpdateForm struct {
	Name               *string
	IsReferenceData    *bool
	AllowInlineCreate  *bool
}

// FormWithFields represents a form with its associated fields for export/import.
type FormWithFields struct {
	Form   Form
	Fields []formfieldbus.FormField
}

// ImportStats represents statistics from an import operation.
type ImportStats struct {
	ImportedCount int
	SkippedCount  int
	UpdatedCount  int
}

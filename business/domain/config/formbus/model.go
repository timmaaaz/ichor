package formbus

import "github.com/google/uuid"

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

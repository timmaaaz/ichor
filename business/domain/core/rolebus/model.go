package rolebus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// Role represents information about an individual role.
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// NewRole contains information needed to create a new role.
type NewRole struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateRole contains information needed to update a role. Fields that are not
// included are intended to have separate endpoints or permissions to update.
type UpdateRole struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

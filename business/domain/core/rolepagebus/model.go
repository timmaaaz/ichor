package rolepagebus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// RolePage represents information about role-page access mapping.
type RolePage struct {
	ID        uuid.UUID `json:"id"`
	RoleID    uuid.UUID `json:"role_id"`
	PageID    uuid.UUID `json:"page_id"`
	CanAccess bool      `json:"can_access"`
}

// NewRolePage contains information needed to create a new role-page mapping.
type NewRolePage struct {
	RoleID    uuid.UUID `json:"role_id"`
	PageID    uuid.UUID `json:"page_id"`
	CanAccess bool      `json:"can_access"`
}

// UpdateRolePage contains information needed to update a role-page mapping.
type UpdateRolePage struct {
	CanAccess *bool `json:"can_access,omitempty"`
}

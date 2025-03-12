package rolebus

import "github.com/google/uuid"

// Role represents information about an individual role.
type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
}

// NewRole contains information needed to create a new role.
type NewRole struct {
	Name        string
	Description string
}

// UpdateRole contains information needed to update a role. Fields that are not
// included are intended to have separate endpoints or permissions to update.
type UpdateRole struct {
	Name        *string
	Description *string
}

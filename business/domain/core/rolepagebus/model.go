package rolepagebus

import "github.com/google/uuid"

// RolePage represents information about role-page access mapping.
type RolePage struct {
	ID         uuid.UUID
	RoleID     uuid.UUID
	PageID     uuid.UUID
	CanAccess  bool
	ShowInMenu bool
}

// NewRolePage contains information needed to create a new role-page mapping.
type NewRolePage struct {
	RoleID     uuid.UUID
	PageID     uuid.UUID
	CanAccess  bool
	ShowInMenu bool
}

// UpdateRolePage contains information needed to update a role-page mapping.
type UpdateRolePage struct {
	CanAccess  *bool
	ShowInMenu *bool
}

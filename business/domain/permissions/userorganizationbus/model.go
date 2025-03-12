package userorganizationbus

import (
	"time"

	"github.com/google/uuid"
)

type UserOrganization struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	OrganizationalUnitID uuid.UUID
	RoleID               uuid.UUID
	IsUnitManager        bool
	StartDate            time.Time
	EndDate              time.Time
	CreatedBy            uuid.UUID
	CreatedAt            time.Time
}

type NewUserOrganization struct {
	UserID               uuid.UUID
	OrganizationalUnitID uuid.UUID
	RoleID               uuid.UUID
	IsUnitManager        bool
	StartDate            time.Time
	EndDate              time.Time
	CreatedBy            uuid.UUID
}

type UpdateUserOrganization struct {
	UserID               *uuid.UUID
	OrganizationalUnitID *uuid.UUID
	RoleID               *uuid.UUID
	IsUnitManager        *bool
	StartDate            *time.Time
	EndDate              *time.Time
	CreatedBy            *uuid.UUID
}

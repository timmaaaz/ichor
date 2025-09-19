package userrolebus

import "github.com/google/uuid"

type UserRole struct {
	ID     uuid.UUID
	UserID uuid.UUID
	RoleID uuid.UUID
}

type NewUserRole struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

type UpdateUserRole struct {
	RoleID *uuid.UUID
}

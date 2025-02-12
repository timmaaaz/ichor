package userapprovalstatusbus

import "github.com/google/uuid"

type UserApprovalStatus struct {
	ID     uuid.UUID
	Name   string
	IconID uuid.UUID
}

type NewUserApprovalStatus struct {
	Name   string
	IconID uuid.UUID
}

type UpdateUserApprovalStatus struct {
	Name   *string
	IconID *uuid.UUID
}

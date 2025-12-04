package approvalbus

import "github.com/google/uuid"

type UserApprovalStatus struct {
	ID             uuid.UUID
	Name           string
	IconID         uuid.UUID
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type NewUserApprovalStatus struct {
	Name           string
	IconID         uuid.UUID
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type UpdateUserApprovalStatus struct {
	Name           *string
	IconID         *uuid.UUID
	PrimaryColor   *string
	SecondaryColor *string
	Icon           *string
}

package approvalstatusbus

import "github.com/google/uuid"

type ApprovalStatus struct {
	ID             uuid.UUID
	Name           string
	IconID         uuid.UUID
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type NewApprovalStatus struct {
	Name           string
	IconID         uuid.UUID
	PrimaryColor   string
	SecondaryColor string
	Icon           string
}

type UpdateApprovalStatus struct {
	Name           *string
	IconID         *uuid.UUID
	PrimaryColor   *string
	SecondaryColor *string
	Icon           *string
}

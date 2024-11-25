package approvalstatusbus

import "github.com/google/uuid"

type ApprovalStatus struct {
	ID     uuid.UUID
	Name   string
	IconID uuid.UUID
}

type NewApprovalStatus struct {
	Name   string
	IconID uuid.UUID
}

type UpdateApprovalStatus struct {
	Name   *string
	IconID *uuid.UUID
}

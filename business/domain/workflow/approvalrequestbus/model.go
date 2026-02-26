// Package approvalrequestbus provides business logic for workflow approval requests.
package approvalrequestbus

import (
	"time"

	"github.com/google/uuid"
)

// Status constants for approval request state.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusTimedOut = "timed_out"
	StatusExpired  = "expired"
)

// ApprovalType constants.
const (
	ApprovalTypeAny      = "any"
	ApprovalTypeAll      = "all"
	ApprovalTypeMajority = "majority"
)

// ApprovalRequest represents a workflow approval request in the system.
type ApprovalRequest struct {
	ID               uuid.UUID
	ExecutionID      uuid.UUID
	RuleID           uuid.UUID
	RuleName         string
	ActionName       string
	Approvers        []uuid.UUID
	ApprovalType     string
	Status           string
	TimeoutHours     int
	TaskToken        string
	ApprovalMessage  string
	ResolvedBy       *uuid.UUID
	ResolutionReason string
	CreatedDate      time.Time
	ResolvedDate     *time.Time
}

// NewApprovalRequest contains information needed to create a new approval request.
type NewApprovalRequest struct {
	ExecutionID     uuid.UUID
	RuleID          uuid.UUID
	ActionName      string
	Approvers       []uuid.UUID
	ApprovalType    string
	TimeoutHours    int
	TaskToken       string
	ApprovalMessage string
}

package approvalrequestbus

import "github.com/google/uuid"

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID          *uuid.UUID
	ExecutionID *uuid.UUID
	RuleID      *uuid.UUID
	Status      *string
	ApproverID  *uuid.UUID // Filter by approver (uses array contains)
}

package actionpermissionsbus

import "github.com/google/uuid"

// QueryFilter holds the available fields a query can be filtered on.
// Pointer semantics are used to distinguish "not provided" from "provided as zero value".
type QueryFilter struct {
	ID         *uuid.UUID
	RoleID     *uuid.UUID
	ActionType *string
	IsAllowed  *bool
}

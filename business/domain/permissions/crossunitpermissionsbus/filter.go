package crossunitpermissionsbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID              *uuid.UUID
	SourceUnitID    *uuid.UUID
	TargetUnitID    *uuid.UUID
	CanRead         *bool
	CanUpdate       *bool
	GrantedBy       *uuid.UUID
	StartValidFrom  *time.Time
	EndValidFrom    *time.Time
	StartValidUntil *time.Time
	EndValidUntil   *time.Time
	Reason          *string
}

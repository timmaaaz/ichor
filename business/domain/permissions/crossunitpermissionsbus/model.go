package crossunitpermissionsbus

import (
	"time"

	"github.com/google/uuid"
)

type CrossUnitPermission struct {
	ID           uuid.UUID
	SourceUnitID uuid.UUID
	TargetUnitID uuid.UUID
	CanRead      bool
	CanUpdate    bool
	GrantedBy    uuid.UUID
	ValidFrom    time.Time
	ValidUntil   time.Time
	Reason       string
}

type NewCrossUnitPermission struct {
	SourceUnitID uuid.UUID
	TargetUnitID uuid.UUID
	CanRead      bool
	CanUpdate    bool
	GrantedBy    uuid.UUID
	ValidFrom    time.Time
	ValidUntil   time.Time
	Reason       string
}

type UpdateCrossUnitPermission struct {
	SourceUnitID *uuid.UUID
	TargetUnitID *uuid.UUID
	CanRead      *bool
	CanUpdate    *bool
	GrantedBy    *uuid.UUID
	ValidFrom    *time.Time
	ValidUntil   *time.Time
	Reason       *string
}

package userbus

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID                 *uuid.UUID
	RequestedBy        *uuid.UUID
	UserApprovalStatus *uuid.UUID
	ApprovedBy         *uuid.UUID
	TitleID            *uuid.UUID
	OfficeID           *uuid.UUID
	Username           *Name
	FirstName          *Name
	LastName           *Name
	Email              *mail.Address
	Enabled            *bool

	// Date filters
	StartBirthday      *time.Time
	EndBirthday        *time.Time
	StartDateHired     *time.Time
	EndDateHired       *time.Time
	StartDateRequested *time.Time
	EndDateRequested   *time.Time
	StartDateApproved  *time.Time
	EndDateApproved    *time.Time
	StartCreatedDate   *time.Time
	EndCreatedDate     *time.Time
}

package userbus

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// User represents information about an individual user.
type User struct {
	ID            uuid.UUID
	RequestedBy   uuid.UUID
	ApprovedBy    uuid.UUID
	TitleID       uuid.UUID
	OfficeID      uuid.UUID
	WorkPhoneID   uuid.UUID
	CellPhoneID   uuid.UUID
	Username      Name
	FirstName     Name
	LastName      Name
	Email         mail.Address
	Birthday      time.Time
	Roles         []Role
	SystemRoles   []Role // TODO: address this data type
	PasswordHash  []byte
	Enabled       bool
	DateHired     time.Time
	DateRequested time.Time
	DateApproved  time.Time
	DateCreated   time.Time
	DateUpdated   time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	RequestedBy uuid.UUID
	TitleID     uuid.UUID
	OfficeID    uuid.UUID
	WorkPhoneID uuid.UUID
	CellPhoneID uuid.UUID
	Username    Name
	FirstName   Name
	LastName    Name
	Email       mail.Address
	Birthday    time.Time
	Roles       []Role
	SystemRoles []Role // TODO: address this data type
	Password    string
	Enabled     bool
}

// UpdateUser contains information needed to update a user. Fields that are not
// included are intended to have separate endpoints or permissions to update.
type UpdateUser struct {
	ApprovedBy   *uuid.UUID // Approval endpoint
	TitleID      *uuid.UUID
	OfficeID     *uuid.UUID
	WorkPhoneID  *uuid.UUID
	CellPhoneID  *uuid.UUID
	Username     *Name
	FirstName    *Name
	LastName     *Name
	Email        *mail.Address
	Birthday     *time.Time
	Roles        []Role // Separate endpoint
	SystemRoles  []Role // Separate endpoint
	Password     *string
	Enabled      *bool
	DateHired    *time.Time
	DateApproved *time.Time // Approval endpoint
}

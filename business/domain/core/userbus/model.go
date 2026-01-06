package userbus

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// User represents information about an individual user.
type User struct {
	ID                 uuid.UUID    `json:"id"`
	RequestedBy        uuid.UUID    `json:"requested_by"`
	ApprovedBy         uuid.UUID    `json:"approved_by"`
	UserApprovalStatus uuid.UUID    `json:"user_approval_status"`
	TitleID            uuid.UUID    `json:"title_id"`
	OfficeID           uuid.UUID    `json:"office_id"`
	WorkPhoneID        uuid.UUID    `json:"work_phone_id"`
	CellPhoneID        uuid.UUID    `json:"cell_phone_id"`
	Username           Name         `json:"username"`
	FirstName          Name         `json:"first_name"`
	LastName           Name         `json:"last_name"`
	Email              mail.Address `json:"email"`
	Birthday           time.Time    `json:"birthday"`
	Roles              []Role       `json:"roles"`
	SystemRoles        []Role       `json:"system_roles"` // TODO: address this data type
	PasswordHash       []byte       `json:"password_hash"`
	Enabled            bool         `json:"enabled"`
	DateHired          time.Time    `json:"date_hired"`
	DateRequested      time.Time    `json:"date_requested"`
	DateApproved       time.Time    `json:"date_approved"`
	CreatedDate        time.Time    `json:"created_date"`
	UpdatedDate        time.Time    `json:"updated_date"`
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	RequestedBy        uuid.UUID    `json:"requested_by"`
	TitleID            uuid.UUID    `json:"title_id"`
	OfficeID           uuid.UUID    `json:"office_id"`
	UserApprovalStatus uuid.UUID    `json:"user_approval_status"`
	WorkPhoneID        uuid.UUID    `json:"work_phone_id"`
	CellPhoneID        uuid.UUID    `json:"cell_phone_id"`
	Username           Name         `json:"username"`
	FirstName          Name         `json:"first_name"`
	LastName           Name         `json:"last_name"`
	Email              mail.Address `json:"email"`
	Birthday           time.Time    `json:"birthday"`
	Roles              []Role       `json:"roles"`
	SystemRoles        []Role       `json:"system_roles"` // TODO: address this data type
	Password           string       `json:"password"`
	Enabled            bool         `json:"enabled"`
}

// UpdateUser contains information needed to update a user. Fields that are not
// included are intended to have separate endpoints or permissions to update.
type UpdateUser struct {
	ApprovedBy         *uuid.UUID    `json:"approved_by,omitempty"` // Approval endpoint
	UserApprovalStatus *uuid.UUID    `json:"user_approval_status,omitempty"`
	TitleID            *uuid.UUID    `json:"title_id,omitempty"`
	OfficeID           *uuid.UUID    `json:"office_id,omitempty"`
	WorkPhoneID        *uuid.UUID    `json:"work_phone_id,omitempty"`
	CellPhoneID        *uuid.UUID    `json:"cell_phone_id,omitempty"`
	Username           *Name         `json:"username,omitempty"`
	FirstName          *Name         `json:"first_name,omitempty"`
	LastName           *Name         `json:"last_name,omitempty"`
	Email              *mail.Address `json:"email,omitempty"`
	Birthday           *time.Time    `json:"birthday,omitempty"`
	Roles              []Role        `json:"roles,omitempty"` // Separate endpoint
	SystemRoles        []Role        `json:"system_roles,omitempty"` // Separate endpoint
	Password           *string       `json:"password,omitempty"`
	Enabled            *bool         `json:"enabled,omitempty"`
	DateHired          *time.Time    `json:"date_hired,omitempty"`
	DateApproved       *time.Time    `json:"date_approved,omitempty"` // Approval endpoint
}

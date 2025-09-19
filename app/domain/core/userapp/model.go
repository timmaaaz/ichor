package userapp

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
)

const dateFormat = "2006-01-02"

// QueryParams represents the set of possible query strings.
type QueryParams struct {
	Page               string
	Rows               string
	OrderBy            string
	ID                 string
	RequestedBy        string
	ApprovedBy         string
	TitleID            string
	OfficeID           string
	Username           string
	FirstName          string
	LastName           string
	Email              string
	Enabled            string
	StartBirthday      string
	EndBirthday        string
	StartDateHired     string
	EndDateHired       string
	StartDateRequested string
	EndDateRequested   string
	StartDateApproved  string
	EndDateApproved    string
	StartCreatedDate   string
	EndCreatedDate     string
}

// =============================================================================

// User represents information about an individual user.
type User struct {
	ID            string   `json:"id"`
	RequestedBy   string   `json:"requested_by"`
	ApprovedBy    string   `json:"approved_by"`
	TitleID       string   `json:"title_id"`
	OfficeID      string   `json:"office_id"`
	WorkPhoneID   string   `json:"work_phone_id"`
	CellPhoneID   string   `json:"cell_phone_id"`
	Username      string   `json:"username"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Email         string   `json:"email"`
	Birthday      string   `json:"birthday"`
	Roles         []string `json:"roles"`
	SystemRoles   []string `json:"system_roles"`
	PasswordHash  []byte   `json:"-"`
	Enabled       bool     `json:"enabled"`
	DateHired     string   `json:"date_hired"`
	DateRequested string   `json:"date_requested"`
	DateApproved  string   `json:"date_approved"`
	CreatedDate   string   `json:"date_created"`
	UpdatedDate   string   `json:"date_updated"`
}

// Encode implements the encoder interface.
func (app User) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppUser(bus userbus.User) User {
	return User{
		ID:            bus.ID.String(),
		RequestedBy:   bus.RequestedBy.String(),
		ApprovedBy:    bus.ApprovedBy.String(),
		TitleID:       bus.TitleID.String(),
		OfficeID:      bus.OfficeID.String(),
		WorkPhoneID:   bus.WorkPhoneID.String(),
		CellPhoneID:   bus.CellPhoneID.String(),
		Username:      bus.Username.String(),
		FirstName:     bus.FirstName.String(),
		LastName:      bus.LastName.String(),
		Email:         bus.Email.Address,
		Birthday:      bus.Birthday.Format(time.RFC3339),
		Roles:         userbus.ParseRolesToString(bus.Roles),
		SystemRoles:   userbus.ParseRolesToString(bus.SystemRoles),
		PasswordHash:  bus.PasswordHash,
		Enabled:       bus.Enabled,
		DateHired:     bus.DateHired.Format(time.RFC3339),
		DateRequested: bus.DateRequested.Format(time.RFC3339),
		DateApproved:  bus.DateApproved.Format(time.RFC3339),
		CreatedDate:   bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate:   bus.UpdatedDate.Format(time.RFC3339),
	}
}

func toAppUsers(users []userbus.User) []User {
	app := make([]User, len(users))
	for i, usr := range users {
		app[i] = toAppUser(usr)
	}

	return app
}

// =============================================================================

// NewUser defines the data needed to add a new user.
type NewUser struct {
	RequestedBy     string   `json:"requested_by" validate:"omitempty"`
	TitleID         string   `json:"title_id" validate:"omitempty"`
	OfficeID        string   `json:"office_id" validate:"omitempty"`
	WorkPhoneID     string   `json:"work_phone_id" validate:"omitempty"`
	CellPhoneID     string   `json:"cell_phone_id" validate:"omitempty"`
	Username        string   `json:"username" validate:"required"`
	FirstName       string   `json:"first_name" validate:"required"`
	LastName        string   `json:"last_name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Birthday        string   `json:"birthday" validate:"required"`
	Roles           []string `json:"roles" validate:"required"`
	SystemRoles     []string `json:"system_roles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
	Enabled         bool     `json:"enabled"`
}

// Decode implements the decoder interface.
func (app *NewUser) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewUser) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewUser(app NewUser) (userbus.NewUser, error) {
	var requestedBy uuid.UUID
	var err error

	if app.RequestedBy != "" {
		requestedBy, err = uuid.Parse(app.RequestedBy)
		if err != nil {
			return userbus.NewUser{}, err
		}
	}

	var titleID uuid.UUID
	if app.TitleID != "" {
		titleID, err = uuid.Parse(app.TitleID)
		if err != nil {
			return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	var officeID uuid.UUID
	if app.OfficeID != "" {
		officeID, err = uuid.Parse(app.OfficeID)
		if err != nil {
			return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	var workPhoneID uuid.UUID
	if app.WorkPhoneID != "" {
		workPhoneID, err = uuid.Parse(app.WorkPhoneID)
		if err != nil {
			return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	var cellPhoneID uuid.UUID
	if app.CellPhoneID != "" {
		cellPhoneID, err = uuid.Parse(app.CellPhoneID)
		if err != nil {
			return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	username, err := userbus.ParseName(app.Username)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	firstName, err := userbus.ParseName(app.FirstName)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	lastName, err := userbus.ParseName(app.LastName)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	addr, err := mail.ParseAddress(app.Email)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	birthday, err := time.Parse(dateFormat, app.Birthday)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	roles, err := userbus.ParseRoles(app.Roles)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	systemRoles, err := userbus.ParseRoles(app.Roles)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	bus := userbus.NewUser{
		RequestedBy: requestedBy,
		TitleID:     titleID,
		OfficeID:    officeID,
		WorkPhoneID: workPhoneID,
		CellPhoneID: cellPhoneID,
		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       *addr,
		Birthday:    birthday,
		Roles:       roles,
		SystemRoles: systemRoles,
		Password:    app.Password,
		Enabled:     app.Enabled,
	}

	return bus, nil
}

// =============================================================================

// UpdateUserRole defines the data needed to update a user role.
type UpdateUserRole struct {
	Roles []string `json:"roles" validate:"required"`
}

// UpdateApproveUser defines the data needed to approve a user.
type UpdateApproveUser struct {
	ApprovedBy string `json:"approvedBy" validate:"required"`
}

// Decode implements the decoder interface.
func (app *UpdateUserRole) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateUserRole) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateUserRole(app UpdateUserRole) (userbus.UpdateUser, error) {
	var roles []userbus.Role
	if app.Roles != nil {
		var err error
		roles, err = userbus.ParseRoles(app.Roles)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	bus := userbus.UpdateUser{
		Roles: roles,
	}

	return bus, nil
}

func toBusUpdateUserSystemRole(app UpdateUserRole) (userbus.UpdateUser, error) {
	var roles []userbus.Role
	if app.Roles != nil {
		var err error
		roles, err = userbus.ParseRoles(app.Roles)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	bus := userbus.UpdateUser{
		SystemRoles: roles,
	}

	return bus, nil
}

func toBusUpdateApproveUser(app UpdateApproveUser) (userbus.UpdateUser, error) {
	approvedBy, err := uuid.Parse(app.ApprovedBy)
	if err != nil {
		return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
	}

	dateApproved, err := time.Parse(time.RFC3339, strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
	}

	bus := userbus.UpdateUser{
		ApprovedBy:   &approvedBy,
		DateApproved: &dateApproved,
	}

	return bus, nil
}

// =============================================================================

// UpdateUser defines the data needed to update a user.
type UpdateUser struct {
	TitleID         *string `json:"title_id"`
	OfficeID        *string `json:"office_id"`
	WorkPhoneID     *string `json:"work_phone_id"`
	CellPhoneID     *string `json:"cell_phone_id"`
	Username        *string `json:"username"`
	FirstName       *string `json:"first_name"`
	LastName        *string `json:"last_name"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Birthday        *string `json:"birthday"`
	Password        *string `json:"password"`
	PasswordConfirm *string `json:"password_confirm" validate:"omitempty,eqfield=Password"`
	Enabled         *bool   `json:"enabled"`
	DateHired       *string `json:"date_hired"`
}

// Decode implements the decoder interface.
func (app *UpdateUser) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateUser) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateUser(app UpdateUser) (userbus.UpdateUser, error) {
	var titleID *uuid.UUID
	if app.TitleID != nil {
		id, err := uuid.Parse(*app.TitleID)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		titleID = &id
	}

	var officeID *uuid.UUID
	if app.OfficeID != nil {
		id, err := uuid.Parse(*app.OfficeID)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		officeID = &id
	}

	var workPhoneID *uuid.UUID
	if app.WorkPhoneID != nil {
		id, err := uuid.Parse(*app.WorkPhoneID)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		workPhoneID = &id
	}

	var cellPhoneID *uuid.UUID
	if app.CellPhoneID != nil {
		id, err := uuid.Parse(*app.CellPhoneID)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		cellPhoneID = &id
	}

	var username *userbus.Name
	if app.Username != nil {
		nm, err := userbus.ParseName(*app.Username)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		username = &nm
	}

	var firstName *userbus.Name
	if app.FirstName != nil {
		nm, err := userbus.ParseName(*app.FirstName)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		firstName = &nm
	}

	var lastName *userbus.Name
	if app.LastName != nil {
		nm, err := userbus.ParseName(*app.LastName)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		lastName = &nm
	}

	var addr *mail.Address
	if app.Email != nil {
		var err error
		addr, err = mail.ParseAddress(*app.Email)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	var birthday *time.Time
	if app.Birthday != nil {
		tm, err := time.Parse(time.RFC3339, *app.Birthday)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		birthday = &tm
	}

	var dateHired *time.Time
	if app.DateHired != nil {
		tm, err := time.Parse(time.RFC3339, *app.DateHired)
		if err != nil {
			return userbus.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		dateHired = &tm
	}

	bus := userbus.UpdateUser{
		TitleID:     titleID,
		OfficeID:    officeID,
		WorkPhoneID: workPhoneID,
		CellPhoneID: cellPhoneID,
		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       addr,
		Birthday:    birthday,
		Password:    app.Password,
		Enabled:     app.Enabled,
		DateHired:   dateHired,
	}

	return bus, nil
}

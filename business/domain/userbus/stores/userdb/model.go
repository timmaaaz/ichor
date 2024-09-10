package userdb

import (
	"database/sql"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/userbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/dbarray"
)

type user struct {
	ID            uuid.UUID      `db:"user_id"`
	RequestedBy   uuid.UUID      `db:"requested_by"`
	ApprovedBy    uuid.UUID      `db:"approved_by"`
	TitleID       uuid.UUID      `db:"title_id"`
	OfficeID      uuid.UUID      `db:"office_id"`
	WorkPhoneID   uuid.UUID      `db:"work_phone_id"`
	CellPhoneID   uuid.UUID      `db:"cell_phone_id"`
	Username      string         `db:"username"`
	FirstName     string         `db:"first_name"`
	LastName      string         `db:"last_name"`
	Email         string         `db:"email"`
	Birthday      sql.NullTime   `db:"birthday"`
	Roles         dbarray.String `db:"roles"`
	SystemRoles   dbarray.String `db:"system_roles"`
	PasswordHash  []byte         `db:"password_hash"`
	Enabled       bool           `db:"enabled"`
	DateHired     sql.NullTime   `db:"date_hired"`
	DateRequested sql.NullTime   `db:"date_requested"`
	DateApproved  sql.NullTime   `db:"date_approved"`
	DateCreated   time.Time      `db:"date_created"`
	DateUpdated   time.Time      `db:"date_updated"`
}

func toDBUser(bus userbus.User) user {

	var birthday sql.NullTime
	if bus.Birthday.IsZero() {
		birthday = sql.NullTime{}
	} else {
		birthday = sql.NullTime{Time: bus.Birthday.UTC(), Valid: true}
	}

	var dateHired sql.NullTime
	if bus.DateHired.IsZero() {
		dateHired = sql.NullTime{}
	} else {
		dateHired = sql.NullTime{Time: bus.DateHired.UTC(), Valid: true}
	}

	var dateRequested sql.NullTime
	if bus.DateRequested.IsZero() {
		dateRequested = sql.NullTime{}
	} else {
		dateRequested = sql.NullTime{Time: bus.DateRequested.UTC(), Valid: true}
	}

	var dateApproved sql.NullTime
	if bus.DateApproved.IsZero() {
		dateApproved = sql.NullTime{}
	} else {
		dateApproved = sql.NullTime{Time: bus.DateApproved.UTC(), Valid: true}
	}

	return user{
		ID:            bus.ID,
		RequestedBy:   bus.RequestedBy,
		ApprovedBy:    bus.ApprovedBy,
		TitleID:       bus.TitleID,
		OfficeID:      bus.OfficeID,
		WorkPhoneID:   bus.WorkPhoneID,
		CellPhoneID:   bus.CellPhoneID,
		Username:      bus.Username.String(),
		FirstName:     bus.FirstName.String(),
		LastName:      bus.LastName.String(),
		Email:         bus.Email.Address,
		Birthday:      birthday,
		Roles:         userbus.ParseRolesToString(bus.Roles),
		SystemRoles:   userbus.ParseRolesToString(bus.SystemRoles),
		PasswordHash:  bus.PasswordHash,
		Enabled:       bus.Enabled,
		DateHired:     dateHired,
		DateRequested: dateRequested,
		DateApproved:  dateApproved,
		DateCreated:   bus.DateCreated.UTC(),
		DateUpdated:   bus.DateUpdated.UTC(),
	}
}

func toBusUser(db user) (userbus.User, error) {
	email := mail.Address{
		Address: db.Email,
	}

	roles, err := userbus.ParseRoles(db.Roles)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse roles: %w", err)
	}
	systemRoles, err := userbus.ParseRoles(db.SystemRoles)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse system roles: %w", err)
	}

	username, err := userbus.ParseName(db.Username)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse username: %w", err)
	}
	firstName, err := userbus.ParseName(db.FirstName)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse first name: %w", err)
	}
	lastName, err := userbus.ParseName(db.LastName)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse last name: %w", err)
	}

	var birthday time.Time
	if db.Birthday.Valid && !db.Birthday.Time.IsZero() {
		birthday = db.Birthday.Time.In(time.Local)
	} else {
		birthday = time.Time{}
	}

	var dateHired time.Time
	if db.DateHired.Valid && !db.DateHired.Time.IsZero() {
		dateHired = db.DateHired.Time.In(time.Local)
	} else {
		dateHired = time.Time{}
	}

	var dateRequested time.Time
	if db.DateRequested.Valid && !db.DateRequested.Time.IsZero() {
		dateRequested = db.DateRequested.Time.In(time.Local)
	} else {
		dateRequested = time.Time{}
	}

	var dateApproved time.Time
	if db.DateApproved.Valid && !db.DateApproved.Time.IsZero() {
		dateApproved = db.DateApproved.Time.In(time.Local)
	} else {
		dateApproved = time.Time{}
	}

	bus := userbus.User{
		ID:            db.ID,
		RequestedBy:   db.RequestedBy,
		ApprovedBy:    db.ApprovedBy,
		TitleID:       db.TitleID,
		OfficeID:      db.OfficeID,
		WorkPhoneID:   db.WorkPhoneID,
		CellPhoneID:   db.CellPhoneID,
		Username:      username,
		FirstName:     firstName,
		LastName:      lastName,
		Email:         email,
		Birthday:      birthday,
		Roles:         roles,
		SystemRoles:   systemRoles,
		PasswordHash:  db.PasswordHash,
		Enabled:       db.Enabled,
		DateHired:     dateHired,
		DateRequested: dateRequested,
		DateApproved:  dateApproved,
		DateCreated:   db.DateCreated.In(time.Local),
		DateUpdated:   db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

func toBusUsers(dbs []user) ([]userbus.User, error) {
	bus := make([]userbus.User, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = toBusUser(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}

package contactinfosdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
)

type contactInfo struct {
	ID                   uuid.UUID `db:"id"`
	FirstName            string    `db:"first_name"`
	LastName             string    `db:"last_name"`
	PrimaryPhone         string    `db:"primary_phone_number"`
	SecondaryPhone       string    `db:"secondary_phone_number"`
	Email                string    `db:"email_address"`
	Address              string    `db:"address"`
	AvailableHoursStart  string    `db:"available_hours_start"`
	AvailableHoursEnd    string    `db:"available_hours_end"`
	Timezone             string    `db:"timezone"`
	PreferredContactType string    `db:"preferred_contact_type"`
	Notes                string    `db:"notes"`
}

func toDBContactInfo(bus contactinfosbus.ContactInfo) contactInfo {
	return contactInfo{
		ID:                   bus.ID,
		FirstName:            bus.FirstName,
		LastName:             bus.LastName,
		PrimaryPhone:         bus.PrimaryPhone,
		Email:                bus.EmailAddress,
		Address:              bus.Address,
		SecondaryPhone:       bus.SecondaryPhone,
		AvailableHoursStart:  bus.AvailableHoursStart,
		AvailableHoursEnd:    bus.AvailableHoursEnd,
		Timezone:             bus.Timezone,
		PreferredContactType: bus.PreferredContactType,
		Notes:                bus.Notes,
	}
}

func toBusContactInfo(db contactInfo) contactinfosbus.ContactInfo {
	return contactinfosbus.ContactInfo{
		ID:                   db.ID,
		FirstName:            db.FirstName,
		LastName:             db.LastName,
		PrimaryPhone:         db.PrimaryPhone,
		EmailAddress:         db.Email,
		Address:              db.Address,
		SecondaryPhone:       db.SecondaryPhone,
		AvailableHoursStart:  db.AvailableHoursStart,
		AvailableHoursEnd:    db.AvailableHoursEnd,
		Timezone:             db.Timezone,
		PreferredContactType: db.PreferredContactType,
		Notes:                db.Notes,
	}
}

func toBusContactInfos(dbs []contactInfo) []contactinfosbus.ContactInfo {
	bus := make([]contactinfosbus.ContactInfo, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusContactInfo(db)
	}
	return bus
}

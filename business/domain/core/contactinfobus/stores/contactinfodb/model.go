package contactinfodb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
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

func toDBContactInfo(bus contactinfobus.ContactInfo) contactInfo {
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

func toBusContactInfo(db contactInfo) contactinfobus.ContactInfo {
	return contactinfobus.ContactInfo{
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

func toBusContactInfos(dbs []contactInfo) []contactinfobus.ContactInfo {
	bus := make([]contactinfobus.ContactInfo, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusContactInfo(db)
	}
	return bus
}

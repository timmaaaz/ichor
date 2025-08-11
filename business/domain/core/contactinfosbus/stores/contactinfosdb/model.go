package contactinfosdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
)

type contactInfos struct {
	ID                   uuid.UUID `db:"id"`
	FirstName            string    `db:"first_name"`
	LastName             string    `db:"last_name"`
	PrimaryPhone         string    `db:"primary_phone_number"`
	SecondaryPhone       string    `db:"secondary_phone_number"`
	Email                string    `db:"email_address"`
	StreetID             uuid.UUID `db:"street_id"`
	DeliveryAddressID    uuid.UUID `db:"delivery_address_id"`
	AvailableHoursStart  string    `db:"available_hours_start"`
	AvailableHoursEnd    string    `db:"available_hours_end"`
	Timezone             string    `db:"timezone"`
	PreferredContactType string    `db:"preferred_contact_type"`
	Notes                string    `db:"notes"`
}

func toDBContactInfos(bus contactinfosbus.ContactInfos) contactInfos {
	return contactInfos{
		ID:                   bus.ID,
		FirstName:            bus.FirstName,
		LastName:             bus.LastName,
		PrimaryPhone:         bus.PrimaryPhone,
		Email:                bus.EmailAddress,
		StreetID:             bus.StreetID,
		DeliveryAddressID:    bus.DeliveryAddressID,
		SecondaryPhone:       bus.SecondaryPhone,
		AvailableHoursStart:  bus.AvailableHoursStart,
		AvailableHoursEnd:    bus.AvailableHoursEnd,
		Timezone:             bus.Timezone,
		PreferredContactType: bus.PreferredContactType,
		Notes:                bus.Notes,
	}
}

func toBusContactInfos(db contactInfos) contactinfosbus.ContactInfos {
	return contactinfosbus.ContactInfos{
		ID:                   db.ID,
		FirstName:            db.FirstName,
		LastName:             db.LastName,
		PrimaryPhone:         db.PrimaryPhone,
		EmailAddress:         db.Email,
		StreetID:             db.StreetID,
		DeliveryAddressID:    db.DeliveryAddressID,
		SecondaryPhone:       db.SecondaryPhone,
		AvailableHoursStart:  db.AvailableHoursStart,
		AvailableHoursEnd:    db.AvailableHoursEnd,
		Timezone:             db.Timezone,
		PreferredContactType: db.PreferredContactType,
		Notes:                db.Notes,
	}
}

func toBusContactInfoss(dbs []contactInfos) []contactinfosbus.ContactInfos {
	bus := make([]contactinfosbus.ContactInfos, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusContactInfos(db)
	}
	return bus
}

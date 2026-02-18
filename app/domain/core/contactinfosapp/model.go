package contactinfosapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/foundation/timeutil/timeonly"
)

type QueryParams struct {
	Page                 string
	Rows                 string
	OrderBy              string
	ID                   string
	FirstName            string
	LastName             string
	EmailAddress         string
	PrimaryPhone         string
	SecondaryPhone       string
	StreetID             string
	DeliveryAddressID    string
	AvailableHoursStart  string
	AvailableHoursEnd    string
	TimezoneID           string
	PreferredContactType string
	Notes                string
}

type ContactInfos struct {
	ID                   string `json:"id"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	EmailAddress         string `json:"email_address"`
	PrimaryPhone         string `json:"primary_phone_number"`
	SecondaryPhone       string `json:"secondary_phone_number"`
	StreetID             string `json:"street_id"`
	DeliveryAddressID    string `json:"delivery_address_id"`
	AvailableHoursStart  string `json:"available_hours_start"`
	AvailableHoursEnd    string `json:"available_hours_end"`
	TimezoneID           string `json:"timezone_id"`
	PreferredContactType string `json:"preferred_contact_type"`
	Notes                string `json:"notes"`
}

func (app ContactInfos) Encode() ([]byte, string, error) {

	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppContactInfo(bus contactinfosbus.ContactInfos) ContactInfos {
	return ContactInfos{
		ID:                   bus.ID.String(),
		FirstName:            bus.FirstName,
		LastName:             bus.LastName,
		EmailAddress:         bus.EmailAddress,
		PrimaryPhone:         bus.PrimaryPhone,
		SecondaryPhone:       bus.SecondaryPhone,
		StreetID:             bus.StreetID.String(),
		DeliveryAddressID:    bus.DeliveryAddressID.String(),
		AvailableHoursStart:  bus.AvailableHoursStart,
		AvailableHoursEnd:    bus.AvailableHoursEnd,
		TimezoneID:           bus.TimezoneID.String(),
		PreferredContactType: bus.PreferredContactType,
		Notes:                bus.Notes,
	}
}

func ToAppContactInfos(bus []contactinfosbus.ContactInfos) []ContactInfos {
	app := make([]ContactInfos, len(bus))
	for i, v := range bus {
		app[i] = ToAppContactInfo(v)
	}
	return app
}

type NewContactInfos struct {
	FirstName            string `json:"first_name" validate:"required"`
	LastName             string `json:"last_name" validate:"required"`
	EmailAddress         string `json:"email_address" validate:"required"`
	PrimaryPhone         string `json:"primary_phone_number" validate:"required"`
	SecondaryPhone       string `json:"secondary_phone_number"`
	StreetID             string `json:"street_id" validate:"required"`
	DeliveryAddressID    string `json:"delivery_address_id" validate:"omitempty"`
	AvailableHoursStart  string `json:"available_hours_start" validate:"required"`
	AvailableHoursEnd    string `json:"available_hours_end" validate:"required"`
	TimezoneID           string `json:"timezone_id" validate:"required"`
	PreferredContactType string `json:"preferred_contact_type" validate:"required"`
	Notes                string `json:"notes"`
}

func (app *NewContactInfos) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewContactInfos) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewContactInfos(app NewContactInfos) (contactinfosbus.NewContactInfos, error) {
	if !timeonly.ValidateTimeOnlyFmt(app.AvailableHoursEnd) {
		return contactinfosbus.NewContactInfos{}, fmt.Errorf("invalid time format for ending hours: %q", app.AvailableHoursEnd)
	}

	if !timeonly.ValidateTimeOnlyFmt(app.AvailableHoursStart) {
		return contactinfosbus.NewContactInfos{}, fmt.Errorf("invalid time format for starting hours: %q", app.AvailableHoursStart)
	}

	streetID, err := uuid.Parse(app.StreetID)
	if err != nil {
		return contactinfosbus.NewContactInfos{}, errs.Newf(errs.InvalidArgument, "parse street_id: %s", err)
	}

	deliveryAddressID, err := uuid.Parse(app.DeliveryAddressID)
	if err != nil {
		return contactinfosbus.NewContactInfos{}, errs.Newf(errs.InvalidArgument, "parse delivery_address_id: %s", err)
	}

	timezoneID, err := uuid.Parse(app.TimezoneID)
	if err != nil {
		return contactinfosbus.NewContactInfos{}, errs.Newf(errs.InvalidArgument, "parse timezone_id: %s", err)
	}

	bus := contactinfosbus.NewContactInfos{
		FirstName:            app.FirstName,
		LastName:             app.LastName,
		EmailAddress:         app.EmailAddress,
		PrimaryPhone:         app.PrimaryPhone,
		SecondaryPhone:       app.SecondaryPhone,
		StreetID:             streetID,
		DeliveryAddressID:    deliveryAddressID,
		AvailableHoursStart:  app.AvailableHoursStart,
		AvailableHoursEnd:    app.AvailableHoursEnd,
		TimezoneID:           timezoneID,
		PreferredContactType: app.PreferredContactType,
		Notes:                app.Notes,
	}

	return bus, nil
}

type UpdateContactInfos struct {
	FirstName            *string `json:"first_name"`
	LastName             *string `json:"last_name"`
	EmailAddress         *string `json:"email_address"`
	PrimaryPhone         *string `json:"primary_phone_number"`
	SecondaryPhone       *string `json:"secondary_phone_number"`
	StreetID             *string `json:"street_id"`
	DeliveryAddressID    *string `json:"delivery_address_id"`
	AvailableHoursStart  *string `json:"available_hours_start"`
	AvailableHoursEnd    *string `json:"available_hours_end"`
	TimezoneID           *string `json:"timezone_id"`
	PreferredContactType *string `json:"preferred_contact_type"`
	Notes                *string `json:"notes"`
}

// Decode implements the decoder interface.
func (app *UpdateContactInfos) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateContactInfos) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateContactInfos(app UpdateContactInfos) (contactinfosbus.UpdateContactInfos, error) {
	if app.AvailableHoursEnd != nil && !timeonly.ValidateTimeOnlyFmt(*app.AvailableHoursEnd) {
		return contactinfosbus.UpdateContactInfos{}, fmt.Errorf("invalid time format for ending hours: %q", *app.AvailableHoursEnd)
	}

	if app.AvailableHoursStart != nil && !timeonly.ValidateTimeOnlyFmt(*app.AvailableHoursStart) {
		return contactinfosbus.UpdateContactInfos{}, fmt.Errorf("invalid time format for starting hours: %q", *app.AvailableHoursStart)
	}

	bus := contactinfosbus.UpdateContactInfos{}

	if app.StreetID != nil {
		id, err := uuid.Parse(*app.StreetID)
		if err != nil {
			return contactinfosbus.UpdateContactInfos{}, errs.Newf(errs.InvalidArgument, "parse street_id: %s", err)
		}
		bus.StreetID = &id
	}

	if app.DeliveryAddressID != nil {
		id, err := uuid.Parse(*app.DeliveryAddressID)
		if err != nil {
			return contactinfosbus.UpdateContactInfos{}, errs.Newf(errs.InvalidArgument, "parse delivery_address_id: %s", err)
		}
		bus.DeliveryAddressID = &id
	}

	if app.TimezoneID != nil {
		id, err := uuid.Parse(*app.TimezoneID)
		if err != nil {
			return contactinfosbus.UpdateContactInfos{}, errs.Newf(errs.InvalidArgument, "parse timezone_id: %s", err)
		}
		bus.TimezoneID = &id
	}

	bus.FirstName = app.FirstName
	bus.LastName = app.LastName
	bus.EmailAddress = app.EmailAddress
	bus.PrimaryPhone = app.PrimaryPhone
	bus.SecondaryPhone = app.SecondaryPhone
	bus.AvailableHoursStart = app.AvailableHoursStart
	bus.AvailableHoursEnd = app.AvailableHoursEnd
	bus.PreferredContactType = app.PreferredContactType
	bus.Notes = app.Notes

	return bus, nil
}

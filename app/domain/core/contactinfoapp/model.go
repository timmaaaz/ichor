package contactinfoapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/foundation/convert"
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
	Address              string
	AvailableHoursStart  string
	AvailableHoursEnd    string
	Timezone             string
	PreferredContactType string
	Notes                string
}

type ContactInfo struct {
	ID                   string `json:"id"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	EmailAddress         string `json:"email_address"`
	PrimaryPhone         string `json:"primary_phone"`
	SecondaryPhone       string `json:"secondary_phone"`
	Address              string `json:"address"`
	AvailableHoursStart  string `json:"available_hours_start"`
	AvailableHoursEnd    string `json:"available_hours_end"`
	Timezone             string `json:"timezone"`
	PreferredContactType string `json:"preferred_contact_type"`
	Notes                string `json:"notes"`
}

func (app ContactInfo) Encode() ([]byte, string, error) {

	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppContactInfo(bus contactinfobus.ContactInfo) ContactInfo {
	return ContactInfo{
		ID:                   bus.ID.String(),
		FirstName:            bus.FirstName,
		LastName:             bus.LastName,
		EmailAddress:         bus.EmailAddress,
		PrimaryPhone:         bus.PrimaryPhone,
		SecondaryPhone:       bus.SecondaryPhone,
		Address:              bus.Address,
		AvailableHoursStart:  bus.AvailableHoursStart,
		AvailableHoursEnd:    bus.AvailableHoursEnd,
		Timezone:             bus.Timezone,
		PreferredContactType: bus.PreferredContactType,
		Notes:                bus.Notes,
	}
}

func ToAppContactInfos(bus []contactinfobus.ContactInfo) []ContactInfo {
	app := make([]ContactInfo, len(bus))
	for i, v := range bus {
		app[i] = ToAppContactInfo(v)
	}
	return app
}

type NewContactInfo struct {
	FirstName            string `json:"first_name" validate:"required"`
	LastName             string `json:"last_name" validate:"required"`
	EmailAddress         string `json:"email_address" validate:"required"`
	PrimaryPhone         string `json:"primary_phone" validate:"required"`
	SecondaryPhone       string `json:"secondary_phone"`
	Address              string `json:"address" validate:"required"`
	AvailableHoursStart  string `json:"available_hours_start" validate:"required"`
	AvailableHoursEnd    string `json:"available_hours_end" validate:"required"`
	Timezone             string `json:"timezone" validate:"required"`
	PreferredContactType string `json:"preferred_contact_type" validate:"required"`
	Notes                string `json:"notes"`
}

func (app *NewContactInfo) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewContactInfo) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewContactInfo(app NewContactInfo) (contactinfobus.NewContactInfo, error) {
	dest := contactinfobus.NewContactInfo{}

	if !timeonly.ValidateTimeOnlyFmt(app.AvailableHoursEnd) {
		return contactinfobus.NewContactInfo{}, fmt.Errorf("invalid time format for ending hours: %q", app.AvailableHoursEnd)
	}

	if !timeonly.ValidateTimeOnlyFmt(app.AvailableHoursStart) {
		return contactinfobus.NewContactInfo{}, fmt.Errorf("invalid time format for starting hours: %q", app.AvailableHoursStart)
	}

	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

type UpdateContactInfo struct {
	FirstName            *string `json:"first_name"`
	LastName             *string `json:"last_name"`
	EmailAddress         *string `json:"email_address"`
	PrimaryPhone         *string `json:"primary_phone"`
	SecondaryPhone       *string `json:"secondary_phone"`
	Address              *string `json:"address"`
	AvailableHoursStart  *string `json:"available_hours_start"`
	AvailableHoursEnd    *string `json:"available_hours_end"`
	Timezone             *string `json:"timezone"`
	PreferredContactType *string `json:"preferred_contact_type"`
	Notes                *string `json:"notes"`
}

// Decode implements the decoder interface.
func (app *UpdateContactInfo) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateContactInfo) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateContactInfo(app UpdateContactInfo) (contactinfobus.UpdateContactInfo, error) {
	dest := contactinfobus.UpdateContactInfo{}

	if app.AvailableHoursEnd != nil && !timeonly.ValidateTimeOnlyFmt(*app.AvailableHoursEnd) {
		return contactinfobus.UpdateContactInfo{}, fmt.Errorf("invalid time format for ending hours: %q", *app.AvailableHoursEnd)
	}

	if app.AvailableHoursStart != nil && !timeonly.ValidateTimeOnlyFmt(*app.AvailableHoursStart) {
		return contactinfobus.UpdateContactInfo{}, fmt.Errorf("invalid time format for starting hours: %q", *app.AvailableHoursStart)
	}

	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}

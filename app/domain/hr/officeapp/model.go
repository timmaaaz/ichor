package officeapp

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
)

type QueryParams struct {
	Page     string
	Rows     string
	OrderBy  string
	ID       string
	Name     string
	StreetID string
}

// =============================================================================

type Office struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	StreetID string `json:"street_id"`
}

func (app Office) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppOffice(bus officebus.Office) Office {
	return Office{
		ID:       bus.ID.String(),
		Name:     bus.Name,
		StreetID: bus.StreetID.String(),
	}
}

func ToAppOffices(bus []officebus.Office) []Office {
	app := make([]Office, len(bus))
	for i, v := range bus {
		app[i] = ToAppOffice(v)
	}
	return app
}

// =============================================================================

type NewOffice struct {
	Name     string `json:"name" validate:"required"`
	StreetID string `json:"street_id" validate:"required"`
}

// Decode implements the decoder interface.
func (app *NewOffice) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewOffice) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewOffice(app NewOffice) (officebus.NewOffice, error) {
	var streetID uuid.UUID
	var err error

	if streetID, err = uuid.Parse(app.StreetID); err != nil {
		return officebus.NewOffice{}, err
	}

	return officebus.NewOffice{
		StreetID: streetID,
		Name:     app.Name,
	}, nil
}

// =============================================================================

type UpdateOffice struct {
	Name     *string `json:"name" validate:"required,min=3"`
	StreetID *string `json:"street_id" validate:"required,min=36,max=36"`
}

func (app *UpdateOffice) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateOffice) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateOffice(app UpdateOffice) (officebus.UpdateOffice, error) {
	var streetID *uuid.UUID
	var name *string

	if app.StreetID != nil {
		id, err := uuid.Parse(*app.StreetID)
		if err != nil {
			return officebus.UpdateOffice{}, err
		}

		streetID = &id
	}

	if app.Name != nil {
		name = app.Name
	}

	return officebus.UpdateOffice{
		Name:     name,
		StreetID: streetID,
	}, nil
}

package streetapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
)

// QueryParams represents the query parameters that can be used.
type QueryParams struct {
	Page       string
	Rows       string
	OrderBy    string
	ID         string
	CityID     string
	Line1      string
	Line2      string
	PostalCode string
}

type Street struct {
	ID         string `json:"id"`
	CityID     string `json:"city_id"`
	Line1      string `json:"line_1"`
	Line2      string `json:"line_2"`
	PostalCode string `json:"postal_code"`
}

// Encode implements the encoder interface.
func (app Street) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppStreet(bus streetbus.Street) Street {
	return Street{
		ID:         bus.ID.String(),
		CityID:     bus.CityID.String(),
		Line1:      bus.Line1,
		Line2:      bus.Line2,
		PostalCode: bus.PostalCode,
	}
}

func ToAppStreets(bus []streetbus.Street) []Street {
	app := make([]Street, len(bus))
	for i, v := range bus {
		app[i] = ToAppStreet(v)
	}
	return app
}

// =============================================================================

// NewStreet defines the data needed to add a street.
type NewStreet struct {
	CityID     string `json:"city_id" validate:"required"`
	Line1      string `json:"line_1" validate:"required,min=3,max=100"`
	Line2      string `json:"line_2" validate:"required,min=3,max=100"`
	PostalCode string `json:"postal_code" validate:"required,min=3,max=20"`
}

// Decode implements the decoder interface.
func (app *NewStreet) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewStreet) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewStreet(app NewStreet) (streetbus.NewStreet, error) {
	var cityID uuid.UUID
	var err error
	if cityID, err = uuid.Parse(app.CityID); err != nil {
		return streetbus.NewStreet{}, errs.Newf(errs.InvalidArgument, "parse: %s", err)
	}

	return streetbus.NewStreet{
		CityID:     cityID,
		Line1:      app.Line1,
		Line2:      app.Line2,
		PostalCode: app.PostalCode,
	}, nil
}

// =============================================================================

// UpdateStreet defines the data needed to update a street.
type UpdateStreet struct {
	CityID     *string `json:"city_id"`
	Line1      *string `json:"line_1"`
	Line2      *string `json:"line_2"`
	PostalCode *string `json:"postal_code"`
}

// Decode implements the decoder interface.
func (app *UpdateStreet) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateStreet) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateStreet(app UpdateStreet) (streetbus.UpdateStreet, error) {
	var cityID *uuid.UUID

	if app.CityID != nil {
		if id, err := uuid.Parse(*app.CityID); err != nil {
			return streetbus.UpdateStreet{}, errs.Newf(errs.InvalidArgument, "parse: %s", err)
		} else {
			cityID = &id
		}
	}

	return streetbus.UpdateStreet{
		CityID:     cityID,
		Line1:      app.Line1,
		Line2:      app.Line2,
		PostalCode: app.PostalCode,
	}, nil
}

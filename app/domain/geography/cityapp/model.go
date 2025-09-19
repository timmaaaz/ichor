package cityapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
)

// QueryParams represents the query parameters that can be used.
type QueryParams struct {
	Page     string
	Rows     string
	OrderBy  string
	ID       string
	RegionID string
	Name     string
}

type City struct {
	ID       string `json:"id"`
	RegionID string `json:"region_id"`
	Name     string `json:"name"`
}

// Encode implements the encoder interface.
func (app City) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppCity(bus citybus.City) City {
	return City{
		ID:       bus.ID.String(),
		RegionID: bus.RegionID.String(),
		Name:     bus.Name,
	}
}

func ToAppCities(bus []citybus.City) []City {
	app := make([]City, len(bus))
	for i, v := range bus {
		app[i] = ToAppCity(v)
	}
	return app
}

// =============================================================================

// NewCity defines the data needed to add a city.
type NewCity struct {
	RegionID string `json:"regionID" validate:"required"`
	Name     string `json:"name" validate:"required,min=3,max=100"`
}

// Decode implements the decoder interface.
func (app *NewCity) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewCity) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewCity(app NewCity) (citybus.NewCity, error) {
	var regionID uuid.UUID
	var err error

	if regionID, err = uuid.Parse(app.RegionID); err != nil {
		return citybus.NewCity{}, err
	}

	return citybus.NewCity{
		RegionID: regionID,
		Name:     app.Name, // TODO: Look at defining custom type
	}, nil
}

// =============================================================================

// UpdateCity defines the data needed to update a city.
type UpdateCity struct {
	RegionID *string `json:"regionID"`
	Name     *string `json:"name"`
}

// Decode implements the decoder interface.
func (app *UpdateCity) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateCity) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateCity(app UpdateCity) (citybus.UpdateCity, error) {
	var regionID *uuid.UUID

	if app.RegionID != nil {
		tmp, err := uuid.Parse(*app.RegionID)
		if err != nil {
			return citybus.UpdateCity{}, err
		}
		regionID = &tmp
	}

	return citybus.UpdateCity{
		RegionID: regionID,
		Name:     app.Name, // TODO: Look at defining custom type
	}, nil
}

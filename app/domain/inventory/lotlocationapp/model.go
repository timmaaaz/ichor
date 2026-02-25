package lotlocationapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page       string
	Rows       string
	OrderBy    string
	ID         string
	LotID      string
	LocationID string
}

type LotLocation struct {
	ID          string `json:"id"`
	LotID       string `json:"lot_id"`
	LocationID  string `json:"location_id"`
	Quantity    string `json:"quantity"`
	CreatedDate string `json:"created_date"`
	UpdatedDate string `json:"updated_date"`
}

func (app LotLocation) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppLotLocation(bus lotlocationbus.LotLocation) LotLocation {
	return LotLocation{
		ID:          bus.ID.String(),
		LotID:       bus.LotID.String(),
		LocationID:  bus.LocationID.String(),
		Quantity:    fmt.Sprintf("%d", bus.Quantity),
		CreatedDate: bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate: bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppLotLocations(bus []lotlocationbus.LotLocation) []LotLocation {
	app := make([]LotLocation, len(bus))
	for i, v := range bus {
		app[i] = ToAppLotLocation(v)
	}
	return app
}

type NewLotLocation struct {
	LotID      string `json:"lot_id" validate:"required,min=36,max=36"`
	LocationID string `json:"location_id" validate:"required,min=36,max=36"`
	Quantity   string `json:"quantity" validate:"required"`
}

func (app *NewLotLocation) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewLotLocation) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func toBusNewLotLocation(app NewLotLocation) (lotlocationbus.NewLotLocation, error) {
	lotID, err := uuid.Parse(app.LotID)
	if err != nil {
		return lotlocationbus.NewLotLocation{}, errs.Newf(errs.InvalidArgument, "parse lotID: %s", err)
	}

	locationID, err := uuid.Parse(app.LocationID)
	if err != nil {
		return lotlocationbus.NewLotLocation{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return lotlocationbus.NewLotLocation{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
	}

	return lotlocationbus.NewLotLocation{
		LotID:      lotID,
		LocationID: locationID,
		Quantity:   quantity,
	}, nil
}

type UpdateLotLocation struct {
	LotID      *string `json:"lot_id" validate:"omitempty,min=36,max=36"`
	LocationID *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	Quantity   *string `json:"quantity"`
}

func (app *UpdateLotLocation) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateLotLocation) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func toBusUpdateLotLocation(app UpdateLotLocation) (lotlocationbus.UpdateLotLocation, error) {
	bus := lotlocationbus.UpdateLotLocation{}

	if app.LotID != nil {
		lotID, err := uuid.Parse(*app.LotID)
		if err != nil {
			return lotlocationbus.UpdateLotLocation{}, errs.Newf(errs.InvalidArgument, "parse lotID: %s", err)
		}
		bus.LotID = &lotID
	}

	if app.LocationID != nil {
		locationID, err := uuid.Parse(*app.LocationID)
		if err != nil {
			return lotlocationbus.UpdateLotLocation{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
		}
		bus.LocationID = &locationID
	}

	if app.Quantity != nil {
		quantity, err := strconv.Atoi(*app.Quantity)
		if err != nil {
			return lotlocationbus.UpdateLotLocation{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
		}
		bus.Quantity = &quantity
	}

	return bus, nil
}

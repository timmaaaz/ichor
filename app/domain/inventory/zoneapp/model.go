package zoneapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ZoneID      string
	WarehouseID string
	Name        string
	Description string
	Stage       string
	CreatedDate string
	UpdatedDate string
}

type Zone struct {
	ZoneID      string `json:"zone_id"`
	WarehouseID string `json:"warehouse_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Stage       string `json:"stage"`
	CreatedDate string `json:"created_date"`
	UpdatedDate string `json:"updated_date"`
}

func (app Zone) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppZone(bus zonebus.Zone) Zone {
	app := Zone{
		ZoneID:      bus.ZoneID.String(),
		WarehouseID: bus.WarehouseID.String(),
		Name:        bus.Name,
		Description: bus.Description,
		CreatedDate: bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate: bus.UpdatedDate.Format(timeutil.FORMAT),
	}
	if bus.Stage != nil {
		app.Stage = bus.Stage.String()
	}
	return app
}

func ToAppZones(zones []zonebus.Zone) []Zone {
	app := make([]Zone, len(zones))
	for i, z := range zones {
		app[i] = ToAppZone(z)
	}
	return app
}

type NewZone struct {
	WarehouseID string `json:"warehouse_id" validate:"required,min=36,max=36"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty"`
	Stage       string `json:"stage" validate:"omitempty"`
}

func (app *NewZone) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewZone) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewZone(app NewZone) (zonebus.NewZone, error) {

	warehouseID, err := uuid.Parse(app.WarehouseID)
	if err != nil {
		return zonebus.NewZone{}, err
	}

	dest := zonebus.NewZone{
		WarehouseID: warehouseID,
		Name:        app.Name,
		Description: app.Description,
	}

	if app.Stage != "" {
		st, err := zonebus.ParseStage(app.Stage)
		if err != nil {
			return zonebus.NewZone{}, errs.Newf(errs.InvalidArgument, "invalid stage: %s", err)
		}
		dest.Stage = &st
	}

	return dest, nil
}

type UpdateZone struct {
	WarehouseID *string `json:"warehouse_id" validate:"omitempty,min=36,min=36"`
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	Stage       *string `json:"stage" validate:"omitempty"`
}

func (app *UpdateZone) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateZone) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateZone(app UpdateZone) (zonebus.UpdateZone, error) {
	dest := zonebus.UpdateZone{}

	if app.WarehouseID != nil {
		warehouseID, err := uuid.Parse(*app.WarehouseID)
		if err != nil {
			return zonebus.UpdateZone{}, err
		}

		dest.WarehouseID = &warehouseID
	}

	if app.Description != nil {
		dest.Description = app.Description
	}

	if app.Name != nil {
		dest.Name = app.Name
	}

	if app.Stage != nil && *app.Stage != "" {
		st, err := zonebus.ParseStage(*app.Stage)
		if err != nil {
			return zonebus.UpdateZone{}, errs.Newf(errs.InvalidArgument, "invalid stage: %s", err)
		}
		dest.Stage = &st
	}

	return dest, nil
}

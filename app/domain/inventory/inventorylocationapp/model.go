package inventorylocationapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"

	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	LocationID         string
	WarehouseID        string
	ZoneID             string
	Aisle              string
	Rack               string
	Shelf              string
	Bin                string
	LocationCode       string
	IsPickLocation     string
	IsReserveLocation  string
	MaxCapacity        string
	CurrentUtilization string
	CreatedDate        string
	UpdatedDate        string
}

type InventoryLocation struct {
	LocationID         string `json:"location_id"`
	WarehouseID        string `json:"warehouse_id"`
	ZoneID             string `json:"zone_id"`
	Aisle              string `json:"aisle"`
	Rack               string `json:"rack"`
	Shelf              string `json:"shelf"`
	Bin                string `json:"bin"`
	LocationCode       string `json:"location_code"`
	IsPickLocation     string `json:"is_pick_location"`
	IsReserveLocation  string `json:"is_reserve_location"`
	MaxCapacity        string `json:"max_capacity"`
	CurrentUtilization string `json:"current_utilization"`
	CreatedDate        string `json:"created_date"`
	UpdatedDate        string `json:"updated_date"`
}

func (app InventoryLocation) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppInventoryLocation(bus inventorylocationbus.InventoryLocation) InventoryLocation {
	locationCode := ""
	if bus.LocationCode != nil {
		locationCode = *bus.LocationCode
	}

	return InventoryLocation{
		LocationID:         bus.LocationID.String(),
		WarehouseID:        bus.WarehouseID.String(),
		ZoneID:             bus.ZoneID.String(),
		Aisle:              bus.Aisle,
		Rack:               bus.Rack,
		Shelf:              bus.Shelf,
		Bin:                bus.Bin,
		LocationCode:       locationCode,
		IsPickLocation:     fmt.Sprintf("%t", bus.IsPickLocation),
		IsReserveLocation:  fmt.Sprintf("%t", bus.IsReserveLocation),
		MaxCapacity:        fmt.Sprintf("%d", bus.MaxCapacity),
		CurrentUtilization: bus.CurrentUtilization.String(),
		CreatedDate:        bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:        bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppInventoryLocations(bus []inventorylocationbus.InventoryLocation) []InventoryLocation {
	app := make([]InventoryLocation, len(bus))
	for i, v := range bus {
		app[i] = ToAppInventoryLocation(v)
	}
	return app
}

type NewInventoryLocation struct {
	WarehouseID        string `json:"warehouse_id" validate:"required,min=36,max=36"`
	ZoneID             string `json:"zone_id" validate:"required,min=36,max=36"`
	Aisle              string `json:"aisle" validate:"required"`
	Rack               string `json:"rack" validate:"required"`
	Shelf              string `json:"shelf" validate:"required"`
	Bin                string `json:"bin" validate:"required"`
	LocationCode       string `json:"location_code" validate:"omitempty,max=100"`
	IsPickLocation     string `json:"is_pick_location" validate:"required"`
	IsReserveLocation  string `json:"is_reserve_location" validate:"required"`
	MaxCapacity        string `json:"max_capacity" validate:"required"`
	CurrentUtilization string `json:"current_utilization" validate:"required"`
}

func (app *NewInventoryLocation) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewInventoryLocation) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewInventoryLocation(app NewInventoryLocation) (inventorylocationbus.NewInventoryLocation, error) {

	warehouseID, err := uuid.Parse(app.WarehouseID)
	if err != nil {
		return inventorylocationbus.NewInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse warehouseID: %s", err)
	}

	zoneID, err := uuid.Parse(app.ZoneID)
	if err != nil {
		return inventorylocationbus.NewInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse zoneID: %s", err)
	}

	maxCapacity, err := strconv.Atoi(app.MaxCapacity)
	if err != nil {
		return inventorylocationbus.NewInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse maxCapacity: %s", err)
	}

	cu, err := types.ParseRoundedFloat(app.CurrentUtilization)
	if err != nil {
		return inventorylocationbus.NewInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse currentUtilization: %s", err)
	}

	isPL, err := strconv.ParseBool(app.IsPickLocation)
	if err != nil {
		return inventorylocationbus.NewInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse isPickLocation: %s", err)
	}

	isRL, err := strconv.ParseBool(app.IsReserveLocation)
	if err != nil {
		return inventorylocationbus.NewInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse isReserveLocation: %s", err)
	}

	var locationCode *string
	if app.LocationCode != "" {
		locationCode = &app.LocationCode
	}

	return inventorylocationbus.NewInventoryLocation{
		WarehouseID:        warehouseID,
		ZoneID:             zoneID,
		Aisle:              app.Aisle,
		Rack:               app.Rack,
		Shelf:              app.Shelf,
		Bin:                app.Bin,
		LocationCode:       locationCode,
		IsPickLocation:     isPL,
		IsReserveLocation:  isRL,
		MaxCapacity:        maxCapacity,
		CurrentUtilization: cu,
	}, nil
}

type UpdateInventoryLocation struct {
	WarehouseID        *string `json:"warehouse_id" validate:"omitempty,min=36,max=36"`
	ZoneID             *string `json:"zone_id" validate:"omitempty,min=36,max=36"`
	Aisle              *string `json:"aisle" validate:"omitempty"`
	Rack               *string `json:"rack" validate:"omitempty"`
	Shelf              *string `json:"shelf" validate:"omitempty"`
	Bin                *string `json:"bin" validate:"omitempty"`
	LocationCode       *string `json:"location_code" validate:"omitempty,max=100"`
	IsPickLocation     *string `json:"is_pick_location" validate:"omitempty"`
	IsReserveLocation  *string `json:"is_reserve_location" validate:"omitempty"`
	MaxCapacity        *string `json:"max_capacity" validate:"omitempty"`
	CurrentUtilization *string `json:"current_utilization" validate:"omitempty"`
}

func (app *UpdateInventoryLocation) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateInventoryLocation) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateInventoryLocation(app UpdateInventoryLocation) (inventorylocationbus.UpdateInventoryLocation, error) {
	bus := inventorylocationbus.UpdateInventoryLocation{
		Aisle:        app.Aisle,
		Rack:         app.Rack,
		Shelf:        app.Shelf,
		Bin:          app.Bin,
		LocationCode: app.LocationCode,
	}

	if app.WarehouseID != nil {
		warehouseID, err := uuid.Parse(*app.WarehouseID)
		if err != nil {
			return inventorylocationbus.UpdateInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse warehouseID: %s", err)
		}
		bus.WarehouseID = &warehouseID
	}

	if app.ZoneID != nil {
		zoneID, err := uuid.Parse(*app.ZoneID)
		if err != nil {
			return inventorylocationbus.UpdateInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse zoneID: %s", err)
		}
		bus.ZoneID = &zoneID
	}

	if app.MaxCapacity != nil {
		maxCapacity, err := strconv.Atoi(*app.MaxCapacity)
		if err != nil {
			return inventorylocationbus.UpdateInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse maxCapacity: %s", err)
		}
		bus.MaxCapacity = &maxCapacity
	}

	if app.CurrentUtilization != nil {
		cu, err := types.ParseRoundedFloat(*app.CurrentUtilization)
		if err != nil {
			return inventorylocationbus.UpdateInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse currentUtilization: %s", err)
		}
		bus.CurrentUtilization = &cu
	}

	if app.IsPickLocation != nil {
		isPL, err := strconv.ParseBool(*app.IsPickLocation)
		if err != nil {
			return inventorylocationbus.UpdateInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse isPickLocation: %s", err)
		}
		bus.IsPickLocation = &isPL
	}

	if app.IsReserveLocation != nil {
		isRL, err := strconv.ParseBool(*app.IsReserveLocation)
		if err != nil {
			return inventorylocationbus.UpdateInventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse isReserveLocation: %s", err)
		}
		bus.IsReserveLocation = &isRL
	}

	return bus, nil
}

package inventorylocationdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus/types"
)

type inventoryLocation struct {
	LocationID         uuid.UUID      `db:"location_id"`
	WarehouseID        uuid.UUID      `db:"warehouse_id"`
	ZoneID             uuid.UUID      `db:"zone_id"`
	Aisle              string         `db:"aisle"`
	Rack               string         `db:"rack"`
	Shelf              string         `db:"shelf"`
	Bin                string         `db:"bin"`
	IsPickLocation     bool           `db:"is_pick_location"`
	IsReserveLocation  bool           `db:"is_reserve_location"`
	MaxCapacity        int            `db:"max_capacity"`
	CurrentUtilization sql.NullString `db:"current_utilization"`
	CreatedDate        time.Time      `db:"created_date"`
	UpdatedDate        time.Time      `db:"updated_date"`
}

func toDBInvLocation(bus inventorylocationbus.InventoryLocation) inventoryLocation {
	return inventoryLocation{
		LocationID:         bus.LocationID,
		WarehouseID:        bus.WarehouseID,
		ZoneID:             bus.ZoneID,
		Aisle:              bus.Aisle,
		Rack:               bus.Rack,
		Shelf:              bus.Shelf,
		Bin:                bus.Bin,
		IsPickLocation:     bus.IsPickLocation,
		IsReserveLocation:  bus.IsReserveLocation,
		MaxCapacity:        bus.MaxCapacity,
		CurrentUtilization: bus.CurrentUtilization.DBValue(),
		CreatedDate:        bus.CreatedDate,
		UpdatedDate:        bus.UpdatedDate,
	}
}

func toBusInvLocation(db inventoryLocation) (inventorylocationbus.InventoryLocation, error) {
	utilization, err := types.ParseRoundedFloat(db.CurrentUtilization.String)
	if err != nil {
		return inventorylocationbus.InventoryLocation{}, fmt.Errorf("toBusInvLocation: failed to parse float from db %w", err)
	}

	return inventorylocationbus.InventoryLocation{
		LocationID:         db.LocationID,
		WarehouseID:        db.WarehouseID,
		ZoneID:             db.ZoneID,
		Aisle:              db.Aisle,
		Rack:               db.Rack,
		Shelf:              db.Shelf,
		Bin:                db.Bin,
		IsPickLocation:     db.IsPickLocation,
		IsReserveLocation:  db.IsReserveLocation,
		MaxCapacity:        db.MaxCapacity,
		CurrentUtilization: utilization,
		CreatedDate:        db.CreatedDate,
		UpdatedDate:        db.UpdatedDate,
	}, nil
}

func toBusInvLocations(db []inventoryLocation) ([]inventorylocationbus.InventoryLocation, error) {
	busInvLocations := make([]inventorylocationbus.InventoryLocation, len(db))

	for i, dbInvLocation := range db {
		busInvLocation, err := toBusInvLocation(dbInvLocation)
		if err != nil {
			return nil, fmt.Errorf("toBusInvLocations: %w", err)
		}
		busInvLocations[i] = busInvLocation
	}

	return busInvLocations, nil
}

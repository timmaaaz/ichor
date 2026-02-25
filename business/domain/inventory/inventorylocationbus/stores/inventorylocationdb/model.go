package inventorylocationdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
)

type inventoryLocation struct {
	LocationID         uuid.UUID      `db:"id"`
	WarehouseID        uuid.UUID      `db:"warehouse_id"`
	ZoneID             uuid.UUID      `db:"zone_id"`
	Aisle              string         `db:"aisle"`
	Rack               string         `db:"rack"`
	Shelf              string         `db:"shelf"`
	Bin                string         `db:"bin"`
	LocationCode       sql.NullString `db:"location_code"`
	IsPickLocation     bool           `db:"is_pick_location"`
	IsReserveLocation  bool           `db:"is_reserve_location"`
	MaxCapacity        int            `db:"max_capacity"`
	CurrentUtilization sql.NullString `db:"current_utilization"`
	CreatedDate        time.Time      `db:"created_date"`
	UpdatedDate        time.Time      `db:"updated_date"`
}

func toDBInvLocation(bus inventorylocationbus.InventoryLocation) inventoryLocation {
	var locationCode sql.NullString
	if bus.LocationCode != nil {
		locationCode = sql.NullString{String: *bus.LocationCode, Valid: true}
	}

	return inventoryLocation{
		LocationID:         bus.LocationID,
		WarehouseID:        bus.WarehouseID,
		ZoneID:             bus.ZoneID,
		Aisle:              bus.Aisle,
		Rack:               bus.Rack,
		Shelf:              bus.Shelf,
		Bin:                bus.Bin,
		LocationCode:       locationCode,
		IsPickLocation:     bus.IsPickLocation,
		IsReserveLocation:  bus.IsReserveLocation,
		MaxCapacity:        bus.MaxCapacity,
		CurrentUtilization: bus.CurrentUtilization.DBValue(),
		CreatedDate:        bus.CreatedDate,
		UpdatedDate:        bus.UpdatedDate,
	}
}

func toBusInvLocation(db inventoryLocation) (inventorylocationbus.InventoryLocation, error) {
	var utilization types.RoundedFloat
	if db.CurrentUtilization.Valid && db.CurrentUtilization.String != "" {
		var err error
		utilization, err = types.ParseRoundedFloat(db.CurrentUtilization.String)
		if err != nil {
			return inventorylocationbus.InventoryLocation{}, fmt.Errorf("toBusInvLocation: failed to parse float from db %w", err)
		}
	}

	var locationCode *string
	if db.LocationCode.Valid {
		locationCode = &db.LocationCode.String
	}

	return inventorylocationbus.InventoryLocation{
		LocationID:         db.LocationID,
		WarehouseID:        db.WarehouseID,
		ZoneID:             db.ZoneID,
		Aisle:              db.Aisle,
		Rack:               db.Rack,
		Shelf:              db.Shelf,
		Bin:                db.Bin,
		LocationCode:       locationCode,
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

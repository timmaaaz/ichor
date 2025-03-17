package zonedb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
)

type zone struct {
	ID          uuid.UUID `db:"zone_id"`
	WarehouseID uuid.UUID `db:"warehouse_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	IsActive    bool      `db:"is_active"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
	CreatedBy   uuid.UUID `db:"created_by"`
	UpdatedBy   uuid.UUID `db:"updated_by"`
}

func toDBZone(bus zonebus.Zone) zone {
	return zone{
		ID:          bus.ID,
		WarehouseID: bus.WarehouseID,
		Name:        bus.Name,
		Description: bus.Description,
		IsActive:    bus.IsActive,
		DateCreated: bus.DateCreated,
		DateUpdated: bus.DateUpdated,
		CreatedBy:   bus.CreatedBy,
		UpdatedBy:   bus.UpdatedBy,
	}
}

func toBusZone(db zone) zonebus.Zone {
	return zonebus.Zone{
		ID:          db.ID,
		WarehouseID: db.WarehouseID,
		Name:        db.Name,
		Description: db.Description,
		IsActive:    db.IsActive,
		DateCreated: db.DateCreated,
		DateUpdated: db.DateUpdated,
		CreatedBy:   db.CreatedBy,
		UpdatedBy:   db.UpdatedBy,
	}
}

func toBusZones(db []zone) []zonebus.Zone {
	buses := make([]zonebus.Zone, len(db))
	for i, d := range db {
		buses[i] = toBusZone(d)
	}
	return buses
}

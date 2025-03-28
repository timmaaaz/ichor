package zonedb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
)

type zone struct {
	ZoneID      uuid.UUID `db:"zone_id"`
	WarehouseID uuid.UUID `db:"warehouse_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedDate time.Time `db:"created_date"`
	UpdatedDate time.Time `db:"updated_date"`
}

func toDBZone(bus zonebus.Zone) zone {
	return zone{
		ZoneID:      bus.ZoneID,
		WarehouseID: bus.WarehouseID,
		Name:        bus.Name,
		Description: bus.Description,
		CreatedDate: bus.CreatedDate.UTC(),
		UpdatedDate: bus.UpdatedDate.UTC(),
	}
}

func toBusZone(db zone) zonebus.Zone {
	return zonebus.Zone{
		ZoneID:      db.ZoneID,
		WarehouseID: db.WarehouseID,
		Name:        db.Name,
		Description: db.Description,
		CreatedDate: db.CreatedDate.Local(),
		UpdatedDate: db.UpdatedDate.Local(),
	}
}

func toBusZones(dbs []zone) []zonebus.Zone {
	bus := make([]zonebus.Zone, len(dbs))

	for i, db := range dbs {
		bus[i] = toBusZone(db)
	}

	return bus
}

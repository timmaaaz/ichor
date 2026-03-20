package zonedb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
)

type zone struct {
	ZoneID      uuid.UUID      `db:"id"`
	WarehouseID uuid.UUID      `db:"warehouse_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Stage       sql.NullString `db:"stage"`
	CreatedDate time.Time      `db:"created_date"`
	UpdatedDate time.Time      `db:"updated_date"`
}

func toDBZone(bus zonebus.Zone) zone {
	dest := zone{
		ZoneID:      bus.ZoneID,
		WarehouseID: bus.WarehouseID,
		Name:        bus.Name,
		CreatedDate: bus.CreatedDate.UTC(),
		UpdatedDate: bus.UpdatedDate.UTC(),
	}
	if bus.Description != "" {
		dest.Description = sql.NullString{String: bus.Description, Valid: true}
	}
	if bus.Stage != nil {
		dest.Stage = sql.NullString{String: bus.Stage.String(), Valid: true}
	}
	return dest
}

func toBusZone(db zone) zonebus.Zone {
	dest := zonebus.Zone{
		ZoneID:      db.ZoneID,
		WarehouseID: db.WarehouseID,
		Name:        db.Name,
		CreatedDate: db.CreatedDate.Local(),
		UpdatedDate: db.UpdatedDate.Local(),
	}
	if db.Description.Valid {
		dest.Description = db.Description.String
	}
	if db.Stage.Valid {
		st, err := zonebus.ParseStage(db.Stage.String)
		if err == nil {
			dest.Stage = &st
		}
	}
	return dest
}

func toBusZones(dbs []zone) []zonebus.Zone {
	bus := make([]zonebus.Zone, len(dbs))

	for i, db := range dbs {
		bus[i] = toBusZone(db)
	}

	return bus
}

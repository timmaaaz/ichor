package warehousedb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
)

type warehouse struct {
	ID          uuid.UUID `db:"id"`
	StreetID    uuid.UUID `db:"street_id"`
	Name        string    `db:"name"`
	IsActive    bool      `db:"is_active"`
	CreatedDate time.Time `db:"created_date"`
	UpdatedDate time.Time `db:"updated_date"`
	CreatedBy   uuid.UUID `db:"created_by"`
	UpdatedBy   uuid.UUID `db:"updated_by"`
}

func toDBWarehouse(bus warehousebus.Warehouse) warehouse {
	return warehouse{
		ID:          bus.ID,
		StreetID:    bus.StreetID,
		Name:        bus.Name,
		IsActive:    bus.IsActive,
		CreatedDate: bus.CreatedDate,
		UpdatedDate: bus.UpdatedDate,
		CreatedBy:   bus.CreatedBy,
		UpdatedBy:   bus.UpdatedBy,
	}
}

func toBusWarehouse(db warehouse) warehousebus.Warehouse {
	return warehousebus.Warehouse{
		ID:          db.ID,
		StreetID:    db.StreetID,
		Name:        db.Name,
		IsActive:    db.IsActive,
		CreatedDate: db.CreatedDate,
		UpdatedDate: db.UpdatedDate,
		CreatedBy:   db.CreatedBy,
		UpdatedBy:   db.UpdatedBy,
	}
}

func toBusWarehouses(db []warehouse) []warehousebus.Warehouse {
	buses := make([]warehousebus.Warehouse, len(db))
	for i, d := range db {
		buses[i] = toBusWarehouse(d)
	}
	return buses
}

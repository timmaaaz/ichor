package lotlocationdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
)

type lotLocation struct {
	ID          uuid.UUID `db:"id"`
	LotID       uuid.UUID `db:"lot_id"`
	LocationID  uuid.UUID `db:"location_id"`
	Quantity    int       `db:"quantity"`
	CreatedDate time.Time `db:"created_date"`
	UpdatedDate time.Time `db:"updated_date"`
}

func toDBLotLocation(bus lotlocationbus.LotLocation) lotLocation {
	return lotLocation{
		ID:          bus.ID,
		LotID:       bus.LotID,
		LocationID:  bus.LocationID,
		Quantity:    bus.Quantity,
		CreatedDate: bus.CreatedDate.UTC(),
		UpdatedDate: bus.UpdatedDate.UTC(),
	}
}

func toBusLotLocation(db lotLocation) lotlocationbus.LotLocation {
	return lotlocationbus.LotLocation{
		ID:          db.ID,
		LotID:       db.LotID,
		LocationID:  db.LocationID,
		Quantity:    db.Quantity,
		CreatedDate: db.CreatedDate.Local(),
		UpdatedDate: db.UpdatedDate.Local(),
	}
}

func toBusLotLocations(dbs []lotLocation) []lotlocationbus.LotLocation {
	bus := make([]lotlocationbus.LotLocation, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusLotLocation(db)
	}
	return bus
}

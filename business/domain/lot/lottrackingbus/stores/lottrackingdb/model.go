package lottrackingdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
)

type lotTracking struct {
	LotID             uuid.UUID `db:"lot_id"`
	SupplierProductID uuid.UUID `db:"supplier_product_id"`
	LotNumber         string    `db:"lot_number"`
	ManufactureDate   time.Time `db:"manufacture_date"`
	ExpirationDate    time.Time `db:"expiration_date"`
	RecievedDate      time.Time `db:"received_date"`
	Quantity          int       `db:"quantity"`
	QualityStatus     string    `db:"quality_status"`
	CreatedDate       time.Time `db:"created_date"`
	UpdatedDate       time.Time `db:"updated_date"`
}

func toDBLotTracking(bus lottrackingbus.LotTracking) lotTracking {
	return lotTracking{
		LotID:             bus.LotID,
		SupplierProductID: bus.SupplierProductID,
		LotNumber:         bus.LotNumber,
		ManufactureDate:   bus.ManufactureDate.UTC(),
		ExpirationDate:    bus.ExpirationDate.UTC(),
		RecievedDate:      bus.RecievedDate.UTC(),
		Quantity:          bus.Quantity,
		QualityStatus:     bus.QualityStatus,
		CreatedDate:       bus.CreatedDate.UTC(),
		UpdatedDate:       bus.UpdatedDate.UTC(),
	}
}

func toBusLotTracking(db lotTracking) lottrackingbus.LotTracking {
	return lottrackingbus.LotTracking{
		LotID:             db.LotID,
		SupplierProductID: db.SupplierProductID,
		LotNumber:         db.LotNumber,
		ManufactureDate:   db.ManufactureDate.Local(),
		ExpirationDate:    db.ExpirationDate.Local(),
		RecievedDate:      db.RecievedDate.Local(),
		Quantity:          db.Quantity,
		QualityStatus:     db.QualityStatus,
		CreatedDate:       db.CreatedDate.Local(),
		UpdatedDate:       db.UpdatedDate.Local(),
	}
}

func toBusLotTrackings(db []lotTracking) []lottrackingbus.LotTracking {
	bus := make([]lottrackingbus.LotTracking, len(db))

	for i, db := range db {
		bus[i] = toBusLotTracking(db)
	}

	return bus
}

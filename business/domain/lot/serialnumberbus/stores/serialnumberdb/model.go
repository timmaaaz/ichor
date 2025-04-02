package serialnumberdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/lot/serialnumberbus"
)

type serialNumber struct {
	SerialID     uuid.UUID `db:"serial_id"`
	LotID        uuid.UUID `db:"lot_id"`
	ProductID    uuid.UUID `db:"product_id"`
	LocationID   uuid.UUID `db:"location_id"`
	SerialNumber string    `db:"serial_number"`
	Status       string    `db:"status"`
	CreatedDate  time.Time `db:"created_date"`
	UpdatedDate  time.Time `db:"updated_date"`
}

func toBusSerialNumber(db serialNumber) serialnumberbus.SerialNumber {
	return serialnumberbus.SerialNumber{
		SerialID:     db.SerialID,
		LotID:        db.LotID,
		ProductID:    db.ProductID,
		LocationID:   db.LocationID,
		SerialNumber: db.SerialNumber,
		Status:       db.Status,
		CreatedDate:  db.CreatedDate,
		UpdatedDate:  db.UpdatedDate,
	}
}

func toBusSerialNumbers(dbs []serialNumber) []serialnumberbus.SerialNumber {
	bus := make([]serialnumberbus.SerialNumber, len(dbs))

	for i, db := range dbs {
		bus[i] = toBusSerialNumber(db)
	}
	return bus
}

func toDBSerialNumber(bus serialnumberbus.SerialNumber) serialNumber {
	return serialNumber{
		SerialID:     bus.SerialID,
		LotID:        bus.LotID,
		LocationID:   bus.LocationID,
		ProductID:    bus.ProductID,
		SerialNumber: bus.SerialNumber,
		Status:       bus.Status,
		CreatedDate:  bus.CreatedDate,
		UpdatedDate:  bus.UpdatedDate,
	}
}

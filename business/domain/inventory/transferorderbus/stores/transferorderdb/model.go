package transferorderdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
)

type transferOrder struct {
	TransferID     uuid.UUID    `db:"id"`
	ProductID      uuid.UUID    `db:"product_id"`
	FromLocationID uuid.UUID    `db:"from_location_id"`
	ToLocationID   uuid.UUID    `db:"to_location_id"`
	RequestedByID  uuid.UUID    `db:"requested_by"`
	ApprovedByID   uuid.NullUUID `db:"approved_by"`
	Quantity       int          `db:"quantity"`
	Status         string       `db:"status"`
	TransferDate   time.Time    `db:"transfer_date"`
	CreatedDate    time.Time    `db:"created_date"`
	UpdatedDate    time.Time    `db:"updated_date"`
}

func toBusTransferOrder(db transferOrder) transferorderbus.TransferOrder {
	to := transferorderbus.TransferOrder{
		TransferID:     db.TransferID,
		ProductID:      db.ProductID,
		FromLocationID: db.FromLocationID,
		ToLocationID:   db.ToLocationID,
		RequestedByID:  db.RequestedByID,
		Quantity:       db.Quantity,
		Status:         db.Status,
		TransferDate:   db.TransferDate,
		CreatedDate:    db.CreatedDate,
		UpdatedDate:    db.UpdatedDate,
	}

	if db.ApprovedByID.Valid {
		to.ApprovedByID = &db.ApprovedByID.UUID
	}

	return to
}

func toBusTransferOrders(dbs []transferOrder) []transferorderbus.TransferOrder {
	app := make([]transferorderbus.TransferOrder, len(dbs))
	for i, db := range dbs {
		app[i] = toBusTransferOrder(db)
	}
	return app
}

func toDBTransferOrder(bus transferorderbus.TransferOrder) transferOrder {
	db := transferOrder{
		TransferID:     bus.TransferID,
		ProductID:      bus.ProductID,
		FromLocationID: bus.FromLocationID,
		ToLocationID:   bus.ToLocationID,
		RequestedByID:  bus.RequestedByID,
		Quantity:       bus.Quantity,
		Status:         bus.Status,
		TransferDate:   bus.TransferDate,
		CreatedDate:    bus.CreatedDate,
		UpdatedDate:    bus.UpdatedDate,
	}

	if bus.ApprovedByID != nil {
		db.ApprovedByID = uuid.NullUUID{UUID: *bus.ApprovedByID, Valid: true}
	}

	return db
}

package transferorderdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/movement/transferorderbus"
)

type transferOrder struct {
	TransferID     uuid.UUID `db:"transfer_id"`
	ProductID      uuid.UUID `db:"product_id"`
	FromLocationID uuid.UUID `db:"from_location_id"`
	ToLocationID   uuid.UUID `db:"to_location_id"`
	RequestedByID  uuid.UUID `db:"requested_by"`
	ApprovedByID   uuid.UUID `db:"approved_by"`
	Quantity       int       `db:"quantity"`
	Status         string    `db:"status"`
	TransferDate   time.Time `db:"transfer_date"`
	CreatedDate    time.Time `db:"created_date"`
	UpdatedDate    time.Time `db:"updated_date"`
}

func toBusTransferOrder(db transferOrder) transferorderbus.TransferOrder {
	return transferorderbus.TransferOrder{
		TransferID:     db.TransferID,
		ProductID:      db.ProductID,
		FromLocationID: db.FromLocationID,
		ToLocationID:   db.ToLocationID,
		RequestedByID:  db.RequestedByID,
		ApprovedByID:   db.ApprovedByID,
		Quantity:       db.Quantity,
		Status:         db.Status,
		TransferDate:   db.TransferDate,
		CreatedDate:    db.CreatedDate,
		UpdatedDate:    db.UpdatedDate,
	}
}

func toBusTransferOrders(dbs []transferOrder) []transferorderbus.TransferOrder {
	app := make([]transferorderbus.TransferOrder, len(dbs))
	for i, db := range dbs {
		app[i] = toBusTransferOrder(db)
	}
	return app
}

func toDBTransferOrder(app transferorderbus.TransferOrder) transferOrder {
	return transferOrder{
		TransferID:     app.TransferID,
		ProductID:      app.ProductID,
		FromLocationID: app.FromLocationID,
		ToLocationID:   app.ToLocationID,
		RequestedByID:  app.RequestedByID,
		ApprovedByID:   app.ApprovedByID,
		Quantity:       app.Quantity,
		Status:         app.Status,
		TransferDate:   app.TransferDate,
		CreatedDate:    app.CreatedDate,
		UpdatedDate:    app.UpdatedDate,
	}
}

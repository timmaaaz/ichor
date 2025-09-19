package ordersdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
)

type dbOrder struct {
	ID                  uuid.UUID `db:"id"`
	Number              string    `db:"number"`
	CustomerID          uuid.UUID `db:"customer_id"`
	DueDate             time.Time `db:"due_date"`
	FulfillmentStatusID uuid.UUID `db:"order_fulfillment_status_id"`
	CreatedBy           uuid.UUID `db:"created_by"`
	UpdatedBy           uuid.UUID `db:"updated_by"`
	CreatedDate         time.Time `db:"created_date"`
	UpdatedDate         time.Time `db:"updated_date"`
}

func toBusOrder(db dbOrder) ordersbus.Order {
	return ordersbus.Order{
		ID:                  db.ID,
		Number:              db.Number,
		CustomerID:          db.CustomerID,
		DueDate:             db.DueDate,
		FulfillmentStatusID: db.FulfillmentStatusID,
		CreatedBy:           db.CreatedBy,
		UpdatedBy:           db.UpdatedBy,
		CreatedDate:         db.CreatedDate,
		UpdatedDate:         db.UpdatedDate,
	}
}

func toBusOrders(dbs []dbOrder) []ordersbus.Order {
	app := make([]ordersbus.Order, len(dbs))
	for i, db := range dbs {
		app[i] = toBusOrder(db)
	}
	return app
}

func toDBOrder(app ordersbus.Order) dbOrder {
	return dbOrder{
		ID:                  app.ID,
		Number:              app.Number,
		CustomerID:          app.CustomerID,
		DueDate:             app.DueDate,
		FulfillmentStatusID: app.FulfillmentStatusID,
		CreatedBy:           app.CreatedBy,
		UpdatedBy:           app.UpdatedBy,
		CreatedDate:         app.CreatedDate,
		UpdatedDate:         app.UpdatedDate,
	}
}

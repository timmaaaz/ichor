package orderlineitemsdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
)

type orderLineItem struct {
	ID                            uuid.UUID `db:"id"`
	OrderID                       uuid.UUID `db:"order_id"`
	ProductID                     uuid.UUID `db:"product_id"`
	Quantity                      int       `db:"quantity"`
	Discount                      float64   `db:"discount"`
	LineItemFulfillmentStatusesID uuid.UUID `db:"line_item_fulfillment_statuses_id"`
	CreatedBy                     uuid.UUID `db:"created_by"`
	CreatedDate                   time.Time `db:"created_date"`
	UpdatedBy                     uuid.UUID `db:"updated_by"`
	UpdatedDate                   time.Time `db:"updated_date"`
}

func toBusOrderLineItem(db orderLineItem) orderlineitemsbus.OrderLineItem {
	return orderlineitemsbus.OrderLineItem{
		ID:                            db.ID,
		OrderID:                       db.OrderID,
		ProductID:                     db.ProductID,
		Quantity:                      db.Quantity,
		Discount:                      db.Discount,
		LineItemFulfillmentStatusesID: db.LineItemFulfillmentStatusesID,
		CreatedBy:                     db.CreatedBy,
		CreatedDate:                   db.CreatedDate,
		UpdatedBy:                     db.UpdatedBy,
		UpdatedDate:                   db.UpdatedDate,
	}
}

func toBusOrderLineItems(dbs []orderLineItem) []orderlineitemsbus.OrderLineItem {
	app := make([]orderlineitemsbus.OrderLineItem, len(dbs))
	for i, db := range dbs {
		app[i] = toBusOrderLineItem(db)
	}
	return app
}

func toDBOrderLineItem(app orderlineitemsbus.OrderLineItem) orderLineItem {
	return orderLineItem{
		ID:                            app.ID,
		OrderID:                       app.OrderID,
		ProductID:                     app.ProductID,
		Quantity:                      app.Quantity,
		Discount:                      app.Discount,
		LineItemFulfillmentStatusesID: app.LineItemFulfillmentStatusesID,
		CreatedBy:                     app.CreatedBy,
		CreatedDate:                   app.CreatedDate,
		UpdatedBy:                     app.UpdatedBy,
		UpdatedDate:                   app.UpdatedDate,
	}
}

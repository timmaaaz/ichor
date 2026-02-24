package orderlineitemsdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
)

type orderLineItem struct {
	ID                            uuid.UUID      `db:"id"`
	OrderID                       uuid.UUID      `db:"order_id"`
	ProductID                     uuid.UUID      `db:"product_id"`
	Description                   string         `db:"description"`
	Quantity                      int            `db:"quantity"`
	UnitPrice                     sql.NullString `db:"unit_price"`
	Discount                      sql.NullString `db:"discount"`
	DiscountType                  string         `db:"discount_type"`
	LineTotal                     sql.NullString `db:"line_total"`
	LineItemFulfillmentStatusesID uuid.UUID      `db:"line_item_fulfillment_statuses_id"`
	PickedQuantity                int            `db:"picked_quantity"`
	BackorderedQuantity           int            `db:"backordered_quantity"`
	ShortPickReason               sql.NullString `db:"short_pick_reason"`
	CreatedBy                     uuid.UUID      `db:"created_by"`
	CreatedDate                   time.Time      `db:"created_date"`
	UpdatedBy                     uuid.UUID      `db:"updated_by"`
	UpdatedDate                   time.Time      `db:"updated_date"`
}

func toBusOrderLineItem(db orderLineItem) (orderlineitemsbus.OrderLineItem, error) {
	unitPrice, err := types.ParseMoney(db.UnitPrice.String)
	if err != nil {
		return orderlineitemsbus.OrderLineItem{}, fmt.Errorf("tobusorderlineitem: unit_price: %w", err)
	}

	discount, err := types.ParseMoney(db.Discount.String)
	if err != nil {
		return orderlineitemsbus.OrderLineItem{}, fmt.Errorf("tobusorderlineitem: discount: %w", err)
	}

	lineTotal, err := types.ParseMoney(db.LineTotal.String)
	if err != nil {
		return orderlineitemsbus.OrderLineItem{}, fmt.Errorf("tobusorderlineitem: line_total: %w", err)
	}

	var shortPickReason *string
	if db.ShortPickReason.Valid {
		shortPickReason = &db.ShortPickReason.String
	}

	return orderlineitemsbus.OrderLineItem{
		ID:                            db.ID,
		OrderID:                       db.OrderID,
		ProductID:                     db.ProductID,
		Description:                   db.Description,
		Quantity:                      db.Quantity,
		UnitPrice:                     unitPrice,
		Discount:                      discount,
		DiscountType:                  db.DiscountType,
		LineTotal:                     lineTotal,
		LineItemFulfillmentStatusesID: db.LineItemFulfillmentStatusesID,
		PickedQuantity:                db.PickedQuantity,
		BackorderedQuantity:           db.BackorderedQuantity,
		ShortPickReason:               shortPickReason,
		CreatedBy:                     db.CreatedBy,
		CreatedDate:                   db.CreatedDate.In(time.Local),
		UpdatedBy:                     db.UpdatedBy,
		UpdatedDate:                   db.UpdatedDate.In(time.Local),
	}, nil
}

func toBusOrderLineItems(dbs []orderLineItem) ([]orderlineitemsbus.OrderLineItem, error) {
	items := make([]orderlineitemsbus.OrderLineItem, len(dbs))
	for i, db := range dbs {
		item, err := toBusOrderLineItem(db)
		if err != nil {
			return nil, fmt.Errorf("tobusorderlineitems[%d]: %w", i, err)
		}
		items[i] = item
	}
	return items, nil
}

func toDBOrderLineItem(app orderlineitemsbus.OrderLineItem) orderLineItem {
	var shortPickReason sql.NullString
	if app.ShortPickReason != nil {
		shortPickReason = sql.NullString{String: *app.ShortPickReason, Valid: true}
	}

	return orderLineItem{
		ID:                            app.ID,
		OrderID:                       app.OrderID,
		ProductID:                     app.ProductID,
		Description:                   app.Description,
		Quantity:                      app.Quantity,
		UnitPrice:                     app.UnitPrice.DBValue(),
		Discount:                      app.Discount.DBValue(),
		DiscountType:                  app.DiscountType,
		LineTotal:                     app.LineTotal.DBValue(),
		LineItemFulfillmentStatusesID: app.LineItemFulfillmentStatusesID,
		PickedQuantity:                app.PickedQuantity,
		BackorderedQuantity:           app.BackorderedQuantity,
		ShortPickReason:               shortPickReason,
		CreatedBy:                     app.CreatedBy,
		CreatedDate:                   app.CreatedDate.UTC(),
		UpdatedBy:                     app.UpdatedBy,
		UpdatedDate:                   app.UpdatedDate.UTC(),
	}
}

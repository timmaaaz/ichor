package purchaseorderlineitemdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
)

type purchaseOrderLineItem struct {
	ID                   uuid.UUID      `db:"id"`
	PurchaseOrderID      uuid.UUID      `db:"purchase_order_id"`
	SupplierProductID    uuid.UUID      `db:"supplier_product_id"`
	QuantityOrdered      int            `db:"quantity_ordered"`
	QuantityReceived     int            `db:"quantity_received"`
	QuantityCancelled    int            `db:"quantity_cancelled"`
	UnitCost             float64        `db:"unit_cost"`
	Discount             float64        `db:"discount"`
	LineTotal            float64        `db:"line_total"`
	LineItemStatusID     uuid.UUID      `db:"line_item_status_id"`
	ExpectedDeliveryDate sql.NullTime   `db:"expected_delivery_date"`
	ActualDeliveryDate   sql.NullTime   `db:"actual_delivery_date"`
	Notes                sql.NullString `db:"notes"`
	CreatedBy            uuid.UUID      `db:"created_by"`
	UpdatedBy            uuid.UUID      `db:"updated_by"`
	CreatedDate          time.Time      `db:"created_date"`
	UpdatedDate          time.Time      `db:"updated_date"`
}

func toDBPurchaseOrderLineItem(bus purchaseorderlineitembus.PurchaseOrderLineItem) purchaseOrderLineItem {
	db := purchaseOrderLineItem{
		ID:                bus.ID,
		PurchaseOrderID:   bus.PurchaseOrderID,
		SupplierProductID: bus.SupplierProductID,
		QuantityOrdered:   bus.QuantityOrdered,
		QuantityReceived:  bus.QuantityReceived,
		QuantityCancelled: bus.QuantityCancelled,
		UnitCost:          bus.UnitCost,
		Discount:          bus.Discount,
		LineTotal:         bus.LineTotal,
		LineItemStatusID:  bus.LineItemStatusID,
		CreatedBy:         bus.CreatedBy,
		UpdatedBy:         bus.UpdatedBy,
		CreatedDate:       bus.CreatedDate,
		UpdatedDate:       bus.UpdatedDate,
	}

	if !bus.ExpectedDeliveryDate.IsZero() {
		db.ExpectedDeliveryDate = sql.NullTime{Time: bus.ExpectedDeliveryDate, Valid: true}
	}

	if !bus.ActualDeliveryDate.IsZero() {
		db.ActualDeliveryDate = sql.NullTime{Time: bus.ActualDeliveryDate, Valid: true}
	}

	if bus.Notes != "" {
		db.Notes = sql.NullString{String: bus.Notes, Valid: true}
	}

	return db
}

func toBusPurchaseOrderLineItem(db purchaseOrderLineItem) purchaseorderlineitembus.PurchaseOrderLineItem {
	bus := purchaseorderlineitembus.PurchaseOrderLineItem{
		ID:                db.ID,
		PurchaseOrderID:   db.PurchaseOrderID,
		SupplierProductID: db.SupplierProductID,
		QuantityOrdered:   db.QuantityOrdered,
		QuantityReceived:  db.QuantityReceived,
		QuantityCancelled: db.QuantityCancelled,
		UnitCost:          db.UnitCost,
		Discount:          db.Discount,
		LineTotal:         db.LineTotal,
		LineItemStatusID:  db.LineItemStatusID,
		CreatedBy:         db.CreatedBy,
		UpdatedBy:         db.UpdatedBy,
		CreatedDate:       db.CreatedDate,
		UpdatedDate:       db.UpdatedDate,
	}

	if db.ExpectedDeliveryDate.Valid {
		bus.ExpectedDeliveryDate = db.ExpectedDeliveryDate.Time
	}

	if db.ActualDeliveryDate.Valid {
		bus.ActualDeliveryDate = db.ActualDeliveryDate.Time
	}

	if db.Notes.Valid {
		bus.Notes = db.Notes.String
	}

	return bus
}

func toBusPurchaseOrderLineItems(dbs []purchaseOrderLineItem) []purchaseorderlineitembus.PurchaseOrderLineItem {
	items := make([]purchaseorderlineitembus.PurchaseOrderLineItem, len(dbs))
	for i, db := range dbs {
		items[i] = toBusPurchaseOrderLineItem(db)
	}
	return items
}
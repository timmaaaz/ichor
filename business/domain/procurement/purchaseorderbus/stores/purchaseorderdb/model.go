package purchaseorderdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
)

type purchaseOrder struct {
	ID                      uuid.UUID      `db:"id"`
	OrderNumber             string         `db:"order_number"`
	SupplierID              uuid.UUID      `db:"supplier_id"`
	PurchaseOrderStatusID   uuid.UUID      `db:"purchase_order_status_id"`
	DeliveryWarehouseID     uuid.UUID      `db:"delivery_warehouse_id"`
	DeliveryLocationID      uuid.NullUUID  `db:"delivery_location_id"`
	DeliveryStreetID        uuid.NullUUID  `db:"delivery_street_id"`
	OrderDate               time.Time      `db:"order_date"`
	ExpectedDeliveryDate    time.Time      `db:"expected_delivery_date"`
	ActualDeliveryDate      sql.NullTime   `db:"actual_delivery_date"`
	Subtotal                float64        `db:"subtotal"`
	TaxAmount               float64        `db:"tax_amount"`
	ShippingCost            float64        `db:"shipping_cost"`
	TotalAmount             float64        `db:"total_amount"`
	CurrencyID              uuid.UUID      `db:"currency_id"`
	RequestedBy             uuid.UUID      `db:"requested_by"`
	ApprovedBy              uuid.NullUUID  `db:"approved_by"`
	ApprovedDate            sql.NullTime   `db:"approved_date"`
	ApprovalReason          sql.NullString `db:"approval_reason"`
	RejectedBy              uuid.NullUUID  `db:"rejected_by"`
	RejectedDate            sql.NullTime   `db:"rejected_date"`
	RejectionReason         sql.NullString `db:"rejection_reason"`
	Notes                   sql.NullString `db:"notes"`
	SupplierReferenceNumber sql.NullString `db:"supplier_reference_number"`
	Priority                string         `db:"priority"`
	CreatedBy               uuid.UUID      `db:"created_by"`
	UpdatedBy               uuid.UUID      `db:"updated_by"`
	CreatedDate             time.Time      `db:"created_date"`
	UpdatedDate             time.Time      `db:"updated_date"`
	ScenarioID              *uuid.UUID     `db:"scenario_id"`
}

func toDBPurchaseOrder(bus purchaseorderbus.PurchaseOrder) purchaseOrder {
	db := purchaseOrder{
		ID:                    bus.ID,
		OrderNumber:           bus.OrderNumber,
		SupplierID:            bus.SupplierID,
		PurchaseOrderStatusID: bus.PurchaseOrderStatusID,
		DeliveryWarehouseID:   bus.DeliveryWarehouseID,
		OrderDate:             bus.OrderDate,
		ExpectedDeliveryDate:  bus.ExpectedDeliveryDate,
		Subtotal:              bus.Subtotal,
		TaxAmount:             bus.TaxAmount,
		ShippingCost:          bus.ShippingCost,
		TotalAmount:           bus.TotalAmount,
		CurrencyID:            bus.CurrencyID,
		RequestedBy:           bus.RequestedBy,
		Priority:              bus.Priority,
		CreatedBy:             bus.CreatedBy,
		UpdatedBy:             bus.UpdatedBy,
		CreatedDate:           bus.CreatedDate,
		UpdatedDate:           bus.UpdatedDate,
		ScenarioID:            bus.ScenarioID,
	}

	if bus.DeliveryLocationID != uuid.Nil {
		db.DeliveryLocationID = uuid.NullUUID{UUID: bus.DeliveryLocationID, Valid: true}
	}

	if bus.DeliveryStreetID != uuid.Nil {
		db.DeliveryStreetID = uuid.NullUUID{UUID: bus.DeliveryStreetID, Valid: true}
	}

	if !bus.ActualDeliveryDate.IsZero() {
		db.ActualDeliveryDate = sql.NullTime{Time: bus.ActualDeliveryDate, Valid: true}
	}

	if bus.ApprovedBy != nil {
		db.ApprovedBy = uuid.NullUUID{UUID: *bus.ApprovedBy, Valid: true}
	}

	if !bus.ApprovedDate.IsZero() {
		db.ApprovedDate = sql.NullTime{Time: bus.ApprovedDate, Valid: true}
	}

	if bus.ApprovalReason != "" {
		db.ApprovalReason = sql.NullString{String: bus.ApprovalReason, Valid: true}
	}

	if bus.RejectedBy != nil {
		db.RejectedBy = uuid.NullUUID{UUID: *bus.RejectedBy, Valid: true}
	}

	if !bus.RejectedDate.IsZero() {
		db.RejectedDate = sql.NullTime{Time: bus.RejectedDate, Valid: true}
	}

	if bus.RejectionReason != "" {
		db.RejectionReason = sql.NullString{String: bus.RejectionReason, Valid: true}
	}

	if bus.Notes != "" {
		db.Notes = sql.NullString{String: bus.Notes, Valid: true}
	}

	if bus.SupplierReferenceNumber != "" {
		db.SupplierReferenceNumber = sql.NullString{String: bus.SupplierReferenceNumber, Valid: true}
	}

	return db
}

func toBusPurchaseOrder(db purchaseOrder) purchaseorderbus.PurchaseOrder {
	bus := purchaseorderbus.PurchaseOrder{
		ID:                    db.ID,
		OrderNumber:           db.OrderNumber,
		SupplierID:            db.SupplierID,
		PurchaseOrderStatusID: db.PurchaseOrderStatusID,
		DeliveryWarehouseID:   db.DeliveryWarehouseID,
		OrderDate:             db.OrderDate,
		ExpectedDeliveryDate:  db.ExpectedDeliveryDate,
		Subtotal:              db.Subtotal,
		TaxAmount:             db.TaxAmount,
		ShippingCost:          db.ShippingCost,
		TotalAmount:           db.TotalAmount,
		CurrencyID:            db.CurrencyID,
		RequestedBy:           db.RequestedBy,
		Priority:              db.Priority,
		CreatedBy:             db.CreatedBy,
		UpdatedBy:             db.UpdatedBy,
		CreatedDate:           db.CreatedDate,
		UpdatedDate:           db.UpdatedDate,
		ScenarioID:            db.ScenarioID,
	}

	if db.DeliveryLocationID.Valid {
		bus.DeliveryLocationID = db.DeliveryLocationID.UUID
	}

	if db.DeliveryStreetID.Valid {
		bus.DeliveryStreetID = db.DeliveryStreetID.UUID
	}

	if db.ActualDeliveryDate.Valid {
		bus.ActualDeliveryDate = db.ActualDeliveryDate.Time
	}

	if db.ApprovedBy.Valid {
		id := db.ApprovedBy.UUID
		bus.ApprovedBy = &id
	}

	if db.ApprovedDate.Valid {
		bus.ApprovedDate = db.ApprovedDate.Time
	}

	if db.ApprovalReason.Valid {
		bus.ApprovalReason = db.ApprovalReason.String
	}

	if db.RejectedBy.Valid {
		id := db.RejectedBy.UUID
		bus.RejectedBy = &id
	}

	if db.RejectedDate.Valid {
		bus.RejectedDate = db.RejectedDate.Time
	}

	if db.RejectionReason.Valid {
		bus.RejectionReason = db.RejectionReason.String
	}

	if db.Notes.Valid {
		bus.Notes = db.Notes.String
	}

	if db.SupplierReferenceNumber.Valid {
		bus.SupplierReferenceNumber = db.SupplierReferenceNumber.String
	}

	return bus
}

func toBusPurchaseOrders(dbs []purchaseOrder) []purchaseorderbus.PurchaseOrder {
	orders := make([]purchaseorderbus.PurchaseOrder, len(dbs))
	for i, db := range dbs {
		orders[i] = toBusPurchaseOrder(db)
	}
	return orders
}
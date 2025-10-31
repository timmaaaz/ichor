package purchaseorderlineitembus

import (
	"time"

	"github.com/google/uuid"
)

// PurchaseOrderLineItem represents a line item on a purchase order.
type PurchaseOrderLineItem struct {
	ID                   uuid.UUID
	PurchaseOrderID      uuid.UUID
	SupplierProductID    uuid.UUID
	QuantityOrdered      int
	QuantityReceived     int
	QuantityCancelled    int
	UnitCost             float64
	Discount             float64
	LineTotal            float64
	LineItemStatusID     uuid.UUID
	ExpectedDeliveryDate time.Time
	ActualDeliveryDate   time.Time
	Notes                string
	CreatedBy            uuid.UUID
	UpdatedBy            uuid.UUID
	CreatedDate          time.Time
	UpdatedDate          time.Time
}

// NewPurchaseOrderLineItem contains information needed to create a new purchase order line item.
type NewPurchaseOrderLineItem struct {
	PurchaseOrderID      uuid.UUID
	SupplierProductID    uuid.UUID
	QuantityOrdered      int
	UnitCost             float64
	Discount             float64
	LineTotal            float64
	LineItemStatusID     uuid.UUID
	ExpectedDeliveryDate time.Time
	Notes                string
	CreatedBy            uuid.UUID
}

// UpdatePurchaseOrderLineItem contains information needed to update a purchase order line item.
type UpdatePurchaseOrderLineItem struct {
	SupplierProductID    *uuid.UUID
	QuantityOrdered      *int
	QuantityReceived     *int
	QuantityCancelled    *int
	UnitCost             *float64
	Discount             *float64
	LineTotal            *float64
	LineItemStatusID     *uuid.UUID
	ExpectedDeliveryDate *time.Time
	ActualDeliveryDate   *time.Time
	Notes                *string
	UpdatedBy            *uuid.UUID
}
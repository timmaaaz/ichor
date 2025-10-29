package purchaseorderbus

import (
	"time"

	"github.com/google/uuid"
)

// PurchaseOrder represents a purchase order in the system.
type PurchaseOrder struct {
	ID                       uuid.UUID
	OrderNumber              string
	SupplierID               uuid.UUID
	PurchaseOrderStatusID    uuid.UUID
	DeliveryWarehouseID      uuid.UUID
	DeliveryLocationID       uuid.UUID
	DeliveryStreetID         uuid.UUID
	OrderDate                time.Time
	ExpectedDeliveryDate     time.Time
	ActualDeliveryDate       time.Time
	Subtotal                 float64
	TaxAmount                float64
	ShippingCost             float64
	TotalAmount              float64
	Currency                 string
	RequestedBy              uuid.UUID
	ApprovedBy               uuid.UUID
	ApprovedDate             time.Time
	Notes                    string
	SupplierReferenceNumber  string
	CreatedBy                uuid.UUID
	UpdatedBy                uuid.UUID
	CreatedDate              time.Time
	UpdatedDate              time.Time
}

// NewPurchaseOrder contains information needed to create a new purchase order.
type NewPurchaseOrder struct {
	OrderNumber              string
	SupplierID               uuid.UUID
	PurchaseOrderStatusID    uuid.UUID
	DeliveryWarehouseID      uuid.UUID
	DeliveryLocationID       uuid.UUID
	DeliveryStreetID         uuid.UUID
	OrderDate                time.Time
	ExpectedDeliveryDate     time.Time
	Subtotal                 float64
	TaxAmount                float64
	ShippingCost             float64
	TotalAmount              float64
	Currency                 string
	RequestedBy              uuid.UUID
	Notes                    string
	SupplierReferenceNumber  string
	CreatedBy                uuid.UUID
}

// UpdatePurchaseOrder contains information needed to update a purchase order.
type UpdatePurchaseOrder struct {
	OrderNumber              *string
	SupplierID               *uuid.UUID
	PurchaseOrderStatusID    *uuid.UUID
	DeliveryWarehouseID      *uuid.UUID
	DeliveryLocationID       *uuid.UUID
	DeliveryStreetID         *uuid.UUID
	OrderDate                *time.Time
	ExpectedDeliveryDate     *time.Time
	ActualDeliveryDate       *time.Time
	Subtotal                 *float64
	TaxAmount                *float64
	ShippingCost             *float64
	TotalAmount              *float64
	Currency                 *string
	ApprovedBy               *uuid.UUID
	ApprovedDate             *time.Time
	Notes                    *string
	SupplierReferenceNumber  *string
	UpdatedBy                *uuid.UUID
}
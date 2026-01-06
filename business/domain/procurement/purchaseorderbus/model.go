package purchaseorderbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// PurchaseOrder represents a purchase order in the system.
type PurchaseOrder struct {
	ID                      uuid.UUID `json:"id"`
	OrderNumber             string    `json:"order_number"`
	SupplierID              uuid.UUID `json:"supplier_id"`
	PurchaseOrderStatusID   uuid.UUID `json:"purchase_order_status_id"`
	DeliveryWarehouseID     uuid.UUID `json:"delivery_warehouse_id"`
	DeliveryLocationID      uuid.UUID `json:"delivery_location_id"`
	DeliveryStreetID        uuid.UUID `json:"delivery_street_id"`
	OrderDate               time.Time `json:"order_date"`
	ExpectedDeliveryDate    time.Time `json:"expected_delivery_date"`
	ActualDeliveryDate      time.Time `json:"actual_delivery_date"`
	Subtotal                float64   `json:"subtotal"`
	TaxAmount               float64   `json:"tax_amount"`
	ShippingCost            float64   `json:"shipping_cost"`
	TotalAmount             float64   `json:"total_amount"`
	Currency                string    `json:"currency"`
	RequestedBy             uuid.UUID `json:"requested_by"`
	ApprovedBy              uuid.UUID `json:"approved_by"`
	ApprovedDate            time.Time `json:"approved_date"`
	Notes                   string    `json:"notes"`
	SupplierReferenceNumber string    `json:"supplier_reference_number"`
	CreatedBy               uuid.UUID `json:"created_by"`
	UpdatedBy               uuid.UUID `json:"updated_by"`
	CreatedDate             time.Time `json:"created_date"`
	UpdatedDate             time.Time `json:"updated_date"`
}

// NewPurchaseOrder contains information needed to create a new purchase order.
type NewPurchaseOrder struct {
	OrderNumber             string     `json:"order_number"`
	SupplierID              uuid.UUID  `json:"supplier_id"`
	PurchaseOrderStatusID   uuid.UUID  `json:"purchase_order_status_id"`
	DeliveryWarehouseID     uuid.UUID  `json:"delivery_warehouse_id"`
	DeliveryLocationID      uuid.UUID  `json:"delivery_location_id"`
	DeliveryStreetID        uuid.UUID  `json:"delivery_street_id"`
	OrderDate               time.Time  `json:"order_date"`
	ExpectedDeliveryDate    time.Time  `json:"expected_delivery_date"`
	Subtotal                float64    `json:"subtotal"`
	TaxAmount               float64    `json:"tax_amount"`
	ShippingCost            float64    `json:"shipping_cost"`
	TotalAmount             float64    `json:"total_amount"`
	Currency                string     `json:"currency"`
	RequestedBy             uuid.UUID  `json:"requested_by"`
	Notes                   string     `json:"notes"`
	SupplierReferenceNumber string     `json:"supplier_reference_number"`
	CreatedBy               uuid.UUID  `json:"created_by"`
	CreatedDate             *time.Time `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

// UpdatePurchaseOrder contains information needed to update a purchase order.
type UpdatePurchaseOrder struct {
	OrderNumber             *string    `json:"order_number,omitempty"`
	SupplierID              *uuid.UUID `json:"supplier_id,omitempty"`
	PurchaseOrderStatusID   *uuid.UUID `json:"purchase_order_status_id,omitempty"`
	DeliveryWarehouseID     *uuid.UUID `json:"delivery_warehouse_id,omitempty"`
	DeliveryLocationID      *uuid.UUID `json:"delivery_location_id,omitempty"`
	DeliveryStreetID        *uuid.UUID `json:"delivery_street_id,omitempty"`
	OrderDate               *time.Time `json:"order_date,omitempty"`
	ExpectedDeliveryDate    *time.Time `json:"expected_delivery_date,omitempty"`
	ActualDeliveryDate      *time.Time `json:"actual_delivery_date,omitempty"`
	Subtotal                *float64   `json:"subtotal,omitempty"`
	TaxAmount               *float64   `json:"tax_amount,omitempty"`
	ShippingCost            *float64   `json:"shipping_cost,omitempty"`
	TotalAmount             *float64   `json:"total_amount,omitempty"`
	Currency                *string    `json:"currency,omitempty"`
	ApprovedBy              *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedDate            *time.Time `json:"approved_date,omitempty"`
	Notes                   *string    `json:"notes,omitempty"`
	SupplierReferenceNumber *string    `json:"supplier_reference_number,omitempty"`
	UpdatedBy               *uuid.UUID `json:"updated_by,omitempty"`
}
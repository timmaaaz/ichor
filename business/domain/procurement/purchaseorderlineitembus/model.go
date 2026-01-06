package purchaseorderlineitembus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// PurchaseOrderLineItem represents a line item on a purchase order.
type PurchaseOrderLineItem struct {
	ID                   uuid.UUID `json:"id"`
	PurchaseOrderID      uuid.UUID `json:"purchase_order_id"`
	SupplierProductID    uuid.UUID `json:"supplier_product_id"`
	QuantityOrdered      int       `json:"quantity_ordered"`
	QuantityReceived     int       `json:"quantity_received"`
	QuantityCancelled    int       `json:"quantity_cancelled"`
	UnitCost             float64   `json:"unit_cost"`
	Discount             float64   `json:"discount"`
	LineTotal            float64   `json:"line_total"`
	LineItemStatusID     uuid.UUID `json:"line_item_status_id"`
	ExpectedDeliveryDate time.Time `json:"expected_delivery_date"`
	ActualDeliveryDate   time.Time `json:"actual_delivery_date"`
	Notes                string    `json:"notes"`
	CreatedBy            uuid.UUID `json:"created_by"`
	UpdatedBy            uuid.UUID `json:"updated_by"`
	CreatedDate          time.Time `json:"created_date"`
	UpdatedDate          time.Time `json:"updated_date"`
}

// NewPurchaseOrderLineItem contains information needed to create a new purchase order line item.
type NewPurchaseOrderLineItem struct {
	PurchaseOrderID      uuid.UUID `json:"purchase_order_id"`
	SupplierProductID    uuid.UUID `json:"supplier_product_id"`
	QuantityOrdered      int       `json:"quantity_ordered"`
	UnitCost             float64   `json:"unit_cost"`
	Discount             float64   `json:"discount"`
	LineTotal            float64   `json:"line_total"`
	LineItemStatusID     uuid.UUID `json:"line_item_status_id"`
	ExpectedDeliveryDate time.Time `json:"expected_delivery_date"`
	Notes                string    `json:"notes"`
	CreatedBy            uuid.UUID `json:"created_by"`
}

// UpdatePurchaseOrderLineItem contains information needed to update a purchase order line item.
type UpdatePurchaseOrderLineItem struct {
	SupplierProductID    *uuid.UUID `json:"supplier_product_id,omitempty"`
	QuantityOrdered      *int       `json:"quantity_ordered,omitempty"`
	QuantityReceived     *int       `json:"quantity_received,omitempty"`
	QuantityCancelled    *int       `json:"quantity_cancelled,omitempty"`
	UnitCost             *float64   `json:"unit_cost,omitempty"`
	Discount             *float64   `json:"discount,omitempty"`
	LineTotal            *float64   `json:"line_total,omitempty"`
	LineItemStatusID     *uuid.UUID `json:"line_item_status_id,omitempty"`
	ExpectedDeliveryDate *time.Time `json:"expected_delivery_date,omitempty"`
	ActualDeliveryDate   *time.Time `json:"actual_delivery_date,omitempty"`
	Notes                *string    `json:"notes,omitempty"`
	UpdatedBy            *uuid.UUID `json:"updated_by,omitempty"`
}
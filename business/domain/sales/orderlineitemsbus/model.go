package orderlineitemsbus

import (
	"time"

	"github.com/google/uuid"
)

// OrderLineItem represents an order line item in the system.
// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.
type OrderLineItem struct {
	ID                            uuid.UUID `json:"id"`
	OrderID                       uuid.UUID `json:"order_id"`
	ProductID                     uuid.UUID `json:"product_id"`
	Quantity                      int       `json:"quantity"`
	Discount                      float64   `json:"discount"`
	LineItemFulfillmentStatusesID uuid.UUID `json:"line_item_fulfillment_statuses_id"`
	CreatedBy                     uuid.UUID `json:"created_by"`
	CreatedDate                   time.Time `json:"created_date"`
	UpdatedBy                     uuid.UUID `json:"updated_by"`
	UpdatedDate                   time.Time `json:"updated_date"`
}

type NewOrderLineItem struct {
	OrderID                       uuid.UUID  `json:"order_id"`
	ProductID                     uuid.UUID  `json:"product_id"`
	Quantity                      int        `json:"quantity"`
	Discount                      float64    `json:"discount"`
	LineItemFulfillmentStatusesID uuid.UUID  `json:"line_item_fulfillment_statuses_id"`
	CreatedBy                     uuid.UUID  `json:"created_by"`
	CreatedDate                   *time.Time `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateOrderLineItem struct {
	OrderID                       *uuid.UUID `json:"order_id,omitempty"`
	ProductID                     *uuid.UUID `json:"product_id,omitempty"`
	Quantity                      *int       `json:"quantity,omitempty"`
	Discount                      *float64   `json:"discount,omitempty"`
	LineItemFulfillmentStatusesID *uuid.UUID `json:"line_item_fulfillment_statuses_id,omitempty"`
	UpdatedBy                     *uuid.UUID `json:"updated_by,omitempty"`
}

package orderlineitemsbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
)

// OrderLineItem represents an order line item in the system.
// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.
type OrderLineItem struct {
	ID                            uuid.UUID   `json:"id"`
	OrderID                       uuid.UUID   `json:"order_id"`
	ProductID                     uuid.UUID   `json:"product_id"`
	Description                   string      `json:"description"`
	Quantity                      int         `json:"quantity"`
	UnitPrice                     types.Money `json:"unit_price"`
	Discount                      types.Money `json:"discount"`
	DiscountType                  string      `json:"discount_type"`
	LineTotal                     types.Money `json:"line_total"`
	LineItemFulfillmentStatusesID uuid.UUID   `json:"line_item_fulfillment_statuses_id"`
	PickedQuantity                int         `json:"picked_quantity"`
	BackorderedQuantity           int         `json:"backordered_quantity"`
	ShortPickReason               *string     `json:"short_pick_reason,omitempty"`
	CreatedBy                     uuid.UUID   `json:"created_by"`
	CreatedDate                   time.Time   `json:"created_date"`
	UpdatedBy                     uuid.UUID   `json:"updated_by"`
	UpdatedDate                   time.Time   `json:"updated_date"`
}

type NewOrderLineItem struct {
	OrderID                       uuid.UUID   `json:"order_id"`
	ProductID                     uuid.UUID   `json:"product_id"`
	Description                   string      `json:"description"`
	Quantity                      int         `json:"quantity"`
	UnitPrice                     types.Money `json:"unit_price"`
	Discount                      types.Money `json:"discount"`
	DiscountType                  string      `json:"discount_type"` // 'flat' or 'percent'
	LineTotal                     types.Money `json:"line_total"`
	LineItemFulfillmentStatusesID uuid.UUID   `json:"line_item_fulfillment_statuses_id"`
	CreatedBy                     uuid.UUID   `json:"created_by"`
	CreatedDate                   *time.Time  `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateOrderLineItem struct {
	OrderID                       *uuid.UUID    `json:"order_id,omitempty"`
	ProductID                     *uuid.UUID    `json:"product_id,omitempty"`
	Description                   *string       `json:"description,omitempty"`
	Quantity                      *int          `json:"quantity,omitempty"`
	UnitPrice                     *types.Money  `json:"unit_price,omitempty"`
	Discount                      *types.Money  `json:"discount,omitempty"`
	DiscountType                  *string       `json:"discount_type,omitempty"`
	LineTotal                     *types.Money  `json:"line_total,omitempty"`
	LineItemFulfillmentStatusesID *uuid.UUID    `json:"line_item_fulfillment_statuses_id,omitempty"`
	PickedQuantity                *int          `json:"picked_quantity,omitempty"`
	BackorderedQuantity           *int          `json:"backordered_quantity,omitempty"`
	ShortPickReason               *string       `json:"short_pick_reason,omitempty"`
	UpdatedBy                     *uuid.UUID    `json:"updated_by,omitempty"`
}

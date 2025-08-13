package orderlineitemsbus

import (
	"time"

	"github.com/google/uuid"
)

type OrderLineItem struct {
	ID                            uuid.UUID
	OrderID                       uuid.UUID
	ProductID                     uuid.UUID
	Quantity                      int
	Discount                      float64
	LineItemFulfillmentStatusesID uuid.UUID
	CreatedBy                     uuid.UUID
	CreatedDate                   time.Time
	UpdatedBy                     uuid.UUID
	UpdatedDate                   time.Time
}

type NewOrderLineItem struct {
	OrderID                       uuid.UUID
	ProductID                     uuid.UUID
	Quantity                      int
	Discount                      float64
	LineItemFulfillmentStatusesID uuid.UUID
	CreatedBy                     uuid.UUID
}

type UpdateOrderLineItem struct {
	OrderID                       *uuid.UUID
	ProductID                     *uuid.UUID
	Quantity                      *int
	Discount                      *float64
	LineItemFulfillmentStatusesID *uuid.UUID
	UpdatedBy                     *uuid.UUID
}

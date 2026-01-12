package orderlineitemsbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                            *uuid.UUID
	OrderID                       *uuid.UUID
	ProductID                     *uuid.UUID
	Description                   *string
	Quantity                      *int
	UnitPrice                     *string
	Discount                      *string
	DiscountType                  *string
	LineTotal                     *string
	LineItemFulfillmentStatusesID *uuid.UUID
	CreatedBy                     *uuid.UUID
	StartCreatedDate              *time.Time
	EndCreatedDate                *time.Time
	UpdatedBy                     *uuid.UUID
	StartUpdatedDate              *time.Time
	EndUpdatedDate                *time.Time
}

package orderlineitemsbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                            *uuid.UUID
	OrderID                       *uuid.UUID
	ProductID                     *uuid.UUID
	Quantity                      *int
	Discount                      *float64
	LineItemFulfillmentStatusesID *uuid.UUID
	CreatedBy                     *uuid.UUID
	StartCreatedDate              *time.Time
	EndCreatedDate                *time.Time
	UpdatedBy                     *uuid.UUID
	StartUpdatedDate              *time.Time
	EndUpdatedDate                *time.Time
}

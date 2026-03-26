package picktaskbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds optional filters for querying pick tasks.
type QueryFilter struct {
	ID                   *uuid.UUID
	SalesOrderID         *uuid.UUID
	SalesOrderLineItemID *uuid.UUID
	ProductID            *uuid.UUID
	LocationID           *uuid.UUID
	Status               *Status
	AssignedTo           *uuid.UUID
	CreatedBy            *uuid.UUID
	CreatedDate          *time.Time
	UpdatedDate          *time.Time
}

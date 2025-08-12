package ordersbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                  *uuid.UUID
	Number              *string
	CustomerID          *uuid.UUID
	FulfillmentStatusID *uuid.UUID
	CreatedBy           *uuid.UUID
	UpdatedBy           *uuid.UUID
	StartDueDate        *time.Time
	EndDueDate          *time.Time
	StartCreatedDate    *time.Time
	EndCreatedDate      *time.Time
	StartUpdatedDate    *time.Time
	EndUpdatedDate      *time.Time
}

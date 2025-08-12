package ordersbus

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID                  uuid.UUID
	Number              string
	CustomerID          uuid.UUID
	DueDate             time.Time
	FulfillmentStatusID uuid.UUID
	CreatedBy           uuid.UUID
	UpdatedBy           uuid.UUID
	CreatedDate         time.Time
	UpdatedDate         time.Time
}

type NewOrder struct {
	Number              string
	CustomerID          uuid.UUID
	DueDate             time.Time
	FulfillmentStatusID uuid.UUID
	CreatedBy           uuid.UUID
}

type UpdateOrder struct {
	Number              *string
	CustomerID          *uuid.UUID
	DueDate             *time.Time
	FulfillmentStatusID *uuid.UUID
	UpdatedBy           *uuid.UUID
}

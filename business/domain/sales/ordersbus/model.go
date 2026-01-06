package ordersbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Order struct {
	ID                  uuid.UUID `json:"id"`
	Number              string    `json:"number"`
	CustomerID          uuid.UUID `json:"customer_id"`
	DueDate             time.Time `json:"due_date"`
	FulfillmentStatusID uuid.UUID `json:"fulfillment_status_id"`
	CreatedBy           uuid.UUID `json:"created_by"`
	UpdatedBy           uuid.UUID `json:"updated_by"`
	CreatedDate         time.Time `json:"created_date"`
	UpdatedDate         time.Time `json:"updated_date"`
}

type NewOrder struct {
	Number              string     `json:"number"`
	CustomerID          uuid.UUID  `json:"customer_id"`
	DueDate             time.Time  `json:"due_date"`
	FulfillmentStatusID uuid.UUID  `json:"fulfillment_status_id"`
	CreatedBy           uuid.UUID  `json:"created_by"`
	CreatedDate         *time.Time `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateOrder struct {
	Number              *string    `json:"number,omitempty"`
	CustomerID          *uuid.UUID `json:"customer_id,omitempty"`
	DueDate             *time.Time `json:"due_date,omitempty"`
	FulfillmentStatusID *uuid.UUID `json:"fulfillment_status_id,omitempty"`
	UpdatedBy           *uuid.UUID `json:"updated_by,omitempty"`
}

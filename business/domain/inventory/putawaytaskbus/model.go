package putawaytaskbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// PutAwayTask represents a single directed work instruction for a floor worker:
// take Quantity units of ProductID and place them at LocationID.
type PutAwayTask struct {
	ID              uuid.UUID `json:"id"`
	ProductID       uuid.UUID `json:"product_id"`
	LocationID      uuid.UUID `json:"location_id"`
	Quantity        int       `json:"quantity"`
	ReferenceNumber string    `json:"reference_number"`
	Status          Status    `json:"status"`
	AssignedTo      uuid.UUID `json:"assigned_to"`  // uuid.Nil when unassigned
	AssignedAt      time.Time `json:"assigned_at"`   // zero when unassigned
	CompletedBy     uuid.UUID `json:"completed_by"`  // uuid.Nil when not completed
	CompletedAt     time.Time `json:"completed_at"`  // zero when not completed
	CreatedBy       uuid.UUID `json:"created_by"`
	CreatedDate     time.Time `json:"created_date"`
	UpdatedDate     time.Time `json:"updated_date"`
}

// NewPutAwayTask contains the information needed to create a new put-away task.
// Status is always set to Statuses.Pending by the business layer.
type NewPutAwayTask struct {
	ProductID       uuid.UUID `json:"product_id"`
	LocationID      uuid.UUID `json:"location_id"`
	Quantity        int       `json:"quantity"`
	ReferenceNumber string    `json:"reference_number"`
	CreatedBy       uuid.UUID `json:"created_by"`
}

// UpdatePutAwayTask contains the information that can be changed on a put-away task.
// All fields are optional pointers; nil means "do not update this field."
type UpdatePutAwayTask struct {
	ProductID       *uuid.UUID `json:"product_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	Quantity        *int       `json:"quantity,omitempty"`
	ReferenceNumber *string    `json:"reference_number,omitempty"`
	Status          *Status    `json:"status,omitempty"`
	AssignedTo      *uuid.UUID `json:"assigned_to,omitempty"`
	AssignedAt      *time.Time `json:"assigned_at,omitempty"`
	CompletedBy     *uuid.UUID `json:"completed_by,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

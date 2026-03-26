package cyclecountsessionbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// marshals business models to JSON for RawData in TriggerEvents.

// CycleCountSession represents a cycle count session in the warehouse.
type CycleCountSession struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Status        Status    `json:"status"`
	CreatedBy     uuid.UUID `json:"created_by"`
	CreatedDate   time.Time `json:"created_date"`
	UpdatedDate   time.Time `json:"updated_date"`
	CompletedDate time.Time `json:"completed_date"`
}

// NewCycleCountSession contains the information needed to create a new cycle count session.
// Status is always set to Statuses.Draft by the business layer.
type NewCycleCountSession struct {
	Name      string    `json:"name"`
	CreatedBy uuid.UUID `json:"created_by"`
}

// UpdateCycleCountSession contains the information that can be changed on a cycle count session.
// All fields are optional pointers; nil means "do not update this field."
type UpdateCycleCountSession struct {
	Name          *string    `json:"name,omitempty"`
	Status        *Status    `json:"status,omitempty"`
	CompletedDate *time.Time `json:"completed_date,omitempty"`
}

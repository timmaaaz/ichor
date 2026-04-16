package cyclecountitembus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// marshals business models to JSON for RawData in TriggerEvents.

// CycleCountItem represents a single inventory item line within a cycle count session.
// Each item records the system quantity, the counted quantity, and the computed variance.
type CycleCountItem struct {
	ID               uuid.UUID `json:"id"`
	ItemCode         *string   `json:"item_code,omitempty"`
	SessionID        uuid.UUID `json:"session_id"`
	ProductID        uuid.UUID `json:"product_id"`
	LocationID       uuid.UUID `json:"location_id"`
	SystemQuantity   int       `json:"system_quantity"`
	CountedQuantity  *int      `json:"counted_quantity,omitempty"`
	Variance         *int      `json:"variance,omitempty"`
	Status           Status    `json:"status"`
	CountedBy        uuid.UUID `json:"counted_by"`
	CountedDate      time.Time `json:"counted_date"`
	CreatedDate      time.Time `json:"created_date"`
	UpdatedDate      time.Time `json:"updated_date"`
}

// NewCycleCountItem contains the information needed to create a new cycle count item.
// Status is always set to Statuses.Pending by the business layer.
type NewCycleCountItem struct {
	ItemCode       *string   `json:"item_code,omitempty"`
	SessionID      uuid.UUID `json:"session_id"`
	ProductID      uuid.UUID `json:"product_id"`
	LocationID     uuid.UUID `json:"location_id"`
	SystemQuantity int       `json:"system_quantity"`
}

// UpdateCycleCountItem contains the information that can be changed on a cycle count item.
// All fields are optional pointers; nil means "do not update this field."
// When CountedQuantity is set, Variance is automatically computed by the business layer.
type UpdateCycleCountItem struct {
	ItemCode        *string    `json:"item_code,omitempty"`
	CountedQuantity *int       `json:"counted_quantity,omitempty"`
	Status          *Status    `json:"status,omitempty"`
	CountedBy       *uuid.UUID `json:"counted_by,omitempty"`
	CountedDate     *time.Time `json:"counted_date,omitempty"`
}

package warehousebus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Warehouse struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	StreetID    uuid.UUID `json:"street_id"`
	Name        string    `json:"name"`
	IsActive    bool      `json:"is_active"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
	CreatedBy   uuid.UUID `json:"created_by"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
}

type NewWarehouse struct {
	Code        string     `json:"code"`
	StreetID    uuid.UUID  `json:"street_id"`
	Name        string     `json:"name"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedDate *time.Time `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateWarehouse struct {
	Code      *string    `json:"code,omitempty"`
	StreetID  *uuid.UUID `json:"street_id,omitempty"`
	Name      *string    `json:"name,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

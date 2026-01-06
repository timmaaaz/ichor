package zonebus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Zone struct {
	ZoneID      uuid.UUID `json:"zone_id"`
	WarehouseID uuid.UUID `json:"warehouse_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

type NewZone struct {
	WarehouseID uuid.UUID `json:"warehouse_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type UpdateZone struct {
	WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"`
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
}

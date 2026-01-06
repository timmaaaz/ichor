package brandbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Brand struct {
	BrandID        uuid.UUID `json:"brand_id"`
	Name           string    `json:"name"`
	ContactInfosID uuid.UUID `json:"contact_infos_id"`
	CreatedDate    time.Time `json:"created_date"`
	UpdatedDate    time.Time `json:"updated_date"`
}

type NewBrand struct {
	Name           string    `json:"name"`
	ContactInfosID uuid.UUID `json:"contact_infos_id"`
}

type UpdateBrand struct {
	Name           *string    `json:"name,omitempty"`
	ContactInfosID *uuid.UUID `json:"contact_infos_id,omitempty"`
}

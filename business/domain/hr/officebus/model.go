package officebus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Office struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	StreetID uuid.UUID `json:"street_id"`
}

type NewOffice struct {
	Name     string    `json:"name"`
	StreetID uuid.UUID `json:"street_id"`
}

type UpdateOffice struct {
	Name     *string    `json:"name,omitempty"`
	StreetID *uuid.UUID `json:"street_id,omitempty"`
}

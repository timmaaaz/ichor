package citybus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// City represents information about an individual city.
type City struct {
	ID       uuid.UUID `json:"id"`
	RegionID uuid.UUID `json:"region_id"`
	Name     string    `json:"name"`
}

type NewCity struct {
	RegionID uuid.UUID `json:"region_id"`
	Name     string    `json:"name"`
}

type UpdateCity struct {
	RegionID *uuid.UUID `json:"region_id,omitempty"`
	Name     *string    `json:"name,omitempty"`
}

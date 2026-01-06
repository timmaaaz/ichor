package tagbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Tag struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type NewTag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateTag struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

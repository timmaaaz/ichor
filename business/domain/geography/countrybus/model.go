package countrybus

import (
	"github.com/google/uuid"
)

// NOTE: Countries are special. They are controlled solely in the database and
// therefore should have ONLY retrive actions available. No create, update, or
// delete actions are allowed. We want only the highest level admins to have any
// way to touch this.

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// Country represents information about an individual country.
type Country struct {
	ID     uuid.UUID `json:"id"`
	Number int       `json:"number"`
	Name   string    `json:"name"`
	Alpha2 string    `json:"alpha_2"`
	Alpha3 string    `json:"alpha_3"`
}

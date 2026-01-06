package regionbus

import "github.com/google/uuid"

// NOTE: Regions are special. They are controlled solely in the database and
// therefore should have ONLY retrive actions available. No create, update, or
// delete actions are allowed. We want only the highest level admins to have
// any way to touch this because it denotes areas we support.

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Region struct {
	ID        uuid.UUID `json:"id"`
	CountryID uuid.UUID `json:"country_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
}

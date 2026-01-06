package streetbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// Street represents information about an individual street.
type Street struct {
	ID         uuid.UUID `json:"id"`
	CityID     uuid.UUID `json:"city_id"`
	Line1      string    `json:"line_1"`
	Line2      string    `json:"line_2"`
	PostalCode string    `json:"postal_code"`
}

type NewStreet struct {
	CityID     uuid.UUID `json:"city_id"`
	Line1      string    `json:"line_1"`
	Line2      string    `json:"line_2"`
	PostalCode string    `json:"postal_code"`
}

type UpdateStreet struct {
	CityID     *uuid.UUID `json:"city_id,omitempty"`
	Line1      *string    `json:"line_1,omitempty"`
	Line2      *string    `json:"line_2,omitempty"`
	PostalCode *string    `json:"postal_code,omitempty"`
}

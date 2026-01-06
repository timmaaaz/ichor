package orderfulfillmentstatusbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type OrderFulfillmentStatus struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	PrimaryColor   string    `json:"primary_color"`
	SecondaryColor string    `json:"secondary_color"`
	Icon           string    `json:"icon"`
}

type NewOrderFulfillmentStatus struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	Icon           string `json:"icon"`
}

type UpdateOrderFulfillmentStatus struct {
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	PrimaryColor   *string `json:"primary_color,omitempty"`
	SecondaryColor *string `json:"secondary_color,omitempty"`
	Icon           *string `json:"icon,omitempty"`
}

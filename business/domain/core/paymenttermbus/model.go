package paymenttermbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// PaymentTerm represents information about a payment term option.
type PaymentTerm struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type NewPaymentTerm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePaymentTerm struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

package customersbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Customers struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	ContactID         uuid.UUID `json:"contact_id"`
	DeliveryAddressID uuid.UUID `json:"delivery_address_id"`
	Notes             string    `json:"notes"`
	CreatedBy         uuid.UUID `json:"created_by"`
	UpdatedBy         uuid.UUID `json:"updated_by"`
	CreatedDate       time.Time `json:"created_date"`
	UpdatedDate       time.Time `json:"updated_date"`
}

type NewCustomers struct {
	Name              string     `json:"name"`
	ContactID         uuid.UUID  `json:"contact_id"`
	DeliveryAddressID uuid.UUID  `json:"delivery_address_id"`
	Notes             string     `json:"notes"`
	CreatedBy         uuid.UUID  `json:"created_by"`
	CreatedDate       *time.Time `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateCustomers struct {
	Name              *string    `json:"name,omitempty"`
	ContactID         *uuid.UUID `json:"contact_id,omitempty"`
	DeliveryAddressID *uuid.UUID `json:"delivery_address_id,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	UpdatedBy         *uuid.UUID `json:"updated_by,omitempty"`
}

package contactinfosbus

import (
	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type ContactInfos struct {
	ID                   uuid.UUID `json:"id"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	EmailAddress         string    `json:"email_address"`
	PrimaryPhone         string    `json:"primary_phone_number"`
	SecondaryPhone       string    `json:"secondary_phone_number"`
	StreetID             uuid.UUID `json:"street_id"`
	DeliveryAddressID    uuid.UUID `json:"delivery_address_id"`
	AvailableHoursStart  string    `json:"available_hours_start"`
	AvailableHoursEnd    string    `json:"available_hours_end"`
	TimezoneID           uuid.UUID `json:"timezone_id"`
	PreferredContactType string    `json:"preferred_contact_type"`
	Notes                string    `json:"notes"`
}

type NewContactInfos struct {
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	EmailAddress         string    `json:"email_address"`
	PrimaryPhone         string    `json:"primary_phone_number"`
	SecondaryPhone       string    `json:"secondary_phone_number"`
	StreetID             uuid.UUID `json:"street_id"`
	DeliveryAddressID    uuid.UUID `json:"delivery_address_id"`
	AvailableHoursStart  string    `json:"available_hours_start"`
	AvailableHoursEnd    string    `json:"available_hours_end"`
	TimezoneID           uuid.UUID `json:"timezone_id"`
	PreferredContactType string    `json:"preferred_contact_type"`
	Notes                string    `json:"notes"`
}

type UpdateContactInfos struct {
	ID                   *uuid.UUID `json:"id,omitempty"`
	FirstName            *string    `json:"first_name,omitempty"`
	LastName             *string    `json:"last_name,omitempty"`
	EmailAddress         *string    `json:"email_address,omitempty"`
	PrimaryPhone         *string    `json:"primary_phone_number,omitempty"`
	SecondaryPhone       *string    `json:"secondary_phone_number,omitempty"`
	StreetID             *uuid.UUID `json:"street_id,omitempty"`
	DeliveryAddressID    *uuid.UUID `json:"delivery_address_id,omitempty"`
	AvailableHoursStart  *string    `json:"available_hours_start,omitempty"`
	AvailableHoursEnd    *string    `json:"available_hours_end,omitempty"`
	TimezoneID           *uuid.UUID `json:"timezone_id,omitempty"`
	PreferredContactType *string    `json:"preferred_contact_type,omitempty"`
	Notes                *string    `json:"notes,omitempty"`
}

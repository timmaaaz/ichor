package homebus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// Address represents an individual address.
type Address struct {
	Address1 string `json:"address_1"`
	Address2 string `json:"address_2"`
	ZipCode  string `json:"zip_code"`
	City     string `json:"city"`
	State    string `json:"state"`
	Country  string `json:"country"`
}

// Home represents an individual home.
type Home struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Type        Type      `json:"type"`
	Address     Address   `json:"address"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

// NewHome is what we require from clients when adding a Home.
type NewHome struct {
	UserID      uuid.UUID  `json:"user_id"`
	Type        Type       `json:"type"`
	Address     Address    `json:"address"`
	CreatedDate *time.Time `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

// UpdateAddress is what fields can be updated in the store.
type UpdateAddress struct {
	Address1 *string `json:"address_1,omitempty"`
	Address2 *string `json:"address_2,omitempty"`
	ZipCode  *string `json:"zip_code,omitempty"`
	City     *string `json:"city,omitempty"`
	State    *string `json:"state,omitempty"`
	Country  *string `json:"country,omitempty"`
}

// UpdateHome defines what informaton may be provided to modify an existing
// Home. All fields are optional so clients can send only the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exepction around
// marshalling/unmarshalling.
type UpdateHome struct {
	Type    *Type          `json:"type,omitempty"`
	Address *UpdateAddress `json:"address,omitempty"`
}

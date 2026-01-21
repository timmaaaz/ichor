package currencybus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// Currency represents information about an individual currency.
type Currency struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Symbol        string     `json:"symbol"`
	Locale        string     `json:"locale"`
	DecimalPlaces int        `json:"decimal_places"`
	IsActive      bool       `json:"is_active"`
	SortOrder     int        `json:"sort_order"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	CreatedDate   time.Time  `json:"created_date"`
	UpdatedBy     *uuid.UUID `json:"updated_by"`
	UpdatedDate   time.Time  `json:"updated_date"`
}

// NewCurrency contains information needed to create a new currency.
type NewCurrency struct {
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Symbol        string     `json:"symbol"`
	Locale        string     `json:"locale"`
	DecimalPlaces int        `json:"decimal_places"`
	IsActive      bool       `json:"is_active"`
	SortOrder     int        `json:"sort_order"`
	CreatedBy     *uuid.UUID `json:"created_by"`
}

// UpdateCurrency contains information needed to update a currency. Fields that are not
// included are intended to have separate endpoints or permissions to update.
type UpdateCurrency struct {
	Code          *string    `json:"code,omitempty"`
	Name          *string    `json:"name,omitempty"`
	Symbol        *string    `json:"symbol,omitempty"`
	Locale        *string    `json:"locale,omitempty"`
	DecimalPlaces *int       `json:"decimal_places,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
	SortOrder     *int       `json:"sort_order,omitempty"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty"`
}

package timezonebus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// Timezone represents information about an individual timezone.
type Timezone struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	UTCOffset   string    `json:"utc_offset"`
	IsActive    bool      `json:"is_active"`
}

// NewTimezone defines the data needed to add a timezone.
type NewTimezone struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	UTCOffset   string `json:"utc_offset"`
	IsActive    bool   `json:"is_active"`
}

// UpdateTimezone defines the data that can be updated for a timezone.
type UpdateTimezone struct {
	Name        *string `json:"name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	UTCOffset   *string `json:"utc_offset,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

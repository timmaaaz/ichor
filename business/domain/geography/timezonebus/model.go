package timezonebus

import "github.com/google/uuid"

// Timezone represents information about an individual timezone.
type Timezone struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
	UTCOffset   string
	IsActive    bool
}

// NewTimezone defines the data needed to add a timezone.
type NewTimezone struct {
	Name        string
	DisplayName string
	UTCOffset   string
	IsActive    bool
}

// UpdateTimezone defines the data that can be updated for a timezone.
type UpdateTimezone struct {
	Name        *string
	DisplayName *string
	UTCOffset   *string
	IsActive    *bool
}

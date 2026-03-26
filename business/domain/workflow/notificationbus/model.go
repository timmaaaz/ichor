package notificationbus

import (
	"time"

	"github.com/google/uuid"
)

// Priority constants for notification importance.
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Notification represents a user notification in the system.
type Notification struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Title            string     `json:"title"`
	Message          string     `json:"message"`
	Priority         string     `json:"priority"`
	IsRead           bool       `json:"is_read"`
	ReadDate         *time.Time `json:"read_date,omitempty"`
	SourceEntityName string     `json:"source_entity_name"`
	SourceEntityID   uuid.UUID  `json:"source_entity_id"`
	ActionURL        string     `json:"action_url"`
	CreatedDate      time.Time  `json:"created_date"`
}

// NewNotification contains the data needed to create a notification.
type NewNotification struct {
	UserID           uuid.UUID
	Title            string
	Message          string
	Priority         string
	SourceEntityName string
	SourceEntityID   uuid.UUID
	ActionURL        string
}

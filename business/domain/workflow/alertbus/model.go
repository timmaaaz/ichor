// Package alertbus provides the core business logic for workflow alerts.
package alertbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status constants for alert state.
const (
	StatusActive       = "active"
	StatusAcknowledged = "acknowledged"
	StatusDismissed    = "dismissed"
)

// Severity constants for alert priority.
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Alert represents a workflow alert in the system.
type Alert struct {
	ID               uuid.UUID
	AlertType        string
	Severity         string
	Title            string
	Message          string
	Context          json.RawMessage
	SourceEntityName string
	SourceEntityID   uuid.UUID
	SourceRuleID     uuid.UUID
	Status           string
	ExpiresDate      *time.Time
	CreatedDate      time.Time
	UpdatedDate      time.Time
}

// AlertRecipient represents a recipient of an alert (user or role).
type AlertRecipient struct {
	ID            uuid.UUID
	AlertID       uuid.UUID
	RecipientType string // "user" or "role"
	RecipientID   uuid.UUID
	CreatedDate   time.Time
}

// AlertAcknowledgment represents a user's acknowledgment of an alert.
type AlertAcknowledgment struct {
	ID               uuid.UUID
	AlertID          uuid.UUID
	AcknowledgedBy   uuid.UUID
	AcknowledgedDate time.Time
	Notes            string
}

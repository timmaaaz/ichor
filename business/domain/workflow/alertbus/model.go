// Package alertbus provides the core business logic for workflow alerts.
package alertbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

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
	ID               uuid.UUID       `json:"id"`
	AlertType        string          `json:"alert_type"`
	Severity         string          `json:"severity"`
	Title            string          `json:"title"`
	Message          string          `json:"message"`
	Context          json.RawMessage `json:"context"`
	SourceEntityName string          `json:"source_entity_name"`
	SourceEntityID   uuid.UUID       `json:"source_entity_id"`
	SourceRuleID     uuid.UUID       `json:"source_rule_id"`
	Status           string          `json:"status"`
	ExpiresDate      *time.Time      `json:"expires_date,omitempty"`
	CreatedDate      time.Time       `json:"created_date"`
	UpdatedDate      time.Time       `json:"updated_date"`
}

// AlertRecipient represents a recipient of an alert (user or role).
type AlertRecipient struct {
	ID            uuid.UUID `json:"id"`
	AlertID       uuid.UUID `json:"alert_id"`
	RecipientType string    `json:"recipient_type"` // "user" or "role"
	RecipientID   uuid.UUID `json:"recipient_id"`
	CreatedDate   time.Time `json:"created_date"`
}

// AlertAcknowledgment represents a user's acknowledgment of an alert.
type AlertAcknowledgment struct {
	ID               uuid.UUID `json:"id"`
	AlertID          uuid.UUID `json:"alert_id"`
	AcknowledgedBy   uuid.UUID `json:"acknowledged_by"`
	AcknowledgedDate time.Time `json:"acknowledged_date"`
	Notes            string    `json:"notes"`
}

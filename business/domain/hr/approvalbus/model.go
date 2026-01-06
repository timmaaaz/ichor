package approvalbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type UserApprovalStatus struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	IconID         uuid.UUID `json:"icon_id"`
	PrimaryColor   string    `json:"primary_color"`
	SecondaryColor string    `json:"secondary_color"`
	Icon           string    `json:"icon"`
}

type NewUserApprovalStatus struct {
	Name           string    `json:"name"`
	IconID         uuid.UUID `json:"icon_id"`
	PrimaryColor   string    `json:"primary_color"`
	SecondaryColor string    `json:"secondary_color"`
	Icon           string    `json:"icon"`
}

type UpdateUserApprovalStatus struct {
	Name           *string    `json:"name,omitempty"`
	IconID         *uuid.UUID `json:"icon_id,omitempty"`
	PrimaryColor   *string    `json:"primary_color,omitempty"`
	SecondaryColor *string    `json:"secondary_color,omitempty"`
	Icon           *string    `json:"icon,omitempty"`
}

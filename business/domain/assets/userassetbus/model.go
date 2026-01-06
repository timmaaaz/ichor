package userassetbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type UserAsset struct {
	ID                  uuid.UUID `json:"id"`
	UserID              uuid.UUID `json:"user_id"`
	AssetID             uuid.UUID `json:"asset_id"`
	ApprovedBy          uuid.UUID `json:"approved_by"`
	ApprovalStatusID    uuid.UUID `json:"approval_status_id"`
	FulfillmentStatusID uuid.UUID `json:"fulfillment_status_id"`

	DateReceived    time.Time `json:"date_received"`
	LastMaintenance time.Time `json:"last_maintenance"`
}

type NewUserAsset struct {
	UserID              uuid.UUID `json:"user_id"`
	AssetID             uuid.UUID `json:"asset_id"`
	ApprovedBy          uuid.UUID `json:"approved_by"`
	ApprovalStatusID    uuid.UUID `json:"approval_status_id"`
	FulfillmentStatusID uuid.UUID `json:"fulfillment_status_id"`

	DateReceived    time.Time `json:"date_received"`
	LastMaintenance time.Time `json:"last_maintenance"`
}

type UpdateUserAsset struct {
	UserID              *uuid.UUID `json:"user_id,omitempty"`
	AssetID             *uuid.UUID `json:"asset_id,omitempty"`
	ApprovedBy          *uuid.UUID `json:"approved_by,omitempty"`
	ApprovalStatusID    *uuid.UUID `json:"approval_status_id,omitempty"`
	FulfillmentStatusID *uuid.UUID `json:"fulfillment_status_id,omitempty"`

	DateReceived    *time.Time `json:"date_received,omitempty"`
	LastMaintenance *time.Time `json:"last_maintenance,omitempty"`
}

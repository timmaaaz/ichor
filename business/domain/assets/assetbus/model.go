package assetbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Asset struct {
	ID               uuid.UUID `json:"id"`
	ValidAssetID     uuid.UUID `json:"valid_asset_id"`
	AssetConditionID uuid.UUID `json:"asset_condition_id"`
	LastMaintenance  time.Time `json:"last_maintenance"` // Zero value = NULL
	SerialNumber     string    `json:"serial_number"`
}

type NewAsset struct {
	ValidAssetID     uuid.UUID `json:"valid_asset_id"`
	AssetConditionID uuid.UUID `json:"asset_condition_id"`
	LastMaintenance  time.Time `json:"last_maintenance"` // Zero value = NULL
	SerialNumber     string    `json:"serial_number"`
}

type UpdateAsset struct {
	ValidAssetID     *uuid.UUID `json:"valid_asset_id,omitempty"`
	AssetConditionID *uuid.UUID `json:"asset_condition_id,omitempty"`
	LastMaintenance  *time.Time `json:"last_maintenance,omitempty"`
	SerialNumber     *string    `json:"serial_number,omitempty"`
}

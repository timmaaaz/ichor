package assettagbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type AssetTag struct {
	ID           uuid.UUID `json:"id"`
	ValidAssetID uuid.UUID `json:"valid_asset_id"`
	TagID        uuid.UUID `json:"tag_id"`
}

type NewAssetTag struct {
	ValidAssetID uuid.UUID `json:"valid_asset_id"`
	TagID        uuid.UUID `json:"tag_id"`
}

type UpdateAssetTag struct {
	ValidAssetID *uuid.UUID `json:"valid_asset_id,omitempty"`
	TagID        *uuid.UUID `json:"tag_id,omitempty"`
}

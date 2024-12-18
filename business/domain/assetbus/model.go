package assetbus

import (
	"time"

	"github.com/google/uuid"
)

type Asset struct {
	ID               uuid.UUID
	ValidAssetID     uuid.UUID
	AssetConditionID uuid.UUID
	LastMaintenance  time.Time
	SerialNumber     string
}

type NewAsset struct {
	ValidAssetID     uuid.UUID
	AssetConditionID uuid.UUID
	LastMaintenance  time.Time
	SerialNumber     string
}

type UpdateAsset struct {
	ValidAssetID     *uuid.UUID
	AssetConditionID *uuid.UUID
	LastMaintenance  *time.Time
	SerialNumber     *string
}

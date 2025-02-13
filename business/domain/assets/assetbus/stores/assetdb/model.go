package assetdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
)

type asset struct {
	ID               uuid.UUID `db:"asset_id"`
	ValidAssetID     uuid.UUID `db:"valid_asset_id"`
	LastMaintenance  time.Time `db:"last_maintenance_time"`
	SerialNumber     string    `db:"serial_number"`
	AssetConditionID uuid.UUID `db:"asset_condition_id"`
}

func toDBAsset(bus assetbus.Asset) asset {
	return asset{
		ID:               bus.ID,
		ValidAssetID:     bus.ValidAssetID,
		LastMaintenance:  bus.LastMaintenance,
		SerialNumber:     bus.SerialNumber,
		AssetConditionID: bus.AssetConditionID,
	}
}

func toBusAsset(db asset) assetbus.Asset {
	return assetbus.Asset{
		ID:               db.ID,
		ValidAssetID:     db.ValidAssetID,
		LastMaintenance:  db.LastMaintenance,
		SerialNumber:     db.SerialNumber,
		AssetConditionID: db.AssetConditionID,
	}
}

func toBusAssets(userAssets []asset) []assetbus.Asset {
	busAssets := make([]assetbus.Asset, len(userAssets))
	for i, dbUserAsset := range userAssets {
		busAssets[i] = toBusAsset(dbUserAsset)
	}
	return busAssets
}

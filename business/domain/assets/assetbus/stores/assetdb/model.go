package assetdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
)

type asset struct {
	ID               uuid.UUID    `db:"id"`
	ValidAssetID     uuid.UUID    `db:"valid_asset_id"`
	LastMaintenance  sql.NullTime `db:"last_maintenance_time"`
	SerialNumber     string       `db:"serial_number"`
	AssetConditionID uuid.UUID    `db:"asset_condition_id"`
}

func toDBAsset(bus assetbus.Asset) asset {
	var lastMaintenance sql.NullTime
	if bus.LastMaintenance.IsZero() {
		lastMaintenance = sql.NullTime{}
	} else {
		lastMaintenance = sql.NullTime{Time: bus.LastMaintenance.UTC(), Valid: true}
	}

	return asset{
		ID:               bus.ID,
		ValidAssetID:     bus.ValidAssetID,
		LastMaintenance:  lastMaintenance,
		SerialNumber:     bus.SerialNumber,
		AssetConditionID: bus.AssetConditionID,
	}
}

func toBusAsset(db asset) assetbus.Asset {
	var lastMaintenance time.Time
	if db.LastMaintenance.Valid && !db.LastMaintenance.Time.IsZero() {
		lastMaintenance = db.LastMaintenance.Time.In(time.Local)
	} else {
		lastMaintenance = time.Time{}
	}

	return assetbus.Asset{
		ID:               db.ID,
		ValidAssetID:     db.ValidAssetID,
		LastMaintenance:  lastMaintenance,
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

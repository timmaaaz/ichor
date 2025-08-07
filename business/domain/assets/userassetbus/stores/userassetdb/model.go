package userassetdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
)

// init sets the time to UTC so that the timezone matches what the db returns
// we may want to put this somewhere else
func init() {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}
	time.Local = loc
}

type userAsset struct {
	ID                  uuid.UUID `db:"id"`
	UserID              uuid.UUID `db:"user_id"`
	AssetID             uuid.UUID `db:"asset_id"`
	ApprovedBy          uuid.UUID `db:"approved_by"`
	ConditionID         uuid.UUID `db:"condition_id"`
	ApprovalStatusID    uuid.UUID `db:"approval_status_id"`
	FulfillmentStatusID uuid.UUID `db:"fulfillment_status_id"`

	DateReceived    time.Time `db:"date_received"`
	LastMaintenance time.Time `db:"last_maintenance"`
}

func toDBUserAsset(bus userassetbus.UserAsset) userAsset {
	return userAsset{
		ID:                  bus.ID,
		UserID:              bus.UserID,
		AssetID:             bus.AssetID,
		ApprovedBy:          bus.ApprovedBy,
		ApprovalStatusID:    bus.ApprovalStatusID,
		FulfillmentStatusID: bus.FulfillmentStatusID,

		DateReceived:    bus.DateReceived,
		LastMaintenance: bus.LastMaintenance,
	}
}

func toBusUserAsset(db userAsset) userassetbus.UserAsset {
	return userassetbus.UserAsset{
		ID:                  db.ID,
		UserID:              db.UserID,
		AssetID:             db.AssetID,
		ApprovedBy:          db.ApprovedBy,
		ApprovalStatusID:    db.ApprovalStatusID,
		FulfillmentStatusID: db.FulfillmentStatusID,

		DateReceived:    db.DateReceived,
		LastMaintenance: db.LastMaintenance,
	}
}

func toBusUserAssets(userAssets []userAsset) []userassetbus.UserAsset {
	busUserAssets := make([]userassetbus.UserAsset, len(userAssets))
	for i, dbUserAsset := range userAssets {
		busUserAssets[i] = toBusUserAsset(dbUserAsset)
	}
	return busUserAssets
}

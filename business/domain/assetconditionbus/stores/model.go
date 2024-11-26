package assetconditiondb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
)

type assetCondition struct {
	ID   uuid.UUID `db:"asset_condition_id"`
	Name string    `db:"name"`
}

func toDBAssetCondition(ac assetconditionbus.AssetCondition) assetCondition {
	return assetCondition{
		ID:   ac.ID,
		Name: ac.Name,
	}
}

func toBusAssetCondition(dbAC assetCondition) assetconditionbus.AssetCondition {
	return assetconditionbus.AssetCondition{
		ID:   dbAC.ID,
		Name: dbAC.Name,
	}
}

func toBusAssetConditions(dbAC []assetCondition) []assetconditionbus.AssetCondition {
	aprvlStatuses := make([]assetconditionbus.AssetCondition, len(dbAC))
	for i, ac := range dbAC {
		aprvlStatuses[i] = toBusAssetCondition(ac)
	}

	return aprvlStatuses

}

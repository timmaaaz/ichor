package assetconditiondb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
)

type assetCondition struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func toDBAssetCondition(bus assetconditionbus.AssetCondition) assetCondition {
	return assetCondition{
		ID:          bus.ID,
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func toBusAssetCondition(dbAssetCondition assetCondition) assetconditionbus.AssetCondition {
	return assetconditionbus.AssetCondition{
		ID:          dbAssetCondition.ID,
		Name:        dbAssetCondition.Name,
		Description: dbAssetCondition.Description,
	}
}

func toBusAssetConditions(dbAssetConditions []assetCondition) []assetconditionbus.AssetCondition {
	assetConditions := make([]assetconditionbus.AssetCondition, len(dbAssetConditions))
	for i, at := range dbAssetConditions {
		assetConditions[i] = toBusAssetCondition(at)
	}
	return assetConditions
}

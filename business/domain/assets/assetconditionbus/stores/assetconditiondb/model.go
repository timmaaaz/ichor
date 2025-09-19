package assetconditiondb

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
)

type assetCondition struct {
	ID          uuid.UUID      `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
}

func toDBAssetCondition(bus assetconditionbus.AssetCondition) assetCondition {
	ac := assetCondition{
		ID:   bus.ID,
		Name: bus.Name,
	}
	if bus.Description != "" {
		ac.Description = sql.NullString{
			String: bus.Description,
			Valid:  true,
		}
	}
	return ac
}

func toBusAssetCondition(dbAssetCondition assetCondition) assetconditionbus.AssetCondition {
	return assetconditionbus.AssetCondition{
		ID:          dbAssetCondition.ID,
		Name:        dbAssetCondition.Name,
		Description: dbAssetCondition.Description.String,
	}
}

func toBusAssetConditions(dbAssetConditions []assetCondition) []assetconditionbus.AssetCondition {
	assetConditions := make([]assetconditionbus.AssetCondition, len(dbAssetConditions))
	for i, at := range dbAssetConditions {
		assetConditions[i] = toBusAssetCondition(at)
	}
	return assetConditions
}

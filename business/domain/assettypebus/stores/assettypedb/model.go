package assettypedb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
)

type assetType struct {
	ID   uuid.UUID `db:"asset_type_id"`
	Name string    `db:"name"`
}

func toDBAssetType(at assettypebus.AssetType) assetType {
	return assetType{
		ID:   at.ID,
		Name: at.Name,
	}
}

func toBusAssetType(dbAT assetType) assettypebus.AssetType {
	return assettypebus.AssetType{
		ID:   dbAT.ID,
		Name: dbAT.Name,
	}
}

func toBusAssetTypes(dbAT []assetType) []assettypebus.AssetType {
	aprvlStatuses := make([]assettypebus.AssetType, len(dbAT))
	for i, as := range dbAT {
		aprvlStatuses[i] = toBusAssetType(as)
	}

	return aprvlStatuses
}

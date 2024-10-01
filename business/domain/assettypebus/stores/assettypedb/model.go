package assettypedb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
)

type assetType struct {
	ID          uuid.UUID `db:"asset_type_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func toDBAssetType(bus assettypebus.AssetType) assetType {
	return assetType{
		ID:          bus.ID,
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func toBusAssetType(dbAssetType assetType) assettypebus.AssetType {
	return assettypebus.AssetType{
		ID:          dbAssetType.ID,
		Name:        dbAssetType.Name,
		Description: dbAssetType.Description,
	}
}

func toBusAssetTypes(dbAssetTypes []assetType) []assettypebus.AssetType {
	assetTypes := make([]assettypebus.AssetType, len(dbAssetTypes))
	for i, at := range dbAssetTypes {
		assetTypes[i] = toBusAssetType(at)
	}
	return assetTypes
}

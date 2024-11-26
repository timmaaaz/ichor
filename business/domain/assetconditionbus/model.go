package assetconditionbus

import "github.com/google/uuid"

type AssetCondition struct {
	ID   uuid.UUID
	Name string
}

type NewAssetCondition struct {
	Name string
}

type UpdateAssetCondition struct {
	Name *string
}

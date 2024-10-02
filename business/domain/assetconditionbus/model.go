package assetconditionbus

import "github.com/google/uuid"

// AssetCondition represents information about an individual asset condition.
type AssetCondition struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type NewAssetCondition struct {
	Name        string
	Description string
}

type UpdateAssetCondition struct {
	Name        *string
	Description *string
}

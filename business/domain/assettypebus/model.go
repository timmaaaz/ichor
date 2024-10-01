package assettypebus

import "github.com/google/uuid"

// AssetType represents information about an individual asset type.
type AssetType struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type NewAssetType struct {
	Name        string
	Description string
}

type UpdateAssetType struct {
	Name        *string
	Description *string
}

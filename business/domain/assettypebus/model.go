package assettypebus

import (
	"github.com/google/uuid"
)

type AssetType struct {
	ID   uuid.UUID
	Name string
}

type NewAssetType struct {
	Name string
}

type UpdateAssetType struct {
	Name *string
}

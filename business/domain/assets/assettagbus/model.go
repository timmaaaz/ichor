package assettagbus

import "github.com/google/uuid"

type AssetTag struct {
	ID           uuid.UUID
	ValidAssetID uuid.UUID
	TagID        uuid.UUID
}

type NewAssetTag struct {
	ValidAssetID uuid.UUID
	TagID        uuid.UUID
}

type UpdateAssetTag struct {
	ValidAssetID *uuid.UUID
	TagID        *uuid.UUID
}

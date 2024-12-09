package assettagbus

import "github.com/google/uuid"

type AssetTag struct {
	ID      uuid.UUID
	AssetID uuid.UUID
	TagID   uuid.UUID
}

type NewAssetTag struct {
	AssetID uuid.UUID
	TagID   uuid.UUID
}

type UpdateAssetTag struct {
	AssetID *uuid.UUID
	TagID   *uuid.UUID
}

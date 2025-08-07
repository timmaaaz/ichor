package assettagdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
)

type assetTag struct {
	ID           uuid.UUID `db:"id"`
	ValidAssetID uuid.UUID `db:"valid_asset_id"`
	TagID        uuid.UUID `db:"tag_id"`
}

func toDBAssetTag(bus assettagbus.AssetTag) assetTag {
	return assetTag{
		ID:           bus.ID,
		ValidAssetID: bus.ValidAssetID,
		TagID:        bus.TagID,
	}
}

func toBusAssetTag(at assetTag) assettagbus.AssetTag {
	return assettagbus.AssetTag{
		ID:           at.ID,
		ValidAssetID: at.ValidAssetID,
		TagID:        at.TagID,
	}
}

func toBusAssetTags(ats []assetTag) []assettagbus.AssetTag {
	busTags := make([]assettagbus.AssetTag, len(ats))
	for i, at := range ats {
		busTags[i] = toBusAssetTag(at)
	}
	return busTags
}

package assettagdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
)

type assetTag struct {
	ID      uuid.UUID `db:"asset_tag_id"`
	AssetID uuid.UUID `db:"asset_id"`
	TagID   uuid.UUID `db:"tag_id"`
}

func toDBAssetTag(bus assettagbus.AssetTag) assetTag {
	return assetTag{
		ID:      bus.ID,
		AssetID: bus.AssetID,
		TagID:   bus.TagID,
	}
}

func toBusAssetTag(at assetTag) assettagbus.AssetTag {
	return assettagbus.AssetTag{
		ID:      at.ID,
		AssetID: at.AssetID,
		TagID:   at.TagID,
	}
}

func toBusAssetTags(ats []assetTag) []assettagbus.AssetTag {
	busTags := make([]assettagbus.AssetTag, len(ats))
	for i, at := range ats {
		busTags[i] = toBusAssetTag(at)
	}
	return busTags
}

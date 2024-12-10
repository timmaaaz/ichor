package assettagbus

import (
	"context"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewAssetTag(n int, assetIDs, tagIDs []uuid.UUID) []NewAssetTag {
	newAssetTag := make([]NewAssetTag, n)

	for i := 0; i < n; i++ {
		nat := NewAssetTag{
			AssetID: assetIDs[rand.Intn(len(assetIDs))],
			TagID:   tagIDs[rand.Intn(len(tagIDs))],
		}
		newAssetTag[i] = nat
	}

	return newAssetTag
}

func TestSeedAssetTag(ctx context.Context, n int, assetIDs, tagIDs []uuid.UUID, api *Business) ([]AssetTag, error) {

	newAssetTags := TestNewAssetTag(n, assetIDs, tagIDs)

	assetTags := make([]AssetTag, len(newAssetTags))
	for i, nat := range newAssetTags {
		assetTag, err := api.Create(ctx, nat)
		if err != nil {
			return nil, err
		}
		assetTags[i] = assetTag
	}

	sort.Slice(assetTags, func(i, j int) bool {
		return assetTags[i].ID.String() < assetTags[j].ID.String()
	})

	return assetTags, nil

}

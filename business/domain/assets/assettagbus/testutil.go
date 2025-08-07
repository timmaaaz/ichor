package assettagbus

import (
	"context"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewAssetTag(n int, ValidAssetIDs, tagIDs []uuid.UUID) []NewAssetTag {
	newAssetTag := make([]NewAssetTag, n)

	for i := 0; i < n; i++ {
		nat := NewAssetTag{
			ValidAssetID: ValidAssetIDs[rand.Intn(len(ValidAssetIDs))],
			TagID:        tagIDs[rand.Intn(len(tagIDs))],
		}
		newAssetTag[i] = nat
	}

	return newAssetTag
}

func TestSeedAssetTag(ctx context.Context, n int, ValidAssetIDs, tagIDs []uuid.UUID, api *Business) ([]AssetTag, error) {

	newAssetTags := TestNewAssetTag(n, ValidAssetIDs, tagIDs)

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

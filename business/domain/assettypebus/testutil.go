package assettypebus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

// TestNewAssetTypes is a helper method for testing.
func TestNewAssetTypes(n int) []NewAssetType {
	newAssetTypes := make([]NewAssetType, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		na := NewAssetType{
			Name:        fmt.Sprintf("AssetType%d", idx),
			Description: fmt.Sprintf("AssetType%d Description", idx),
		}

		newAssetTypes[i] = na
	}

	return newAssetTypes
}

// TestSeedAssetTypes is a helper method for testing.
func TestSeedAssetTypes(ctx context.Context, n int, api *Business) ([]AssetType, error) {
	newAssetTypes := TestNewAssetTypes(n)

	assetTypes := make([]AssetType, len(newAssetTypes))
	for i, na := range newAssetTypes {
		assetType, err := api.Create(ctx, na)
		if err != nil {
			return nil, fmt.Errorf("seeding asset type: idx: %d : %w", i, err)
		}

		assetTypes[i] = assetType
	}

	// sort asset types by name
	sort.Slice(assetTypes, func(i, j int) bool {
		return assetTypes[i].Name <= assetTypes[j].Name
	})

	return assetTypes, nil
}

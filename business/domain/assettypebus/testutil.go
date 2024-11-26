package assettypebus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

// TestNewAssetType is a helper method for testing
func TestNewAssetType(n int) []NewAssetType {
	newAssetType := make([]NewAssetType, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nas := NewAssetType{
			Name: fmt.Sprintf("AssetType%d", idx),
		}

		newAssetType[i] = nas
	}

	return newAssetType
}

// TestSeedAssetType is a helper method for testing.
func TestSeedAssetType(ctx context.Context, n int, api *Business) ([]AssetType, error) {
	newAssetType := TestNewAssetType(n)

	assetTypes := make([]AssetType, len(newAssetType))
	for i, nc := range newAssetType {
		as, err := api.Create(ctx, nc)
		if err != nil {
			return nil, fmt.Errorf("seeding asset type: idx: %d : %w", i, err)
		}

		assetTypes[i] = as
	}

	// sort cities by name
	sort.Slice(assetTypes, func(i, j int) bool {
		return assetTypes[i].Name <= assetTypes[j].Name
	})

	return assetTypes, nil
}

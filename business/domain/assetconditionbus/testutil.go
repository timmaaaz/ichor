package assetconditionbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

// TestNewAssetConditions is a helper method for testing.
func TestNewAssetConditions(n int) []NewAssetCondition {
	newAssetConditions := make([]NewAssetCondition, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		na := NewAssetCondition{
			Name:        fmt.Sprintf("AssetCondition%d", idx),
			Description: fmt.Sprintf("AssetCondition%d Description", idx),
		}

		newAssetConditions[i] = na
	}

	return newAssetConditions
}

// TestSeedAssetConditions is a helper method for testing.
func TestSeedAssetConditions(ctx context.Context, n int, api *Business) ([]AssetCondition, error) {
	newAssetConditions := TestNewAssetConditions(n)

	assetConditions := make([]AssetCondition, len(newAssetConditions))
	for i, na := range newAssetConditions {
		assetCondition, err := api.Create(ctx, na)
		if err != nil {
			return nil, fmt.Errorf("seeding asset condition: idx: %d : %w", i, err)
		}

		assetConditions[i] = assetCondition
	}

	// sort asset conditions by name
	sort.Slice(assetConditions, func(i, j int) bool {
		return assetConditions[i].Name <= assetConditions[j].Name
	})

	return assetConditions, nil
}

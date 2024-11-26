package assetconditionbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

// TestNewApprovalStatus is a helper method for testing.
func TestNewAssetCondition(n int) []NewAssetCondition {
	newAssetCondition := make([]NewAssetCondition, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nac := NewAssetCondition{
			Name: fmt.Sprintf("AssetCondition%d", idx),
		}

		newAssetCondition[i] = nac
	}

	return newAssetCondition
}

// TestAssetCondition is a helper method for testing.
func TestSeedAssetCondition(ctx context.Context, n int, api *Business) ([]AssetCondition, error) {
	newAssetConditions := TestNewAssetCondition(n)

	assetConditions := make([]AssetCondition, len(newAssetConditions))
	for i, nc := range newAssetConditions {
		as, err := api.Create(ctx, nc)
		if err != nil {
			return nil, fmt.Errorf("seeding city: idx: %d : %w", i, err)
		}

		assetConditions[i] = as
	}

	// sort cities by name
	sort.Slice(assetConditions, func(i, j int) bool {
		return assetConditions[i].Name <= assetConditions[j].Name
	})

	return assetConditions, nil
}

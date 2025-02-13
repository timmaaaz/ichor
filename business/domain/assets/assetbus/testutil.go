package assetbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewAssets(n int, validAssetID, conditionID uuid.UUIDs) []NewAsset {
	newAssets := make([]NewAsset, n)

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++

		na := NewAsset{
			SerialNumber:     fmt.Sprintf("SerialNumber%d", idx),
			LastMaintenance:  time.Now(),
			ValidAssetID:     validAssetID[rand.Intn(len(validAssetID))],
			AssetConditionID: conditionID[rand.Intn(len(conditionID))],
		}

		newAssets[i] = na
	}

	return newAssets
}

func TestSeedAssets(ctx context.Context, n int, validAssetID, conditionID uuid.UUIDs, api *Business) ([]Asset, error) {
	newAssets := TestNewAssets(n, validAssetID, conditionID)

	assets := make([]Asset, len(newAssets))

	for i, na := range newAssets {
		asset, err := api.Create(ctx, na)
		if err != nil {
			return nil, fmt.Errorf("seeding asset: idx: %d : %w", i, err)
		}

		assets[i] = asset
	}

	// Sort by name
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].ID.String() <= assets[j].ID.String()
	})

	return assets, nil

}

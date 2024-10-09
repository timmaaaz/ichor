package assetbus

import (
	// "context"
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assetbus/types"
)

// TestNewAssets is a helper method for testing.
func TestNewAssets(n int, typeIDs []uuid.UUID, conditionIDs []uuid.UUID, userID uuid.UUID) []NewAsset {
	newAssets := make([]NewAsset, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		estPrice, err := types.ParseMoney(fmt.Sprintf("%.2f", rand.Float64()*1000))
		if err != nil {
			panic(err)
		}

		price, err := types.ParseMoney(fmt.Sprintf("%.2f", rand.Float64()*1000))
		if err != nil {
			panic(err)
		}

		maintenanceInterval, _ := types.ParseInterval("14 days")
		lifeExpectancy, _ := types.ParseInterval("1 year")

		na := NewAsset{
			TypeID:              typeIDs[rand.Intn(len(typeIDs))],
			ConditionID:         conditionIDs[rand.Intn(len(conditionIDs))],
			Name:                fmt.Sprintf("Name%d", idx),
			EstPrice:            estPrice,
			Price:               price,
			MaintenanceInterval: maintenanceInterval,
			LifeExpectancy:      lifeExpectancy,
			ModelNumber:         fmt.Sprintf("ModelNumber%d", idx),
			IsEnabled:           true,
			CreatedBy:           userID,
		}

		newAssets[i] = na
	}

	return newAssets
}

// TestSeedAssets is a helper method for testing.
func TestSeedAssets(ctx context.Context, n int, typeIDs []uuid.UUID, conditionIDs []uuid.UUID, userID uuid.UUID, api *Business) ([]Asset, error) {
	newAssets := TestNewAssets(n, typeIDs, conditionIDs, userID)

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
		return assets[i].Name <= assets[j].Name
	})

	return assets, nil
}

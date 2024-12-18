package validassetbus

import (
	// "context"
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/validassetbus/types"
)

// TestNewAssets is a helper method for testing.
func TestNewAssets(n int, typeIDs []uuid.UUID, userID uuid.UUID) []NewValidAsset {
	newAssets := make([]NewValidAsset, n)

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

		na := NewValidAsset{
			TypeID:              typeIDs[rand.Intn(len(typeIDs))],
			Name:                fmt.Sprintf("Name%d", idx),
			EstPrice:            estPrice,
			Price:               price,
			MaintenanceInterval: maintenanceInterval,
			LifeExpectancy:      lifeExpectancy,
			SerialNumber:        fmt.Sprintf("SerialNumber%d", idx),
			ModelNumber:         fmt.Sprintf("ModelNumber%d", idx),
			IsEnabled:           true,
			CreatedBy:           userID,
		}

		newAssets[i] = na
	}

	return newAssets
}

// TestSeedAssets is a helper method for testing.
func TestSeedValidAssets(ctx context.Context, n int, typeIDs uuid.UUIDs, userID uuid.UUID, api *Business) ([]ValidAsset, error) {
	newAssets := TestNewAssets(n, typeIDs, userID)

	assets := make([]ValidAsset, len(newAssets))
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

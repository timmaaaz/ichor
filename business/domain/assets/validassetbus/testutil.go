package validassetbus

import (
	// "context"
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus/types"
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

// TestNewAssetsHistorical creates valid assets distributed across a time range for seeding.
// yearsBack specifies how many years of history to generate (1-3 years recommended for assets).
// Assets are evenly distributed across the time range.
func TestNewAssetsHistorical(n int, yearsBack int, typeIDs []uuid.UUID, userID uuid.UUID) []NewValidAsset {
	newAssets := make([]NewValidAsset, n)
	now := time.Now()

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

		// Distribute evenly across the time range
		daysAgo := (i * yearsBack * 365) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

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
			CreatedDate:         &createdDate,
		}

		newAssets[i] = na
	}

	return newAssets
}

// TestSeedValidAssetsHistorical seeds valid assets with historical date distribution.
func TestSeedValidAssetsHistorical(ctx context.Context, n int, yearsBack int, typeIDs uuid.UUIDs, userID uuid.UUID, api *Business) ([]ValidAsset, error) {
	newAssets := TestNewAssetsHistorical(n, yearsBack, typeIDs, userID)

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

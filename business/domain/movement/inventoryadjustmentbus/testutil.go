package inventoryadjustmentbus

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewInventoryAdjustment(n int, productIDs, locationIDs, adjustedByIDs, approvedByIDs uuid.UUIDs) []NewInventoryAdjustment {
	newInventoryAdjustments := make([]NewInventoryAdjustment, n)

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		newInventoryAdjustments[i] = NewInventoryAdjustment{
			ProductID:      productIDs[idx%len(productIDs)],
			LocationID:     locationIDs[idx%len(locationIDs)],
			AdjustedBy:     adjustedByIDs[idx%len(adjustedByIDs)],
			ApprovedBy:     approvedByIDs[idx%len(approvedByIDs)],
			QuantityChange: rand.Intn(100) - 50,
			ReasonCode:     "Test Reason",
			Notes:          "Test Notes",
			AdjustmentDate: time.Now().AddDate(0, 0, i%30),
		}
	}

	return newInventoryAdjustments
}

func TestSeedInventoryAdjustments(ctx context.Context, n int, productIDs, locationIDs, adjustedByIDs, approvedByIDs uuid.UUIDs, api *Business) ([]InventoryAdjustment, error) {
	newInventoryAdjustments := TestNewInventoryAdjustment(n, productIDs, locationIDs, adjustedByIDs, approvedByIDs)

	inventoryAdjustments := make([]InventoryAdjustment, len(newInventoryAdjustments))
	for i, nia := range newInventoryAdjustments {
		ia, err := api.Create(ctx, nia)
		if err != nil {
			return []InventoryAdjustment{}, err
		}
		inventoryAdjustments[i] = ia
	}

	sort.Slice(inventoryAdjustments, func(i, j int) bool {
		return inventoryAdjustments[i].InventoryAdjustmentID.String() < inventoryAdjustments[j].InventoryAdjustmentID.String()
	})

	return inventoryAdjustments, nil
}

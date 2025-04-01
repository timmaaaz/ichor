package inventoryitembus

import (
	"context"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewInventoryProducts(n int, locationIDs, productIDs uuid.UUIDs) []NewInventoryItem {
	newInventoryProducts := make([]NewInventoryItem, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newInventoryProducts[i] = NewInventoryItem{
			LocationID:            locationIDs[idx%len(locationIDs)],
			ProductID:             productIDs[idx%len(productIDs)],
			Quantity:              idx,
			ReservedQuantity:      idx,
			AllocatedQuantity:     idx,
			MinimumStock:          idx,
			MaximumStock:          idx,
			ReorderPoint:          idx,
			EconomicOrderQuantity: idx,
			SafetyStock:           idx,
			AvgDailyUsage:         idx,
		}
	}

	return newInventoryProducts
}

func TestSeedInventoryItems(ctx context.Context, n int, locationIDs, productIDs uuid.UUIDs, api *Business) ([]InventoryItem, error) {

	items := make([]InventoryItem, n)

	newItems := TestNewInventoryProducts(n, locationIDs, productIDs)

	for i, newItem := range newItems {
		item, err := api.Create(ctx, newItem)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ItemID.String() < items[j].ItemID.String()
	})

	return items, nil
}

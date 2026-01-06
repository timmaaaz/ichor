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
		// Set quantity to a base value with reserved/allocated as small fractions
		// This ensures (quantity - reserved - allocated) > 0 for allocation availability
		baseQty := 100 + idx
		newInventoryProducts[i] = NewInventoryItem{
			LocationID:            locationIDs[idx%len(locationIDs)],
			ProductID:             productIDs[idx%len(productIDs)],
			Quantity:              baseQty,
			ReservedQuantity:      0, // No reserved quantity - all available
			AllocatedQuantity:     0, // No allocated quantity - all available
			MinimumStock:          10,
			MaximumStock:          baseQty * 2,
			ReorderPoint:          20,
			EconomicOrderQuantity: 50,
			SafetyStock:           15,
			AvgDailyUsage:         5,
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
		return items[i].ID.String() < items[j].ID.String()
	})

	return items, nil
}

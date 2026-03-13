package inventoryitembus

import (
	"context"
	"sort"

	"github.com/google/uuid"
)

func TestNewInventoryProducts(n int, locationIDs, productIDs uuid.UUIDs) []NewInventoryItem {
	newInventoryProducts := make([]NewInventoryItem, n)

	// Use separate indices for location and product to guarantee unique (product_id, location_id) pairs.
	// Pairing locationIDs[i%nL] with productIDs[(i/nL)%nP] walks an nL×nP grid one cell at a time.
	nL := len(locationIDs)
	nP := len(productIDs)

	for i := range n {
		baseQty := 100 + i
		newInventoryProducts[i] = NewInventoryItem{
			LocationID:            locationIDs[i%nL],
			ProductID:             productIDs[(i/nL)%nP],
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

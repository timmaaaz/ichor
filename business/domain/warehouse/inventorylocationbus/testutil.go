package inventorylocationbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus/types"
)

func TestNewInventoryLocation(n int, warehouseIDs, zoneIDs []uuid.UUID) []NewInventoryLocation {
	newInventoryLocations := make([]NewInventoryLocation, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		newInventoryLocations[i] = NewInventoryLocation{
			WarehouseID:        warehouseIDs[idx%len(warehouseIDs)],
			ZoneID:             zoneIDs[idx%len(zoneIDs)],
			Aisle:              fmt.Sprintf("Aisle%d", idx),
			Rack:               fmt.Sprintf("Rack%d", idx),
			Shelf:              fmt.Sprintf("Shelf%d", idx),
			Bin:                fmt.Sprintf("Bin%d", idx),
			IsPickLocation:     idx%2 == 0,
			IsReserveLocation:  idx%2 == 0 && idx%5 == 0,
			MaxCapacity:        idx%100 + 10,
			CurrentUtilization: types.RoundedFloat{Value: float64(idx % 100)},
		}
	}

	return newInventoryLocations
}

func TestSeedInventoryLocations(ctx context.Context, n int, warehouseIDs, zoneIDs []uuid.UUID, api *Business) ([]InventoryLocation, error) {
	newInventoryLocations := TestNewInventoryLocation(n, warehouseIDs, zoneIDs)

	inventoryLocations := make([]InventoryLocation, len(newInventoryLocations))

	for i, newInvLoc := range newInventoryLocations {
		il, err := api.Create(ctx, newInvLoc)
		if err != nil {
			return nil, fmt.Errorf("seeding error: %v", err)
		}
		inventoryLocations[i] = il
	}

	sort.Slice(inventoryLocations, func(i, j int) bool {
		return inventoryLocations[i].LocationID.String() < inventoryLocations[j].LocationID.String()
	})

	return inventoryLocations, nil
}

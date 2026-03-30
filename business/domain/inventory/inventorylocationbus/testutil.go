package inventorylocationbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
)

func TestNewInventoryLocation(n int, warehouseIDs, zoneIDs []uuid.UUID) []NewInventoryLocation {
	newInventoryLocations := make([]NewInventoryLocation, n)

	aisles := []string{"A", "B", "C"}

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		aisle := aisles[i%len(aisles)]
		rack := fmt.Sprintf("%02d", (i/12)%5+1)
		shelf := fmt.Sprintf("%02d", (i/3)%4+1)
		bin := fmt.Sprintf("%02d", i%6+1)
		locationCode := fmt.Sprintf("%s-%s-%s-%s", aisle, rack, shelf, bin)

		// Alternate: even indices are pick locations, odd are reserve
		isPickLocation := i%2 == 0

		newInventoryLocations[i] = NewInventoryLocation{
			WarehouseID:        warehouseIDs[idx%len(warehouseIDs)],
			ZoneID:             zoneIDs[idx%len(zoneIDs)],
			Aisle:              aisle,
			Rack:               rack,
			Shelf:              shelf,
			Bin:                bin,
			LocationCode:       &locationCode,
			IsPickLocation:     isPickLocation,
			IsReserveLocation:  !isPickLocation,
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

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

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Warehouse-realistic codes: A-01-02-03 format.
		// Aisles A/B/C, racks 01-05, shelves 01-04, bins 01-06.
		// Uses idx (monotonically increasing) to guarantee uniqueness.
		aisles := []string{"A", "B", "C"}
		aisle := aisles[idx%3]
		rack := fmt.Sprintf("%02d", (idx/3)%5+1)
		shelf := fmt.Sprintf("%02d", (idx/15)%4+1)
		bin := fmt.Sprintf("%02d", (idx/60)%6+1)
		locationCode := fmt.Sprintf("%s-%s-%s-%s", aisle, rack, shelf, bin)

		newInventoryLocations[i] = NewInventoryLocation{
			WarehouseID:        warehouseIDs[idx%len(warehouseIDs)],
			ZoneID:             zoneIDs[idx%len(zoneIDs)],
			Aisle:              aisle,
			Rack:               rack,
			Shelf:              shelf,
			Bin:                bin,
			LocationCode:       &locationCode,
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

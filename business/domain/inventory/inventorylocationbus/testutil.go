package inventorylocationbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
)

func TestNewInventoryLocation(n int, warehouseIDs []uuid.UUID, zones []zonebus.Zone) []NewInventoryLocation {
	newInventoryLocations := make([]NewInventoryLocation, n)

	aisles := []string{"A", "B", "C"}

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Warehouse-realistic codes: {ZoneCode}-{Aisle}{Rack}{Shelf}{Bin} format (spec §3.7, user-approved 2026-04-16).
		// Aisles A/B/C, racks 01-05, shelves 01-04, bins 01-06.
		// Zone prefix comes from the owning zone's ZoneCode (populated by TestNewZone in phase 0c.1).
		aisle := aisles[idx%3]
		rack := fmt.Sprintf("%02d", (idx/3)%5+1)
		shelf := fmt.Sprintf("%02d", (idx/15)%4+1)
		bin := fmt.Sprintf("%02d", (idx/60)%6+1)

		zone := zones[idx%len(zones)]
		zoneCode := ""
		if zone.ZoneCode != nil {
			zoneCode = *zone.ZoneCode
		}
		locationCode := fmt.Sprintf("%s-%s%s%s%s", zoneCode, aisle, rack, shelf, bin)

		// Alternate: even indices are pick locations, odd are reserve
		isPickLocation := i%2 == 0

		newInventoryLocations[i] = NewInventoryLocation{
			WarehouseID:        warehouseIDs[idx%len(warehouseIDs)],
			ZoneID:             zone.ZoneID,
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

func TestSeedInventoryLocations(ctx context.Context, n int, warehouseIDs []uuid.UUID, zones []zonebus.Zone, api *Business) ([]InventoryLocation, error) {
	newInventoryLocations := TestNewInventoryLocation(n, warehouseIDs, zones)

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

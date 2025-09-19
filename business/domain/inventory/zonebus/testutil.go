package zonebus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewZone(n int, warehouseIDs []uuid.UUID) []NewZone {
	newZones := make([]NewZone, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newZones[i] = NewZone{
			WarehouseID: warehouseIDs[idx%len(warehouseIDs)],
			Name:        fmt.Sprintf("Name %d", idx),
			Description: fmt.Sprintf("Description %d", idx),
		}
	}

	return newZones
}

func TestSeedZone(ctx context.Context, n int, warehouseIDs []uuid.UUID, api *Business) ([]Zone, error) {
	newZones := TestNewZone(n, warehouseIDs)

	zones := make([]Zone, len(newZones))

	for i, nz := range newZones {
		zone, err := api.Create(ctx, nz)
		if err != nil {
			return nil, fmt.Errorf("seeding error: %v", err)
		}
		zones[i] = zone
	}

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].ZoneID.String() < zones[j].ZoneID.String()
	})

	return zones, nil
}

package lotlocationbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewLotLocations(n int, lotIDs uuid.UUIDs, locationIDs uuid.UUIDs) []NewLotLocation {
	newLotLocations := make([]NewLotLocation, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		newLotLocations[i] = NewLotLocation{
			LotID:      lotIDs[idx%len(lotIDs)],
			LocationID: locationIDs[idx%len(locationIDs)],
			Quantity:   rand.Intn(500) + 1,
		}
	}

	return newLotLocations
}

func TestSeedLotLocations(ctx context.Context, n int, lotIDs uuid.UUIDs, locationIDs uuid.UUIDs, api *Business) ([]LotLocation, error) {
	newLotLocations := TestNewLotLocations(n, lotIDs, locationIDs)

	lotLocations := make([]LotLocation, len(newLotLocations))

	for i, nl := range newLotLocations {
		ll, err := api.Create(ctx, nl)
		if err != nil {
			return nil, fmt.Errorf("seeding error: %v", err)
		}
		lotLocations[i] = ll
	}

	sort.Slice(lotLocations, func(i, j int) bool {
		return lotLocations[i].ID.String() < lotLocations[j].ID.String()
	})

	return lotLocations, nil
}

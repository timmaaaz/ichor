package citybus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// TestNewCities is a helper method for testing.
func TestNewCities(n int, regionIDs []uuid.UUID) []NewCity {
	newCities := make([]NewCity, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nc := NewCity{
			RegionID: regionIDs[rand.Intn(len(regionIDs))],
			Name:     fmt.Sprintf("City%d", idx),
		}

		newCities[i] = nc
	}

	return newCities
}

// TestSeedCities is a helper method for testing.
func TestSeedCities(ctx context.Context, n int, regionIDs []uuid.UUID, api *Business) ([]City, error) {
	newCities := TestNewCities(n, regionIDs)

	cities := make([]City, len(newCities))
	for i, nc := range newCities {
		city, err := api.Create(ctx, nc)
		if err != nil {
			return nil, fmt.Errorf("seeding city: idx: %d : %w", i, err)
		}

		cities[i] = city
	}

	// sort cities by name
	sort.Slice(cities, func(i, j int) bool {
		return cities[i].Name <= cities[j].Name
	})

	return cities, nil
}

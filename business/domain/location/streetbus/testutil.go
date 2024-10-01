package streetbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// TestNewStreets is a helper method for testing.
func TestNewStreets(n int, cityIDs []uuid.UUID) []NewStreet {
	newStreets := make([]NewStreet, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		ns := NewStreet{
			CityID:     cityIDs[rand.Intn(len(cityIDs))],
			Line1:      fmt.Sprintf("Street%d Line1", idx),
			Line2:      fmt.Sprintf("Street%d Line2", idx),
			PostalCode: fmt.Sprintf("PostalCode%d", idx),
		}

		newStreets[i] = ns
	}

	return newStreets
}

// TestSeedStreets is a helper method for testing.
func TestSeedStreets(ctx context.Context, n int, cityIDs []uuid.UUID, api *Business) ([]Street, error) {
	newStreets := TestNewStreets(n, cityIDs)

	streets := make([]Street, len(newStreets))
	for i, ns := range newStreets {
		street, err := api.Create(ctx, ns)
		if err != nil {
			return nil, fmt.Errorf("seeding street: idx: %d : %w", i, err)
		}

		streets[i] = street
	}

	// sort streets by line1
	sort.Slice(streets, func(i, j int) bool {
		return streets[i].Line1 <= streets[j].Line1
	})

	return streets, nil
}

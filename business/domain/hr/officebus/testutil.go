package officebus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewOffice(n int, streetIDs []uuid.UUID) []NewOffice {
	newOffices := make([]NewOffice, n)

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		idx++
		no := NewOffice{
			Name:     fmt.Sprintf("Office%d", idx),
			StreetID: streetIDs[rand.Intn(len(streetIDs))],
		}

		newOffices[i] = no
	}

	return newOffices
}

func TestSeedOffices(ctx context.Context, n int, streetIDs []uuid.UUID, api *Business) ([]Office, error) {
	newOffices := TestNewOffice(n, streetIDs)

	offices := make([]Office, len(newOffices))
	for i, no := range newOffices {
		office, err := api.Create(ctx, no)
		if err != nil {
			return nil, fmt.Errorf("seeding office: idx: %d : %w", i, err)
		}

		offices[i] = office
	}

	sort.Slice(offices, func(i, j int) bool {
		return offices[i].Name <= offices[j].Name
	})

	return offices, nil
}

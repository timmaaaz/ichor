package brandbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewBrands(n int, contacts []uuid.UUID) []NewBrand {
	newBrands := make([]NewBrand, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++
		nb := NewBrand{
			Name:           fmt.Sprintf("Brand%d", idx),
			ContactInfosID: contacts[rand.Intn(len(contacts))],
		}
		newBrands[i] = nb
	}

	return newBrands
}

func TestSeedBrands(ctx context.Context, n int, contacts []uuid.UUID, api *Business) ([]Brand, error) {
	newBrands := TestNewBrands(n, contacts)

	brands := make([]Brand, len(newBrands))

	for i, nb := range newBrands {
		brand, err := api.Create(ctx, nb)
		if err != nil {
			return nil, err
		}
		brands[i] = brand
	}

	sort.Slice(brands, func(i, j int) bool {
		return brands[i].Name < brands[j].Name
	})

	return brands, nil
}

package productcategorybus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

func TestNewProductCategories(n int) []NewProductCategory {
	newProductCategories := make([]NewProductCategory, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		npc := NewProductCategory{
			Name:        "ProductCategory" + fmt.Sprintf("%d", idx),
			Description: "Description" + fmt.Sprintf("%d", idx),
		}
		newProductCategories[i] = npc
	}

	return newProductCategories
}

func TestSeedProductCategories(ctx context.Context, n int, api *Business) ([]ProductCategory, error) {
	newCategories := TestNewProductCategories(n)

	categories := make([]ProductCategory, len(newCategories))

	for i, nc := range newCategories {
		c, err := api.Create(ctx, nc)
		if err != nil {
			return nil, fmt.Errorf("seeding product category: idx: %d : %w", i, err)
		}

		categories[i] = c
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name <= categories[j].Name
	})

	return categories, nil

}

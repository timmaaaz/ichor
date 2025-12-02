package productbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewProducts(n int, brandIDs, productCategoryIDs uuid.UUIDs) []NewProduct {
	newProducts := make([]NewProduct, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		np := NewProduct{
			Name:                 fmt.Sprintf("Product%d", idx),
			BrandID:              brandIDs[rand.Intn(len(brandIDs))],
			ProductCategoryID:    productCategoryIDs[rand.Intn(len(productCategoryIDs))],
			Description:          "Description" + fmt.Sprintf("%d", idx),
			SKU:                  fmt.Sprintf("SKU%d", idx),
			ModelNumber:          fmt.Sprintf("ModelNumber%d", idx),
			UpcCode:              fmt.Sprintf("UpcCode%d", idx),
			Status:               fmt.Sprintf("Status%d", idx),
			IsActive:             idx%2 == 0,
			IsPerishable:         idx%2 == 0 && idx%5 == 0,
			HandlingInstructions: fmt.Sprintf("Handling instructions %d", idx),
			UnitsPerCase:         idx * 5,
		}
		newProducts[i] = np
	}

	return newProducts
}

func TestSeedProducts(ctx context.Context, n int, brandIDs, productCategoryIDs uuid.UUIDs, api *Business) ([]Product, error) {
	newProducts := TestNewProducts(n, brandIDs, productCategoryIDs)

	products := make([]Product, len(newProducts))

	for i, np := range newProducts {
		product, err := api.Create(ctx, np)
		if err != nil {
			return nil, err
		}
		products[i] = product
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].Name < products[j].Name
	})

	return products, nil
}

// TestNewProductsHistorical creates products distributed across a time range for seeding.
// daysBack specifies how many days of history to generate (180-365 days recommended for products).
// Products are evenly distributed across the time range.
func TestNewProductsHistorical(n int, daysBack int, brandIDs, productCategoryIDs uuid.UUIDs) []NewProduct {
	newProducts := make([]NewProduct, n)
	now := time.Now()

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Distribute evenly across the time range
		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

		np := NewProduct{
			Name:                 fmt.Sprintf("Product%d", idx),
			BrandID:              brandIDs[rand.Intn(len(brandIDs))],
			ProductCategoryID:    productCategoryIDs[rand.Intn(len(productCategoryIDs))],
			Description:          "Description" + fmt.Sprintf("%d", idx),
			SKU:                  fmt.Sprintf("SKU%d", idx),
			ModelNumber:          fmt.Sprintf("ModelNumber%d", idx),
			UpcCode:              fmt.Sprintf("UpcCode%d", idx),
			Status:               fmt.Sprintf("Status%d", idx),
			IsActive:             idx%2 == 0,
			IsPerishable:         idx%2 == 0 && idx%5 == 0,
			HandlingInstructions: fmt.Sprintf("Handling instructions %d", idx),
			UnitsPerCase:         idx * 5,
			CreatedDate:          &createdDate,
		}
		newProducts[i] = np
	}

	return newProducts
}

// TestSeedProductsHistorical seeds products with historical date distribution.
func TestSeedProductsHistorical(ctx context.Context, n int, daysBack int, brandIDs, productCategoryIDs uuid.UUIDs, api *Business) ([]Product, error) {
	newProducts := TestNewProductsHistorical(n, daysBack, brandIDs, productCategoryIDs)

	products := make([]Product, len(newProducts))

	for i, np := range newProducts {
		product, err := api.Create(ctx, np)
		if err != nil {
			return nil, err
		}
		products[i] = product
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].Name < products[j].Name
	})

	return products, nil
}

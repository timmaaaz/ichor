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

	productNames := []string{
		"Industrial Bearing 6205",
		"Nitrile Gloves Box/100",
		"Hydraulic Filter HF-302",
		"LED Panel Light 60W",
		"Stainless Steel Bolt M10",
		"Thermal Paste Tube 5g",
		"Safety Goggles Clear",
		"Rubber Gasket Set",
		"Wire Spool CAT6 100m",
		"Epoxy Adhesive 2-Part",
	}

	trackingTypes := []string{"none", "lot", "serial"}

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		trackingType := trackingTypes[i%len(trackingTypes)]

		np := NewProduct{
			Name:                 productNames[i%len(productNames)],
			BrandID:              brandIDs[rand.Intn(len(brandIDs))],
			ProductCategoryID:    productCategoryIDs[rand.Intn(len(productCategoryIDs))],
			Description:          fmt.Sprintf("High-quality %s for warehouse operations", productNames[i%len(productNames)]),
			SKU:                  fmt.Sprintf("SKU-%04d", i+1),
			ModelNumber:          fmt.Sprintf("MDL-%04d", i+1),
			UpcCode:              fmt.Sprintf("0123456789%02d", i%100),
			Status:               "active",
			IsActive:             idx%2 == 0,
			IsPerishable:         trackingType == "lot",
			HandlingInstructions: fmt.Sprintf("Handling instructions %d", idx),
			UnitsPerCase:         idx * 5,
			TrackingType:         trackingType,
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
		if products[i].Name != products[j].Name {
			return products[i].Name < products[j].Name
		}
		return products[i].SKU < products[j].SKU
	})

	return products, nil
}

// TestNewProductsHistorical creates products distributed across a time range for seeding.
// daysBack specifies how many days of history to generate (180-365 days recommended for products).
// Products are evenly distributed across the time range.
func TestNewProductsHistorical(n int, daysBack int, brandIDs, productCategoryIDs uuid.UUIDs) []NewProduct {
	newProducts := make([]NewProduct, n)
	now := time.Now()

	productNames := []string{
		"Industrial Bearing 6205",
		"Nitrile Gloves Box/100",
		"Hydraulic Filter HF-302",
		"LED Panel Light 60W",
		"Stainless Steel Bolt M10",
		"Thermal Paste Tube 5g",
		"Safety Goggles Clear",
		"Rubber Gasket Set",
		"Wire Spool CAT6 100m",
		"Epoxy Adhesive 2-Part",
	}

	trackingTypes := []string{"none", "lot", "serial"}

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

		trackingType := trackingTypes[i%len(trackingTypes)]

		np := NewProduct{
			Name:                 productNames[i%len(productNames)],
			BrandID:              brandIDs[rand.Intn(len(brandIDs))],
			ProductCategoryID:    productCategoryIDs[rand.Intn(len(productCategoryIDs))],
			Description:          fmt.Sprintf("High-quality %s for warehouse operations", productNames[i%len(productNames)]),
			SKU:                  fmt.Sprintf("SKU-%04d", i+1),
			ModelNumber:          fmt.Sprintf("MDL-%04d", i+1),
			UpcCode:              fmt.Sprintf("0123456789%02d", i%100),
			Status:               "active",
			IsActive:             idx%2 == 0,
			IsPerishable:         trackingType == "lot",
			HandlingInstructions: fmt.Sprintf("Handling instructions %d", idx),
			UnitsPerCase:         idx * 5,
			TrackingType:         trackingType,
			CreatedDate:          &createdDate,
		}
		newProducts[i] = np
	}

	return newProducts
}

// TestSeedProductsHistoricalWithDistribution generates n products with the
// caller-provided tracking-type distribution. distribution must have len == n;
// each entry must be one of "none", "lot", "serial". Pass nil to fall back to
// the modulo-based default ({"none","lot","serial"}[i%len(defaultTypes)]).
func TestSeedProductsHistoricalWithDistribution(
	ctx context.Context,
	n int,
	daysBack int,
	distribution []string,
	brandIDs, productCategoryIDs uuid.UUIDs,
	api *Business,
) ([]Product, error) {
	if distribution != nil && len(distribution) != n {
		return nil, fmt.Errorf("distribution length %d != n %d", len(distribution), n)
	}

	newProducts := TestNewProductsHistorical(n, daysBack, brandIDs, productCategoryIDs)
	if distribution != nil {
		for i := range newProducts {
			newProducts[i].TrackingType = distribution[i]
			newProducts[i].IsPerishable = distribution[i] == "lot"
		}
	}

	products := make([]Product, 0, n)
	for i, np := range newProducts {
		p, err := api.Create(ctx, np)
		if err != nil {
			return nil, fmt.Errorf("create product %d: %w", i, err)
		}
		products = append(products, p)
	}

	sort.Slice(products, func(i, j int) bool {
		if products[i].Name != products[j].Name {
			return products[i].Name < products[j].Name
		}
		return products[i].SKU < products[j].SKU
	})

	return products, nil
}

// TestSeedProductsHistorical seeds products with historical date distribution.
func TestSeedProductsHistorical(ctx context.Context, n int, daysBack int, brandIDs, productCategoryIDs uuid.UUIDs, api *Business) ([]Product, error) {
	return TestSeedProductsHistoricalWithDistribution(ctx, n, daysBack, nil, brandIDs, productCategoryIDs, api)
}

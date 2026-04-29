package productbus

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
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

	idx := 0
	for i := 0; i < n; i++ {
		idx++

		trackingType := trackingTypes[i%len(trackingTypes)]

		np := NewProduct{
			Name:                 productNames[i%len(productNames)],
			BrandID:              brandIDs[i%len(brandIDs)],
			ProductCategoryID:    productCategoryIDs[i%len(productCategoryIDs)],
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
		p := newProductToSeedProduct(np)
		if err := api.SeedCreate(ctx, p); err != nil {
			return nil, err
		}
		// Round-trip via QueryByID so callers receive whatever the DB
		// normalised (UTC timestamps, defaulted TrackingType) — matches
		// the shape they previously got back from api.Create.
		stored, err := api.QueryByID(ctx, p.ProductID)
		if err != nil {
			return nil, fmt.Errorf("querying seeded product %s: %w", np.SKU, err)
		}
		products[i] = stored
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

	idx := 0
	for i := 0; i < n; i++ {
		idx++

		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

		trackingType := trackingTypes[i%len(trackingTypes)]

		np := NewProduct{
			Name:                 productNames[i%len(productNames)],
			BrandID:              brandIDs[i%len(brandIDs)],
			ProductCategoryID:    productCategoryIDs[i%len(productCategoryIDs)],
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
// the modulo-based default that TestNewProductsHistorical applies via
// {"none","lot","serial"}[i%len(trackingTypes)].
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
		p := newProductToSeedProduct(np)
		if err := api.SeedCreate(ctx, p); err != nil {
			return nil, fmt.Errorf("create product %d: %w", i, err)
		}
		stored, err := api.QueryByID(ctx, p.ProductID)
		if err != nil {
			return nil, fmt.Errorf("querying seeded product %d: %w", i, err)
		}
		products = append(products, stored)
	}

	sort.Slice(products, func(i, j int) bool {
		if products[i].Name != products[j].Name {
			return products[i].Name < products[j].Name
		}
		return products[i].SKU < products[j].SKU
	})

	return products, nil
}

// newProductToSeedProduct copies a NewProduct into a Product struct,
// deriving a deterministic ProductID from the SKU. Historical
// CreatedDate (set by TestNewProductsHistorical) is preserved verbatim
// so timestamp-sensitive tests still see the same dates they did under
// the api.Create path.
func newProductToSeedProduct(np NewProduct) Product {
	var createdDate time.Time
	if np.CreatedDate != nil {
		createdDate = *np.CreatedDate
	}
	return Product{
		ProductID:            seedid.Stable("product:" + np.SKU),
		SKU:                  np.SKU,
		BrandID:              np.BrandID,
		ProductCategoryID:    np.ProductCategoryID,
		Name:                 np.Name,
		Description:          np.Description,
		ModelNumber:          np.ModelNumber,
		UpcCode:              np.UpcCode,
		Status:               np.Status,
		IsActive:             np.IsActive,
		IsPerishable:         np.IsPerishable,
		HandlingInstructions: np.HandlingInstructions,
		UnitsPerCase:         np.UnitsPerCase,
		TrackingType:         np.TrackingType,
		InventoryType:        np.InventoryType,
		CreatedDate:          createdDate, // zero → defaulted by SeedCreate
	}
}

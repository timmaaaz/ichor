package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/products/physicalattributebus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
)

// ProductsSeed holds the results of seeding product data.
type ProductsSeed struct {
	Products []productbus.Product
}

func seedProducts(ctx context.Context, busDomain BusDomain, geoHR GeographyHRSeed, foundation FoundationSeed) (ProductsSeed, error) {
	contactIDs := make(uuid.UUIDs, len(geoHR.ContactInfos))
	for i, c := range geoHR.ContactInfos {
		contactIDs[i] = c.ID
	}

	brand, err := brandbus.TestSeedBrands(ctx, 5, contactIDs, busDomain.Brand)
	if err != nil {
		return ProductsSeed{}, fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brand))
	for i, b := range brand {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 10, busDomain.ProductCategory)
	if err != nil {
		return ProductsSeed{}, fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	const productCount = 40

	// 28×"none" + 8×"lot" + 4×"serial" tracking-type distribution for the
	// 40-product seed. Order is intentional and stable across reseeds.
	distribution := make([]string, 0, productCount)
	for i := 0; i < 28; i++ {
		distribution = append(distribution, "none")
	}
	for i := 0; i < 8; i++ {
		distribution = append(distribution, "lot")
	}
	for i := 0; i < 4; i++ {
		distribution = append(distribution, "serial")
	}

	products, err := productbus.TestSeedProductsHistoricalWithDistribution(
		ctx, productCount, 180, distribution, brandIDs, productCategoryIDs, busDomain.Product,
	)
	if err != nil {
		return ProductsSeed{}, fmt.Errorf("seeding product : %w", err)
	}
	productIDs := make([]uuid.UUID, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ProductID)
	}

	// All product costs use USD - single base currency for consistency
	_, err = productcostbus.TestSeedProductCosts(ctx, productCount, productIDs, uuid.UUIDs{foundation.USDCurrencyID}, busDomain.ProductCost)
	if err != nil {
		return ProductsSeed{}, fmt.Errorf("seeding product cost : %w", err)
	}

	_, err = physicalattributebus.TestSeedPhysicalAttributes(ctx, productCount, productIDs, busDomain.PhysicalAttribute)
	if err != nil {
		return ProductsSeed{}, fmt.Errorf("seeding physical attribute : %w", err)
	}

	_, err = metricsbus.TestSeedMetrics(ctx, 2*productCount, productIDs, busDomain.Metrics)
	if err != nil {
		return ProductsSeed{}, fmt.Errorf("seeding metrics : %w", err)
	}

	// Cost history also uses USD for consistency
	_, err = costhistorybus.TestSeedCostHistoriesHistorical(ctx, 2*productCount, 180, productIDs, uuid.UUIDs{foundation.USDCurrencyID}, busDomain.CostHistory)
	if err != nil {
		return ProductsSeed{}, fmt.Errorf("seeding cost history : %w", err)
	}

	return ProductsSeed{
		Products: products,
	}, nil
}

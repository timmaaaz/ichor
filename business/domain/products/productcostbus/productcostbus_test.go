package productcostbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ProductCost(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ProductCost")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}
	// -------------------------------------------------------------------------
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	count := 5

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	ctys, err := citybus.TestSeedCities(ctx, count, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, count, ctyIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

	// Query timezones from seed data
	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 5, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brand, err := brandbus.TestSeedBrands(ctx, 5, contactIDs, busDomain.Brand)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brand))
	for i, b := range brand {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 10, busDomain.ProductCategory)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))

	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 20, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product : %w", err)
	}

	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	currencies, err := currencybus.TestSeedCurrencies(ctx, 5, busDomain.Currency)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding currencies: %w", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	productCosts, err := productcostbus.TestSeedProductCosts(ctx, 20, productIDs, currencyIDs, busDomain.ProductCost)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product cost : %w", err)
	}

	return unitest.SeedData{
		Admins:            []unitest.User{{User: admins[0]}},
		Brands:            brand,
		ProductCategories: productCategories,
		Products:          products,
		ProductCosts:      productCosts,
		ContactInfos:      contactInfos,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []productcostbus.ProductCost{
				sd.ProductCosts[0],
				sd.ProductCosts[1],
				sd.ProductCosts[2],
				sd.ProductCosts[3],
				sd.ProductCosts[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.ProductCost.Query(ctx, productcostbus.QueryFilter{}, productcostbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]productcostbus.ProductCost)
				if !exists {
					return fmt.Sprintf("got is not a slice of product costs: %v", got)
				}

				expResp := exp.([]productcostbus.ProductCost)

				return cmp.Diff(gotResp, expResp)

			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: productcostbus.ProductCost{
				ProductID:         sd.Products[2].ProductID,
				PurchaseCost:      sd.ProductCosts[2].PurchaseCost,
				SellingPrice:      sd.ProductCosts[2].SellingPrice,
				CurrencyID:          sd.ProductCosts[2].CurrencyID,
				MSRP:              sd.ProductCosts[2].MSRP,
				MarkupPercentage:  sd.ProductCosts[2].MarkupPercentage,
				LandedCost:        sd.ProductCosts[2].LandedCost,
				CarryingCost:      sd.ProductCosts[2].CarryingCost,
				ABCClassification: sd.ProductCosts[2].ABCClassification,
				DepreciationValue: sd.ProductCosts[2].DepreciationValue,
				InsuranceValue:    sd.ProductCosts[2].InsuranceValue,
				EffectiveDate:     sd.ProductCosts[2].EffectiveDate,
			},
			ExcFunc: func(ctx context.Context) any {
				newProductCost := productcostbus.NewProductCost{
					ProductID:         sd.Products[2].ProductID,
					PurchaseCost:      sd.ProductCosts[2].PurchaseCost,
					SellingPrice:      sd.ProductCosts[2].SellingPrice,
					CurrencyID:          sd.ProductCosts[2].CurrencyID,
					MSRP:              sd.ProductCosts[2].MSRP,
					MarkupPercentage:  sd.ProductCosts[2].MarkupPercentage,
					LandedCost:        sd.ProductCosts[2].LandedCost,
					CarryingCost:      sd.ProductCosts[2].CarryingCost,
					ABCClassification: sd.ProductCosts[2].ABCClassification,
					DepreciationValue: sd.ProductCosts[2].DepreciationValue,
					InsuranceValue:    sd.ProductCosts[2].InsuranceValue,
					EffectiveDate:     sd.ProductCosts[2].EffectiveDate,
				}

				ci, err := busDomain.ProductCost.Create(ctx, newProductCost)
				if err != nil {
					return err
				}

				return ci
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(productcostbus.ProductCost)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(productcostbus.ProductCost)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: productcostbus.ProductCost{
				ProductID:         sd.Products[2].ProductID,
				PurchaseCost:      sd.ProductCosts[2].PurchaseCost,
				SellingPrice:      sd.ProductCosts[2].SellingPrice,
				CurrencyID:          sd.ProductCosts[2].CurrencyID,
				MSRP:              sd.ProductCosts[2].MSRP,
				MarkupPercentage:  sd.ProductCosts[2].MarkupPercentage,
				LandedCost:        sd.ProductCosts[2].LandedCost,
				CarryingCost:      sd.ProductCosts[2].CarryingCost,
				ABCClassification: sd.ProductCosts[2].ABCClassification,
				DepreciationValue: sd.ProductCosts[2].DepreciationValue,
				InsuranceValue:    sd.ProductCosts[2].InsuranceValue,
				EffectiveDate:     sd.ProductCosts[2].EffectiveDate,
				CreatedDate:       sd.ProductCosts[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				uc := productcostbus.UpdateProductCost{
					ProductID:         &sd.Products[2].ProductID,
					PurchaseCost:      &sd.ProductCosts[2].PurchaseCost,
					SellingPrice:      &sd.ProductCosts[2].SellingPrice,
					CurrencyID:          &sd.ProductCosts[2].CurrencyID,
					MSRP:              &sd.ProductCosts[2].MSRP,
					MarkupPercentage:  &sd.ProductCosts[2].MarkupPercentage,
					LandedCost:        &sd.ProductCosts[2].LandedCost,
					CarryingCost:      &sd.ProductCosts[2].CarryingCost,
					ABCClassification: &sd.ProductCosts[2].ABCClassification,
					DepreciationValue: &sd.ProductCosts[2].DepreciationValue,
					InsuranceValue:    &sd.ProductCosts[2].InsuranceValue,
					EffectiveDate:     &sd.ProductCosts[2].EffectiveDate,
				}

				got, err := busDomain.ProductCost.Update(ctx, sd.ProductCosts[0], uc)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(productcostbus.ProductCost)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(productcostbus.ProductCost)

				expResp.ID = gotResp.ID
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.ProductCost.Delete(ctx, sd.ProductCosts[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

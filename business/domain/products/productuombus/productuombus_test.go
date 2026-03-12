package productuombus_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ProductUOM(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ProductUOM")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user: %w", err)
	}

	count := 5

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions: %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, count, regionIDs, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities: %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, count, cityIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets: %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying timezones: %w", err)
	}

	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, count, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}

	contactIDs := make(uuid.UUIDs, 0, len(contactInfos))
	for _, c := range contactInfos {
		contactIDs = append(contactIDs, c.ID)
	}

	brands, err := brandbus.TestSeedBrands(ctx, count, contactIDs, busDomain.Brand)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding brands: %w", err)
	}

	brandIDs := make(uuid.UUIDs, 0, len(brands))
	for _, b := range brands {
		brandIDs = append(brandIDs, b.BrandID)
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, count, busDomain.ProductCategory)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, 0, len(productCategories))
	for _, pc := range productCategories {
		productCategoryIDs = append(productCategoryIDs, pc.ProductCategoryID)
	}

	products, err := productbus.TestSeedProducts(ctx, 10, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding products: %w", err)
	}

	productIDs := make([]uuid.UUID, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ProductID)
	}

	_, err = productuombus.TestSeedProductUOMs(ctx, 10, productIDs, busDomain.ProductUOM)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding product uoms: %w", err)
	}

	// Re-query to get DB-canonical values (NUMERIC(10,4) rounds float64 values).
	uoms, err := busDomain.ProductUOM.Query(ctx, productuombus.QueryFilter{}, productuombus.DefaultOrderBy, page.MustParse("1", "10"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("re-querying product uoms: %w", err)
	}

	return unitest.SeedData{
		Admins:      []unitest.User{{User: admins[0]}},
		Products:    products,
		ProductUOMs: uoms,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "query",
			ExpResp: sd.ProductUOMs[:5],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.ProductUOM.Query(ctx, productuombus.QueryFilter{}, productuombus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]productuombus.ProductUOM)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp.([]productuombus.ProductUOM), gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	idx := rand.Intn(len(sd.Products))
	product := sd.Products[idx]

	return []unitest.Table{
		{
			Name: "create",
			ExpResp: productuombus.ProductUOM{
				ProductID:        product.ProductID,
				Name:             "TestUOM",
				Abbreviation:     "TU",
				ConversionFactor: 1.0,
				IsBase:           false,
				IsApproximate:    false,
				Notes:            "test note",
			},
			ExcFunc: func(ctx context.Context) any {
				nu := productuombus.NewProductUOM{
					ProductID:        product.ProductID,
					Name:             "TestUOM",
					Abbreviation:     "TU",
					ConversionFactor: 1.0,
					IsBase:           false,
					IsApproximate:    false,
					Notes:            "test note",
				}
				uom, err := busDomain.ProductUOM.Create(ctx, nu)
				if err != nil {
					return err
				}
				return uom
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(productuombus.ProductUOM)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(productuombus.ProductUOM)
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	uom := sd.ProductUOMs[0]
	newName := "UpdatedUOM"

	return []unitest.Table{
		{
			Name: "update",
			ExpResp: productuombus.ProductUOM{
				ID:               uom.ID,
				ProductID:        uom.ProductID,
				Name:             newName,
				Abbreviation:     uom.Abbreviation,
				ConversionFactor: uom.ConversionFactor,
				IsBase:           uom.IsBase,
				IsApproximate:    uom.IsApproximate,
				Notes:            uom.Notes,
				CreatedDate:      uom.CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				uu := productuombus.UpdateProductUOM{
					Name: &newName,
				}
				updated, err := busDomain.ProductUOM.Update(ctx, uom, uu)
				if err != nil {
					return err
				}
				return updated
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(productuombus.ProductUOM)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(productuombus.ProductUOM)
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	uom := sd.ProductUOMs[len(sd.ProductUOMs)-1]

	return []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.ProductUOM.Delete(ctx, uom); err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(exp, got)
			},
		},
	}
}

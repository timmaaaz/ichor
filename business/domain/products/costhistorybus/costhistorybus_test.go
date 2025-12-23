package costhistorybus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_CostHistory(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CostHistory")

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
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	// ADDRESSES
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

	costHistory, err := costhistorybus.TestSeedCostHistories(ctx, 40, productIDs, busDomain.CostHistory)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cost history : %w", err)
	}

	return unitest.SeedData{
		Admins:            []unitest.User{{User: admins[0]}},
		ContactInfos:      contactInfos,
		Brands:            brand,
		Products:          products,
		ProductCategories: productCategories,
		CostHistory:       costHistory,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []costhistorybus.CostHistory{
				sd.CostHistory[0],
				sd.CostHistory[1],
				sd.CostHistory[2],
				sd.CostHistory[3],
				sd.CostHistory[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.CostHistory.Query(ctx, costhistorybus.QueryFilter{}, costhistorybus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]costhistorybus.CostHistory)
				if !exists {
					return fmt.Sprintf("got is not a slice of cost histories: %v", got)
				}

				expResp := exp.([]costhistorybus.CostHistory)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: costhistorybus.CostHistory{
				ProductID:     sd.Products[0].ProductID,
				CostType:      sd.CostHistory[2].CostType,
				Amount:        sd.CostHistory[2].Amount,
				Currency:      sd.CostHistory[2].Currency,
				EffectiveDate: sd.CostHistory[2].EffectiveDate,
				EndDate:       sd.CostHistory[2].EndDate,
			},
			ExcFunc: func(ctx context.Context) any {

				newCostHistory := costhistorybus.NewCostHistory{
					ProductID:     sd.Products[0].ProductID,
					CostType:      sd.CostHistory[2].CostType,
					Amount:        sd.CostHistory[2].Amount,
					Currency:      sd.CostHistory[2].Currency,
					EffectiveDate: sd.CostHistory[2].EffectiveDate,
					EndDate:       sd.CostHistory[2].EndDate,
				}

				costHistory, err := busDomain.CostHistory.Create(ctx, newCostHistory)
				if err != nil {
					return err
				}
				return costHistory
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(costhistorybus.CostHistory)
				if !exists {
					return fmt.Sprintf("got is not a cost history: %v", got)
				}

				expResp := exp.(costhistorybus.CostHistory)
				expResp.CostHistoryID = gotResp.CostHistoryID
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
			ExpResp: costhistorybus.CostHistory{
				ProductID:     sd.Products[0].ProductID,
				CostType:      sd.CostHistory[1].CostType,
				Amount:        sd.CostHistory[2].Amount,
				Currency:      sd.CostHistory[3].Currency,
				EffectiveDate: sd.CostHistory[4].EffectiveDate,
				EndDate:       sd.CostHistory[5].EndDate,
				CreatedDate:   sd.CostHistory[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				uch := costhistorybus.UpdateCostHistory{
					ProductID:     &sd.Products[0].ProductID,
					CostType:      &sd.CostHistory[1].CostType,
					Amount:        &sd.CostHistory[2].Amount,
					Currency:      &sd.CostHistory[3].Currency,
					EffectiveDate: &sd.CostHistory[4].EffectiveDate,
					EndDate:       &sd.CostHistory[5].EndDate,
				}

				costHistory, err := busDomain.CostHistory.Update(ctx, sd.CostHistory[0], uch)
				if err != nil {
					return err
				}

				return costHistory
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(costhistorybus.CostHistory)
				if !exists {
					return fmt.Sprintf("got is not a cost history: %v", got)
				}

				expResp := exp.(costhistorybus.CostHistory)
				expResp.CostHistoryID = gotResp.CostHistoryID
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.CostHistory.Delete(ctx, sd.CostHistory[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

package lottrackingsbus_test

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
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_LotTrackings(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_LotTrackings")

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

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 25, contactIDs, busDomain.Supplier)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding suppliers : %w", err)
	}

	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 10, productIDs, supplierIDs, busDomain.SupplierProduct)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding supplier product : %w", err)
	}

	supplierProductIDs := make(uuid.UUIDs, len(supplierProducts))
	for i, sp := range supplierProducts {
		supplierProductIDs[i] = sp.SupplierProductID
	}

	lotTrackings, err := lottrackingsbus.TestSeedLotTrackings(ctx, 15, supplierProductIDs, busDomain.LotTrackings)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding lot tracking : %w", err)
	}

	return unitest.SeedData{
		Admins:            []unitest.User{{User: admins[0]}},
		Brands:            brand,
		ProductCategories: productCategories,
		Products:          products,
		Suppliers:         suppliers,
		SupplierProducts:  supplierProducts,
		LotTrackings:      lotTrackings,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []lottrackingsbus.LotTrackings{
				sd.LotTrackings[0],
				sd.LotTrackings[1],
				sd.LotTrackings[2],
				sd.LotTrackings[3],
				sd.LotTrackings[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.LotTrackings.Query(ctx, lottrackingsbus.QueryFilter{}, lottrackingsbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]lottrackingsbus.LotTrackings)
				if !exists {
					return fmt.Sprintf("got is not a slice of lot trackings: %v", got)
				}

				expResp := exp.([]lottrackingsbus.LotTrackings)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	md := lottrackingsbus.RandomDate()
	ed := lottrackingsbus.RandomDate()
	rd := lottrackingsbus.RandomDate()

	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: lottrackingsbus.LotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md,
				ExpirationDate:    ed,
				RecievedDate:      rd,
				Quantity:          15,
				QualityStatus:     "good",
			},
			ExcFunc: func(ctx context.Context) any {
				newLotTrackings := lottrackingsbus.NewLotTrackings{
					SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
					LotNumber:         "LotNumber",
					ManufactureDate:   md,
					ExpirationDate:    ed,
					RecievedDate:      rd,
					Quantity:          15,
					QualityStatus:     "good",
				}

				s, err := busDomain.LotTrackings.Create(ctx, newLotTrackings)
				if err != nil {
					return err
				}

				return s

			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(lottrackingsbus.LotTrackings)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(lottrackingsbus.LotTrackings)

				expResp.LotID = gotResp.LotID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	md := lottrackingsbus.RandomDate()
	ed := lottrackingsbus.RandomDate()
	rd := lottrackingsbus.RandomDate()

	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: lottrackingsbus.LotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "UpdatedLotNumber",
				ManufactureDate:   md,
				ExpirationDate:    ed,
				RecievedDate:      rd,
				Quantity:          20,
				QualityStatus:     "good",
			},
			ExcFunc: func(ctx context.Context) any {
				updatedLotTrackings := lottrackingsbus.UpdateLotTrackings{
					SupplierProductID: &sd.SupplierProducts[0].SupplierProductID,
					LotNumber:         dbtest.StringPointer("UpdatedLotNumber"),
					ManufactureDate:   &md,
					ExpirationDate:    &ed,
					RecievedDate:      &rd,
					Quantity:          dbtest.IntPointer(20),
					QualityStatus:     dbtest.StringPointer("good"),
				}

				lt, err := busDomain.LotTrackings.Update(ctx, sd.LotTrackings[3], updatedLotTrackings)
				if err != nil {
					return err
				}

				return lt
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(lottrackingsbus.LotTrackings)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(lottrackingsbus.LotTrackings)

				expResp.LotID = gotResp.LotID
				expResp.CreatedDate = gotResp.CreatedDate
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
				err := busDomain.LotTrackings.Delete(ctx, sd.LotTrackings[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

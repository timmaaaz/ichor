package supplierproductbus_test

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
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus/types"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_SupplierProduct(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_SupplierProduct")

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

	ContactInfosIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		ContactInfosIDs[i] = c.ID
	}

	brand, err := brandbus.TestSeedBrands(ctx, 5, ContactInfosIDs, busDomain.Brand)
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

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 25, ContactInfosIDs, busDomain.Supplier)
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

	return unitest.SeedData{
		Admins:            []unitest.User{{User: admins[0]}},
		ContactInfos:      contactInfos,
		Brands:            brand,
		ProductCategories: productCategories,
		Products:          products,
		Suppliers:         suppliers,
		SupplierProducts:  supplierProducts,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []supplierproductbus.SupplierProduct{
				sd.SupplierProducts[0],
				sd.SupplierProducts[1],
				sd.SupplierProducts[2],
				sd.SupplierProducts[3],
				sd.SupplierProducts[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.SupplierProduct.Query(ctx, supplierproductbus.QueryFilter{}, supplierproductbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]supplierproductbus.SupplierProduct)
				if !exists {
					return fmt.Sprintf("got is not a slice of supplier products: %v", got)
				}

				expResp := exp.([]supplierproductbus.SupplierProduct)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: supplierproductbus.SupplierProduct{
				ProductID:          sd.Products[2].ProductID,
				SupplierID:         sd.Suppliers[2].SupplierID,
				SupplierPartNumber: "NewSupplierPartNumber",
				MinOrderQuantity:   10,
				MaxOrderQuantity:   50,
				LeadTimeDays:       15,
				UnitCost:           types.MustParseMoney("15.99"),
				IsPrimarySupplier:  true,
			},
			ExcFunc: func(ctx context.Context) any {
				newSupplierProduct := supplierproductbus.NewSupplierProduct{
					ProductID:          sd.Products[2].ProductID,
					SupplierID:         sd.Suppliers[2].SupplierID,
					SupplierPartNumber: "NewSupplierPartNumber",
					MinOrderQuantity:   10,
					MaxOrderQuantity:   50,
					LeadTimeDays:       15,
					UnitCost:           types.MustParseMoney("15.99"),
					IsPrimarySupplier:  true,
				}

				supplierProduct, err := busDomain.SupplierProduct.Create(ctx, newSupplierProduct)
				if err != nil {
					return err
				}

				return supplierProduct
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(supplierproductbus.SupplierProduct)
				if !exists {
					return fmt.Sprintf("got is not a supplier product: %v", got)
				}

				expResp := exp.(supplierproductbus.SupplierProduct)
				expResp.SupplierProductID = gotResp.SupplierProductID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		}}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: supplierproductbus.SupplierProduct{
				SupplierProductID:  sd.SupplierProducts[0].SupplierProductID,
				ProductID:          sd.Products[1].ProductID,
				SupplierID:         sd.Suppliers[2].SupplierID,
				SupplierPartNumber: "UpdatedSupplierPartNumber",
				MinOrderQuantity:   15,
				MaxOrderQuantity:   25,
				LeadTimeDays:       2,
				UnitCost:           types.MustParseMoney("12.29"),
				IsPrimarySupplier:  true,
				CreatedDate:        sd.SupplierProducts[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				usp := supplierproductbus.UpdateSupplierProduct{
					ProductID:          &sd.Products[1].ProductID,
					SupplierID:         &sd.Suppliers[2].SupplierID,
					SupplierPartNumber: dbtest.StringPointer("UpdatedSupplierPartNumber"),
					MinOrderQuantity:   dbtest.IntPointer(15),
					MaxOrderQuantity:   dbtest.IntPointer(25),
					LeadTimeDays:       dbtest.IntPointer(2),
					UnitCost:           types.MustParseMoney("12.29").Ptr(),
					IsPrimarySupplier:  dbtest.BoolPointer(true),
				}

				updatedSupplierProduct, err := busDomain.SupplierProduct.Update(ctx, sd.SupplierProducts[0], usp)
				if err != nil {
					return err
				}

				return updatedSupplierProduct
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(supplierproductbus.SupplierProduct)
				if !exists {
					return fmt.Sprintf("got is not a supplier product: %v", got)
				}

				expResp := exp.(supplierproductbus.SupplierProduct)
				expResp.SupplierProductID = gotResp.SupplierProductID
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
				err := busDomain.SupplierProduct.Delete(ctx, sd.SupplierProducts[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

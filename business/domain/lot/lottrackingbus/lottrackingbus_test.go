package lottrackingbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_LotTracking(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_LotTracking")

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
	contactInfo, err := contactinfobus.TestSeedContactInfo(ctx, 5, busDomain.ContactInfo)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfo))
	for i, c := range contactInfo {
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

	lotTracking, err := lottrackingbus.TestSeedLotTracking(ctx, 15, supplierProductIDs, busDomain.LotTracking)
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
		LotTracking:       lotTracking,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []lottrackingbus.LotTracking{
				sd.LotTracking[0],
				sd.LotTracking[1],
				sd.LotTracking[2],
				sd.LotTracking[3],
				sd.LotTracking[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.LotTracking.Query(ctx, lottrackingbus.QueryFilter{}, lottrackingbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]lottrackingbus.LotTracking)
				if !exists {
					return fmt.Sprintf("got is not a slice of lot trackings: %v", got)
				}

				expResp := exp.([]lottrackingbus.LotTracking)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	md := lottrackingbus.RandomDate()
	ed := lottrackingbus.RandomDate()
	rd := lottrackingbus.RandomDate()

	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: lottrackingbus.LotTracking{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md,
				ExpirationDate:    ed,
				RecievedDate:      rd,
				Quantity:          15,
				QualityStatus:     "good",
			},
			ExcFunc: func(ctx context.Context) any {
				newLotTracking := lottrackingbus.NewLotTracking{
					SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
					LotNumber:         "LotNumber",
					ManufactureDate:   md,
					ExpirationDate:    ed,
					RecievedDate:      rd,
					Quantity:          15,
					QualityStatus:     "good",
				}

				s, err := busDomain.LotTracking.Create(ctx, newLotTracking)
				if err != nil {
					return err
				}

				return s

			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(lottrackingbus.LotTracking)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(lottrackingbus.LotTracking)

				expResp.LotID = gotResp.LotID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	md := lottrackingbus.RandomDate()
	ed := lottrackingbus.RandomDate()
	rd := lottrackingbus.RandomDate()

	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: lottrackingbus.LotTracking{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "UpdatedLotNumber",
				ManufactureDate:   md,
				ExpirationDate:    ed,
				RecievedDate:      rd,
				Quantity:          20,
				QualityStatus:     "good",
			},
			ExcFunc: func(ctx context.Context) any {
				updatedLotTracking := lottrackingbus.UpdateLotTracking{
					SupplierProductID: &sd.SupplierProducts[0].SupplierProductID,
					LotNumber:         dbtest.StringPointer("UpdatedLotNumber"),
					ManufactureDate:   &md,
					ExpirationDate:    &ed,
					RecievedDate:      &rd,
					Quantity:          dbtest.IntPointer(20),
					QualityStatus:     dbtest.StringPointer("good"),
				}

				lt, err := busDomain.LotTracking.Update(ctx, sd.LotTracking[3], updatedLotTracking)
				if err != nil {
					return err
				}

				return lt
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(lottrackingbus.LotTracking)
				if !exists {
					return fmt.Sprintf("got is not a product cost: %v", got)
				}

				expResp := exp.(lottrackingbus.LotTracking)

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
				err := busDomain.LotTracking.Delete(ctx, sd.LotTracking[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

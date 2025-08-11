package serialnumberbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/lot/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_SerialNumbers(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_SerialNumbers")
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

	adminIDs := make([]uuid.UUID, len(admins))
	for i, admin := range admins {
		adminIDs[i] = admin.ID
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

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 5, strIDs, busDomain.ContactInfos)
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

	lotTrackingsIDs := make(uuid.UUIDs, len(lotTrackings))

	for i, lt := range lotTrackings {
		lotTrackingsIDs[i] = lt.LotID
	}

	warehouseCount := 5

	// WAREHOUSES
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, warehouseCount, adminIDs[0], strIDs, busDomain.Warehouse)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 12, warehouseIDs, busDomain.Zones)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 25, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationIDs := make(uuid.UUIDs, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	serialNumbers, err := serialnumberbus.TestSeedSerialNumbers(ctx, 50, lotTrackingsIDs, productIDs, inventoryLocationIDs, busDomain.SerialNumber)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding serial numbers : %w", err)
	}

	return unitest.SeedData{
		Admins:             []unitest.User{{User: admins[0]}},
		Products:           products,
		LotTrackings:       lotTrackings,
		InventoryLocations: inventoryLocations,
		SerialNumbers:      serialNumbers,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []serialnumberbus.SerialNumber{
				sd.SerialNumbers[0],
				sd.SerialNumbers[1],
				sd.SerialNumbers[2],
				sd.SerialNumbers[3],
				sd.SerialNumbers[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.SerialNumber.Query(ctx, serialnumberbus.QueryFilter{}, serialnumberbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]serialnumberbus.SerialNumber)
				if !exists {
					return fmt.Sprintf("got is not a slice of lot trackings: %v", got)
				}

				expResp := exp.([]serialnumberbus.SerialNumber)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: serialnumberbus.SerialNumber{
				LotID:        sd.LotTrackings[0].LotID,
				ProductID:    sd.Products[0].ProductID,
				LocationID:   sd.InventoryLocations[0].LocationID,
				SerialNumber: "SN123456789",
				Status:       "active",
			},
			ExcFunc: func(ctx context.Context) any {
				newSN := serialnumberbus.NewSerialNumber{
					LotID:        sd.LotTrackings[0].LotID,
					ProductID:    sd.Products[0].ProductID,
					LocationID:   sd.InventoryLocations[0].LocationID,
					SerialNumber: "SN123456789",
					Status:       "active",
				}
				got, err := busDomain.SerialNumber.Create(ctx, newSN)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(serialnumberbus.SerialNumber)
				if !exists {
					return fmt.Sprintf("got is not a lot tracking: %v", got)
				}

				expResp := exp.(serialnumberbus.SerialNumber)
				expResp.SerialID = gotResp.SerialID
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
			Name: "update",
			ExpResp: serialnumberbus.SerialNumber{
				SerialID:     sd.SerialNumbers[0].SerialID,
				LotID:        sd.LotTrackings[1].LotID,
				ProductID:    sd.Products[1].ProductID,
				LocationID:   sd.InventoryLocations[1].LocationID,
				SerialNumber: "SN987654321",
				Status:       "inactive",
				CreatedDate:  sd.SerialNumbers[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				updateSN := serialnumberbus.UpdateSerialNumber{
					LotID:        &sd.LotTrackings[1].LotID,
					ProductID:    &sd.Products[1].ProductID,
					LocationID:   &sd.InventoryLocations[1].LocationID,
					SerialNumber: dbtest.StringPointer("SN987654321"),
					Status:       dbtest.StringPointer("inactive"),
				}
				got, err := busDomain.SerialNumber.Update(ctx, sd.SerialNumbers[0], updateSN)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(serialnumberbus.SerialNumber)
				if !exists {
					return fmt.Sprintf("got is not a lot tracking: %v", got)
				}

				expResp := exp.(serialnumberbus.SerialNumber)
				expResp.SerialID = gotResp.SerialID
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Delete",
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.SerialNumber.Delete(ctx, sd.SerialNumbers[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

package inventoryitembus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_InventoryItem(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_InventoryItem")

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

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 5, busDomain.ContactInfos)
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

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	// ADDRESSES
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

	warehouseCount := 5

	ctys, err := citybus.TestSeedCities(ctx, warehouseCount, ids, busDomain.City)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	ctyIDs := make([]uuid.UUID, 0, len(ctys))
	for _, c := range ctys {
		ctyIDs = append(ctyIDs, c.ID)
	}

	strs, err := streetbus.TestSeedStreets(ctx, warehouseCount, ctyIDs, busDomain.Street)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding streets : %w", err)
	}
	strIDs := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		strIDs = append(strIDs, s.ID)
	}

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

	inventoryLocationsIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationsIDs[i] = il.LocationID
	}

	inventoryItems, err := inventoryitembus.TestSeedInventoryItems(ctx, 30, inventoryLocationsIDs, productIDs, busDomain.InventoryItem)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inventory products : %w", err)
	}

	return unitest.SeedData{
		Admins:             []unitest.User{{User: admins[0]}},
		Products:           products,
		InventoryLocations: inventoryLocations,
		InventoryItems:     inventoryItems,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []inventoryitembus.InventoryItem{
				sd.InventoryItems[0],
				sd.InventoryItems[1],
				sd.InventoryItems[2],
				sd.InventoryItems[3],
				sd.InventoryItems[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryItem.Query(ctx, inventoryitembus.QueryFilter{}, inventoryitembus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]inventoryitembus.InventoryItem)
				if !exists {
					return fmt.Sprintf("got is not a slice of inventory products: %v", got)
				}

				expResp := exp.([]inventoryitembus.InventoryItem)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: inventoryitembus.InventoryItem{
				LocationID:            sd.InventoryLocations[0].LocationID,
				ProductID:             sd.Products[0].ProductID,
				Quantity:              10,
				ReservedQuantity:      15,
				AllocatedQuantity:     20,
				MinimumStock:          1,
				MaximumStock:          100,
				ReorderPoint:          5,
				EconomicOrderQuantity: 25,
				SafetyStock:           40,
				AvgDailyUsage:         6,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryItem.Create(ctx, inventoryitembus.NewInventoryItem{
					LocationID:            sd.InventoryLocations[0].LocationID,
					ProductID:             sd.Products[0].ProductID,
					Quantity:              10,
					ReservedQuantity:      15,
					AllocatedQuantity:     20,
					MinimumStock:          1,
					MaximumStock:          100,
					ReorderPoint:          5,
					EconomicOrderQuantity: 25,
					SafetyStock:           40,
					AvgDailyUsage:         6,
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventoryitembus.InventoryItem)
				if !exists {
					return fmt.Sprintf("got is not an inventory product: %v", got)
				}

				expResp := exp.(inventoryitembus.InventoryItem)
				expResp.ItemID = gotResp.ItemID
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
			ExpResp: inventoryitembus.InventoryItem{
				ItemID:                sd.InventoryItems[0].ItemID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				ProductID:             sd.Products[0].ProductID,
				Quantity:              20,
				ReservedQuantity:      25,
				AllocatedQuantity:     30,
				MinimumStock:          2,
				MaximumStock:          102,
				ReorderPoint:          10,
				EconomicOrderQuantity: 50,
				SafetyStock:           60,
				AvgDailyUsage:         10,
				CreatedDate:           sd.InventoryItems[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryItem.Update(ctx, sd.InventoryItems[0], inventoryitembus.UpdateInventoryItem{
					LocationID:            &sd.InventoryLocations[0].LocationID,
					ProductID:             &sd.Products[0].ProductID,
					Quantity:              dbtest.IntPointer(20),
					ReservedQuantity:      dbtest.IntPointer(25),
					AllocatedQuantity:     dbtest.IntPointer(30),
					MinimumStock:          dbtest.IntPointer(2),
					MaximumStock:          dbtest.IntPointer(102),
					ReorderPoint:          dbtest.IntPointer(10),
					EconomicOrderQuantity: dbtest.IntPointer(50),
					SafetyStock:           dbtest.IntPointer(60),
					AvgDailyUsage:         dbtest.IntPointer(10),
				})
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventoryitembus.InventoryItem)
				if !exists {
					return fmt.Sprintf("got is not an inventory product: %v", got)
				}

				expResp := exp.(inventoryitembus.InventoryItem)
				expResp.ItemID = gotResp.ItemID
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
				return busDomain.InventoryItem.Delete(ctx, sd.InventoryItems[0])
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

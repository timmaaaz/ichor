package inventorytransactionbus_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_InventoryTransaction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_InventoryTransaction")

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
	warehouseCount := 5

	// USERS
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}
	adminIDs := make([]uuid.UUID, len(admins))
	for i, admin := range admins {
		adminIDs[i] = admin.ID
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

	// ADDRESSES
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	ids := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		ids = append(ids, r.ID)
	}

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

	inventoryLocationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 3, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	inventoryTransactions, err := inventorytransactionbus.TestSeedInventoryTransaction(ctx, 40, inventoryLocationIDs, productIDs, userIDs, busDomain.InventoryTransaction)

	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inventory transactions : %w", err)
	}

	return unitest.SeedData{
		Admins:                []unitest.User{{User: admins[0]}},
		Products:              products,
		Users:                 []unitest.User{{User: users[0]}, {User: users[1]}, {User: users[2]}},
		InventoryLocations:    inventoryLocations,
		InventoryTransactions: inventoryTransactions,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "query",
			ExpResp: sd.InventoryTransactions[:5],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryTransaction.Query(ctx, inventorytransactionbus.QueryFilter{}, inventorytransactionbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return fmt.Errorf("querying inventory transactions: %w", err)
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]inventorytransactionbus.InventoryTransaction)
				if !exists {
					return fmt.Sprintf("got is not a slice of inventorytransactionbus.InventoryTransaction: %v", got)
				}

				expResp := exp.([]inventorytransactionbus.InventoryTransaction)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now()
	return []unitest.Table{
		{
			Name: "create",
			ExpResp: inventorytransactionbus.InventoryTransaction{
				LocationID:      sd.InventoryLocations[0].LocationID,
				ProductID:       sd.Products[0].ProductID,
				UserID:          sd.Users[0].ID,
				Quantity:        10,
				TransactionType: "IN",
				ReferenceNumber: "ABC123",
				TransactionDate: now,
			},
			ExcFunc: func(ctx context.Context) any {
				newIT := inventorytransactionbus.NewInventoryTransaction{
					LocationID:      sd.InventoryLocations[0].LocationID,
					ProductID:       sd.Products[0].ProductID,
					UserID:          sd.Users[0].ID,
					Quantity:        10,
					TransactionType: "IN",
					ReferenceNumber: "ABC123",
					TransactionDate: now,
				}

				it, err := busDomain.InventoryTransaction.Create(ctx, newIT)
				if err != nil {
					return fmt.Errorf("creating inventory transaction: %w", err)
				}
				return it
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventorytransactionbus.InventoryTransaction)
				if !exists {
					return fmt.Sprintf("got is not an inventorytransactionbus.InventoryTransaction: %v", got)
				}
				expResp := exp.(inventorytransactionbus.InventoryTransaction)

				expResp.InventoryTransactionID = gotResp.InventoryTransactionID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now()

	return []unitest.Table{
		{
			Name: "update",
			ExpResp: inventorytransactionbus.InventoryTransaction{
				InventoryTransactionID: sd.InventoryTransactions[0].InventoryTransactionID,
				LocationID:             sd.InventoryLocations[0].LocationID,
				ProductID:              sd.Products[0].ProductID,
				UserID:                 sd.Users[0].ID,
				Quantity:               20,
				TransactionType:        "OUT",
				ReferenceNumber:        "XYZ789",
				TransactionDate:        now,
				CreatedDate:            sd.InventoryTransactions[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				updateIT := inventorytransactionbus.UpdateInventoryTransaction{
					ProductID:       &sd.Products[0].ProductID,
					UserID:          &sd.Users[0].ID,
					LocationID:      &sd.InventoryLocations[0].LocationID,
					Quantity:        dbtest.IntPointer(20),
					TransactionType: dbtest.StringPointer("OUT"),
					ReferenceNumber: dbtest.StringPointer("XYZ789"),
					TransactionDate: &now,
				}

				it, err := busDomain.InventoryTransaction.Update(ctx, sd.InventoryTransactions[0], updateIT)
				if err != nil {
					return fmt.Errorf("updating inventory transaction: %w", err)
				}

				return it
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventorytransactionbus.InventoryTransaction)
				if !exists {
					return fmt.Sprintf("got is not an inventorytransactionbus.InventoryTransaction: %v", got)
				}
				expResp := exp.(inventorytransactionbus.InventoryTransaction)

				expResp.InventoryTransactionID = gotResp.InventoryTransactionID
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.InventoryTransaction.Delete(ctx, sd.InventoryTransactions[0])
				if err != nil {
					return fmt.Errorf("deleting inventory transaction: %w", err)
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

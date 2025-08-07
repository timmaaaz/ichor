package inventoryadjustmentbus_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/movement/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_InventoryAdjustment(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_InventoryAdjustment")
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

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 6, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	inventoryAdjustments, err := inventoryadjustmentbus.TestSeedInventoryAdjustments(ctx, 20, productIDs, inventoryLocationIDs, userIDs[:2], userIDs[2:], busDomain.InventoryAdjustment)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding inventory adjustments : %w", err)
	}

	return unitest.SeedData{
		Admins:               []unitest.User{{User: admins[0]}},
		Products:             products,
		InventoryAdjustments: inventoryAdjustments,
		InventoryLocations:   inventoryLocations,
	}, nil

}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "query",
			ExpResp: sd.InventoryAdjustments[:5],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryAdjustment.Query(ctx, inventoryadjustmentbus.QueryFilter{}, inventoryadjustmentbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return fmt.Errorf("querying inventory adjustments: %w", err)
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]inventoryadjustmentbus.InventoryAdjustment)
				if !exists {
					return fmt.Sprintf("got is not a slice of inventorytransactionbus.InventoryTransaction: %v", got)
				}

				expResp := exp.([]inventoryadjustmentbus.InventoryAdjustment)
				return (cmp.Diff(gotResp, expResp))
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now()

	return []unitest.Table{
		{
			Name: "create",
			ExpResp: inventoryadjustmentbus.InventoryAdjustment{
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: 10,
				ReasonCode:     "Purchase",
				Notes:          "New purchase",
				AdjustmentDate: now,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryAdjustment.Create(ctx, inventoryadjustmentbus.NewInventoryAdjustment{
					ProductID:      sd.Products[0].ProductID,
					LocationID:     sd.InventoryLocations[0].LocationID,
					AdjustedBy:     sd.InventoryAdjustments[0].AdjustedBy,
					ApprovedBy:     sd.InventoryAdjustments[0].ApprovedBy,
					QuantityChange: 10,
					ReasonCode:     "Purchase",
					Notes:          "New purchase",
					AdjustmentDate: now,
				})
				if err != nil {
					return fmt.Errorf("creating inventory adjustment: %w", err)
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventoryadjustmentbus.InventoryAdjustment)
				if !exists {
					return fmt.Sprintf("got is not inventorytransactionbus.InventoryTransaction: %v", got)
				}

				expResp := exp.(inventoryadjustmentbus.InventoryAdjustment)

				expResp.InventoryAdjustmentID = gotResp.InventoryAdjustmentID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return (cmp.Diff(gotResp, expResp))
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now()

	return []unitest.Table{
		{
			Name: "update",
			ExpResp: inventoryadjustmentbus.InventoryAdjustment{
				InventoryAdjustmentID: sd.InventoryAdjustments[1].InventoryAdjustmentID,
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				AdjustedBy:            sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:            sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange:        20,
				ReasonCode:            "Adjustment",
				Notes:                 "Updated adjustment",
				AdjustmentDate:        now,
				CreatedDate:           sd.InventoryAdjustments[1].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryAdjustment.Update(ctx, sd.InventoryAdjustments[1], inventoryadjustmentbus.UpdateInventoryAdjustment{
					ProductID:      &sd.Products[0].ProductID,
					LocationID:     &sd.InventoryLocations[0].LocationID,
					AdjustedBy:     &sd.InventoryAdjustments[0].AdjustedBy,
					ApprovedBy:     &sd.InventoryAdjustments[0].ApprovedBy,
					QuantityChange: dbtest.IntPointer(20),
					ReasonCode:     dbtest.StringPointer("Adjustment"),
					Notes:          dbtest.StringPointer("Updated adjustment"),
					AdjustmentDate: &now,
				})
				if err != nil {
					return fmt.Errorf("updating inventory adjustment: %w", err)
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventoryadjustmentbus.InventoryAdjustment)
				if !exists {
					return fmt.Sprintf("got is not inventorytransactionbus.InventoryTransaction: %v", got)
				}

				expResp := exp.(inventoryadjustmentbus.InventoryAdjustment)

				expResp.InventoryAdjustmentID = gotResp.InventoryAdjustmentID
				expResp.UpdatedDate = gotResp.UpdatedDate

				return (cmp.Diff(gotResp, expResp))
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "delete",
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.InventoryAdjustment.Delete(ctx, sd.InventoryAdjustments[2])
				if err != nil {
					return fmt.Errorf("deleting inventory adjustment: %w", err)
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

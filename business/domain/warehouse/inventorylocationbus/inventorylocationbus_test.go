package inventorylocationbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus/types"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_InventoryLocations(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_InventoryLocations")

	sd, err := insertSeedData(db.BusDomain)

	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Run tests
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

	return unitest.SeedData{
		Admins:             []unitest.User{{User: admins[0]}},
		Warehouses:         warehouses,
		Zones:              zones,
		InventoryLocations: inventoryLocations,
	}, nil

}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []inventorylocationbus.InventoryLocation{
				sd.InventoryLocations[0],
				sd.InventoryLocations[1],
				sd.InventoryLocations[2],
				sd.InventoryLocations[3],
				sd.InventoryLocations[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.InventoryLocation.Query(ctx, inventorylocationbus.QueryFilter{}, inventorylocationbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				return cmp.Diff(exp, got)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: inventorylocationbus.InventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     true,
				IsReserveLocation:  false,
				MaxCapacity:        100,
				CurrentUtilization: types.MustParseRoundedFloat("0.57"),
			},
			ExcFunc: func(ctx context.Context) any {
				newInventoryLocation := inventorylocationbus.NewInventoryLocation{
					WarehouseID:        sd.Warehouses[0].ID,
					ZoneID:             sd.Zones[0].ZoneID,
					Aisle:              "NewAisle",
					Rack:               "NewRack",
					Shelf:              "NewShelf",
					Bin:                "NewBin",
					IsPickLocation:     true,
					IsReserveLocation:  false,
					MaxCapacity:        100,
					CurrentUtilization: types.MustParseRoundedFloat("0.57"),
				}
				got, err := busDomain.InventoryLocation.Create(ctx, newInventoryLocation)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {

				gotResp, exists := got.(inventorylocationbus.InventoryLocation)
				if !exists {
					return fmt.Sprintf("got is not an inventory location %v", got)
				}

				expResp := exp.(inventorylocationbus.InventoryLocation)
				expResp.LocationID = gotResp.LocationID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(expResp, gotResp)

			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: inventorylocationbus.InventoryLocation{
				LocationID:         sd.InventoryLocations[0].LocationID,
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "UpdatedAisle",
				Rack:               "UpdatedRack",
				Shelf:              "UpdatedShelf",
				Bin:                "UpdatedBin",
				IsPickLocation:     true,
				IsReserveLocation:  false,
				MaxCapacity:        100,
				CurrentUtilization: types.MustParseRoundedFloat("0.57"),
				CreatedDate:        sd.InventoryLocations[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				updatedInventoryLocation := inventorylocationbus.UpdateInventoryLocation{
					WarehouseID:        &sd.Warehouses[0].ID,
					ZoneID:             &sd.Zones[0].ZoneID,
					Aisle:              dbtest.StringPointer("UpdatedAisle"),
					Rack:               dbtest.StringPointer("UpdatedRack"),
					Shelf:              dbtest.StringPointer("UpdatedShelf"),
					Bin:                dbtest.StringPointer("UpdatedBin"),
					IsPickLocation:     dbtest.BoolPointer(true),
					IsReserveLocation:  dbtest.BoolPointer(false),
					MaxCapacity:        dbtest.IntPointer(100),
					CurrentUtilization: types.MustParseRoundedFloat("0.57").Ptr(),
				}
				got, err := busDomain.InventoryLocation.Update(ctx, sd.InventoryLocations[0], updatedInventoryLocation)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(inventorylocationbus.InventoryLocation)
				if !exists {
					return fmt.Sprintf("got is not an inventory location %v", got)
				}

				expResp := exp.(inventorylocationbus.InventoryLocation)

				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(expResp, gotResp)
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
				err := busDomain.InventoryLocation.Delete(ctx, sd.InventoryLocations[0])
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

package zonebus_test

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
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Zones(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Zones")

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

	return unitest.SeedData{
		Admins:     []unitest.User{{User: admins[0]}},
		Warehouses: warehouses,
		Zones:      zones,
	}, nil

}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []zonebus.Zone{
				sd.Zones[0],
				sd.Zones[1],
				sd.Zones[2],
				sd.Zones[3],
				sd.Zones[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Zones.Query(ctx, zonebus.QueryFilter{}, zonebus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]zonebus.Zone)
				if !exists {
					return fmt.Sprintf("got is not a slice of zones: %v", got)
				}

				expResp := exp.([]zonebus.Zone)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: zonebus.Zone{
				Name:        "New Zone",
				Description: "a new zone",
				WarehouseID: sd.Warehouses[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				newZone := zonebus.NewZone{
					Name:        "New Zone",
					Description: "a new zone",
					WarehouseID: sd.Warehouses[0].ID,
				}
				got, err := busDomain.Zones.Create(ctx, newZone)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(zonebus.Zone)
				if !exists {
					return fmt.Sprintf("got is not a zone: %v", got)
				}

				expResp := exp.(zonebus.Zone)
				expResp.ZoneID = gotResp.ZoneID
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
			ExpResp: zonebus.Zone{
				ZoneID:      sd.Zones[0].ZoneID,
				Name:        "Updated Zone",
				Description: "an updated zone",
				WarehouseID: sd.Warehouses[0].ID,
				CreatedDate: sd.Zones[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				updatedZone := zonebus.UpdateZone{
					WarehouseID: &sd.Warehouses[0].ID,
					Name:        dbtest.StringPointer("Updated Zone"),
					Description: dbtest.StringPointer("an updated zone"),
				}
				got, err := busDomain.Zones.Update(ctx, sd.Zones[0], updatedZone)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(zonebus.Zone)
				if !exists {
					return fmt.Sprintf("got is not a zone: %v", got)
				}

				expResp := exp.(zonebus.Zone)
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
				err := busDomain.Zones.Delete(ctx, sd.Zones[0])
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

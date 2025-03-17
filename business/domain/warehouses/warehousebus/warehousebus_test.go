package warehousebus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouses/warehousebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Warehouse(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Warehouse")
	sd, err := insertSeedData(db.BusDomain)

	if err != nil {
		t.Fatalf("Error inserting seed data: %s", err)
	}
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

	seedUsers := make([]unitest.User, len(admins))
	for i, u := range admins {
		seedUsers[i] = unitest.User{
			User: u,
		}
	}

	return unitest.SeedData{
		Streets:    strs,
		Warehouses: warehouses,
		Admins:     seedUsers,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Copy over to exp
	exp := make([]warehousebus.Warehouse, len(sd.Warehouses))
	copy(exp, sd.Warehouses)

	sort.Slice(exp, func(i, j int) bool {
		return exp[i].ID.String() < exp[j].ID.String()
	})

	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []warehousebus.Warehouse{
				exp[0],
				exp[1],
				exp[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Warehouse.Query(ctx, warehousebus.QueryFilter{}, order.NewBy(warehousebus.OrderByID, order.ASC), page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.([]warehousebus.Warehouse)
				if !exists {
					return fmt.Sprintf("expected []warehousebus.Warehouse, got %T", got)
				}

				expResp := exp.([]warehousebus.Warehouse)
				if len(gotResp) != len(expResp) {
					return fmt.Sprintf("expected %d rows, got %d", len(expResp), len(gotResp))
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: warehousebus.Warehouse{
				StreetID:  sd.Streets[0].ID,
				Name:      "Test Warehouse",
				IsActive:  true,
				CreatedBy: sd.Admins[0].ID,
				UpdatedBy: sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				new := warehousebus.NewWarehouse{
					StreetID:  sd.Streets[0].ID,
					Name:      "Test Warehouse",
					CreatedBy: sd.Admins[0].ID,
				}
				got, err := busDomain.Warehouse.Create(ctx, new)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(warehousebus.Warehouse)
				if !exists {
					return fmt.Sprintf("expected warehousebus.Warehouse, got %T", got)
				}

				expResp := exp.(warehousebus.Warehouse)
				if gotResp.ID == uuid.Nil {
					return "expected non-zero ID"
				}

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: warehousebus.Warehouse{
				ID:        sd.Warehouses[0].ID,
				StreetID:  sd.Streets[0].ID,
				Name:      "Update Test Warehouse",
				IsActive:  true,
				CreatedBy: sd.Admins[0].ID,
				UpdatedBy: sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				update := warehousebus.UpdateWarehouse{
					StreetID:  &sd.Streets[0].ID,
					Name:      dbtest.StringPointer("Update Test Warehouse"),
					UpdatedBy: &sd.Admins[0].ID,
				}
				got, err := busDomain.Warehouse.Update(ctx, sd.Warehouses[0], update)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(warehousebus.Warehouse)
				if !exists {
					return fmt.Sprintf("expected warehousebus.Warehouse, got %T", got)
				}

				expResp := exp.(warehousebus.Warehouse)
				expResp.DateUpdated = gotResp.DateUpdated
				expResp.DateCreated = gotResp.DateCreated

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "admin",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Warehouse.Delete(ctx, sd.Warehouses[0]); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

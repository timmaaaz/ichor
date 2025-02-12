package reportstobus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ReportsTo(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ReportsTo")
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

	// ============= User Creation =================

	reporters, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding reporter : %w", err)
	}

	reporterIDs := make([]uuid.UUID, len(reporters))
	for i, r := range reporters {
		reporterIDs[i] = r.ID
	}

	bosses, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding reporter : %w", err)
	}

	bossIDs := make([]uuid.UUID, len(bosses))
	for i, b := range bosses {
		bossIDs[i] = b.ID
	}

	// ============= ReportsTo Creation =================

	reportsTo, err := reportstobus.TestSeedReportsTo(ctx, 20, reporterIDs, bossIDs, busDomain.ReportsTo)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding reportsto : %w", err)
	}

	u := append(reporters, bosses...)

	users := []unitest.User{}

	for _, usr := range u {
		users = append(users, unitest.User{User: usr})
	}

	return unitest.SeedData{
		Admins:    []unitest.User{{User: admins[0]}},
		ReportsTo: reportsTo,
		Users:     users,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []reportstobus.ReportsTo{
				sd.ReportsTo[0],
				sd.ReportsTo[1],
				sd.ReportsTo[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.ReportsTo.Query(ctx, reportstobus.QueryFilter{}, reportstobus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.([]reportstobus.ReportsTo)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]reportstobus.ReportsTo)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Create",
			ExpResp: sd.ReportsTo[0],
			ExcFunc: func(ctx context.Context) any {
				nat := reportstobus.NewReportsTo{
					BossID:     sd.ReportsTo[0].BossID,
					ReporterID: sd.ReportsTo[0].ReporterID,
				}

				got, err := busDomain.ReportsTo.Create(ctx, nat)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(reportstobus.ReportsTo)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(reportstobus.ReportsTo)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: reportstobus.ReportsTo{
				ID:         sd.ReportsTo[0].ID,
				BossID:     sd.ReportsTo[1].BossID,
				ReporterID: sd.ReportsTo[1].ReporterID,
			},
			ExcFunc: func(ctx context.Context) any {
				uat := reportstobus.UpdateReportsTo{
					BossID:     &sd.ReportsTo[1].BossID,
					ReporterID: &sd.ReportsTo[1].ReporterID,
				}

				got, err := busDomain.ReportsTo.Update(ctx, sd.ReportsTo[0], uat)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(reportstobus.ReportsTo)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(reportstobus.ReportsTo)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.ReportsTo.Delete(ctx, sd.ReportsTo[0])
				if err != nil {
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

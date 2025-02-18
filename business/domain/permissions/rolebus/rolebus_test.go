package rolebus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Role(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Role")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}
	unitest.Run(t, query(db.BusDomain, sd), "query")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	roles, err := rolebus.TestSeedRoles(ctx, 5, busDomain.Role)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	return unitest.SeedData{
		Roles: roles,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []rolebus.Role{
				sd.Roles[0],
				sd.Roles[1],
				sd.Roles[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Role.Query(ctx, rolebus.QueryFilter{}, rolebus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]rolebus.Role)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]rolebus.Role)
				if !exists {
					return "error occurred"
				}

				for i := range gotResp {
					expResp[i].ID = gotResp[i].ID
				}

				return cmp.Diff(exp, got)
			},
		},
	}
}

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
	unitest.Run(t, create(db.BusDomain), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
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

func create(busDomain dbtest.BusDomain) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: rolebus.Role{
				Name:        "TestRole",
				Description: "TestRole Description",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Role.Create(ctx, rolebus.NewRole{
					Name:        "TestRole",
					Description: "TestRole Description",
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(rolebus.Role)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(rolebus.Role)
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
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: rolebus.Role{
				ID:          sd.Roles[0].ID,
				Name:        "UpdatedRole",
				Description: "UpdatedRole Description",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Role.Update(ctx, sd.Roles[0], rolebus.UpdateRole{
					Name:        dbtest.StringPointer("UpdatedRole"),
					Description: dbtest.StringPointer("UpdatedRole Description"),
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(rolebus.Role)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(rolebus.Role)
				if !exists {
					return "error occurred"
				}

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
				err := busDomain.Role.Delete(ctx, sd.Roles[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}

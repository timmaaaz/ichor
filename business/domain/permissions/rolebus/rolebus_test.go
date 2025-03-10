package rolebus_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Role(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Role")

	unitest.Run(t, query(db.BusDomain), "query")
	unitest.Run(t, create(db.BusDomain), "create")
	unitest.Run(t, update(db.BusDomain), "update")
	unitest.Run(t, delete(db.BusDomain), "delete")
}

func query(busDomain dbtest.BusDomain) []unitest.Table {
	exp := []rolebus.Role{
		{
			Name:        "ADMIN",
			Description: "System Administrator with full access",
		},
		{
			Name:        "EMPLOYEE",
			Description: "Regular employee with standard access",
		},
		{
			Name:        "FINANCE_ADMIN",
			Description: "Finance Department Administrator",
		},
	}

	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: exp,
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

func update(busDomain dbtest.BusDomain) []unitest.Table {
	r, err := busDomain.Role.Query(context.Background(), rolebus.QueryFilter{}, rolebus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		panic(err)
	}

	if len(r) == 0 {
		panic("no role found")
	}

	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: rolebus.Role{
				ID:          r[0].ID,
				Name:        "UpdatedRole",
				Description: "UpdatedRole Description",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Role.Update(ctx, r[0], rolebus.UpdateRole{
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

func delete(busDomain dbtest.BusDomain) []unitest.Table {
	r, err := busDomain.Role.Query(context.Background(), rolebus.QueryFilter{}, rolebus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		panic(err)
	}

	if len(r) == 0 {
		panic("no role found")
	}

	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				return busDomain.Role.Delete(ctx, r[0])
			},
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}

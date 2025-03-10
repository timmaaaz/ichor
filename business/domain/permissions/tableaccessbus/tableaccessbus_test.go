package tableaccessbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_TableAccess(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_TableAccess")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	// unitest.Run(t, create(db.BusDomain, sd), "create")
	// unitest.Run(t, update(db.BusDomain, sd), "update")
	// unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 3, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	seedUsers := make([]unitest.User, len(usrs))
	for i, u := range usrs {
		seedUsers[i] = unitest.User{
			User: u,
		}
	}

	IWork := busDomain.ApprovalStatus

	IDont := busDomain.Role

	roles, err := rolebus.TestSeedRoles(ctx, 3, busDomain.Role)

	return unitest.SeedData{
		Users: seedUsers,
		Roles: roles,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	ctx := context.Background()

	ta1, err := busDomain.TableAccess.Query(ctx, tableaccessbus.QueryFilter{RoleID: &sd.Roles[0].ID}, order.NewBy(tableaccessbus.OrderByID, order.ASC), page.MustParse("1", "1"))
	if err != nil {
		panic(err)
	}
	ta2, err := busDomain.TableAccess.Query(ctx, tableaccessbus.QueryFilter{RoleID: &sd.Roles[0].ID}, order.NewBy(tableaccessbus.OrderByID, order.ASC), page.MustParse("1", "1"))
	if err != nil {
		panic(err)
	}
	ta3, err := busDomain.TableAccess.Query(ctx, tableaccessbus.QueryFilter{RoleID: &sd.Roles[0].ID}, order.NewBy(tableaccessbus.OrderByID, order.ASC), page.MustParse("1", "1"))
	if err != nil {
		panic(err)
	}

	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []tableaccessbus.TableAccess{
				ta1[0],
				ta2[0],
				ta3[0],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.TableAccess.Query(ctx, tableaccessbus.QueryFilter{RoleID: &sd.Roles[0].ID}, order.NewBy(tableaccessbus.OrderByID, order.ASC), page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.([]tableaccessbus.TableAccess)
				if !exists {
					return fmt.Sprintf("expected []tableaccessbus.TableAccess, got %T", got)
				}

				expResp := exp.([]tableaccessbus.TableAccess)
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
			ExpResp: tableaccessbus.TableAccess{
				RoleID:    sd.Roles[0].ID,
				TableName: "valid_assets",
				CanCreate: true,
				CanRead:   true,
				CanUpdate: true,
				CanDelete: true,
			},
			ExcFunc: func(ctx context.Context) any {
				nta := tableaccessbus.NewTableAccess{
					RoleID:    sd.Roles[0].ID,
					TableName: "valid_assets",
					CanCreate: true,
					CanRead:   true,
					CanUpdate: true,
					CanDelete: true,
				}

				got, err := busDomain.TableAccess.Create(ctx, nta)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(tableaccessbus.TableAccess)
				if !exists {
					return fmt.Sprintf("expected tableaccessbus.TableAccess, got %T", got)
				}

				expResp := exp.(tableaccessbus.TableAccess)
				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	ta, err := busdomain.TableAccess.Query(context.Background(), tableaccessbus.QueryFilter{}, order.NewBy(tableaccessbus.OrderByID, order.ASC), page.MustParse("1", "1"))
	if err != nil {
		panic(err)
	}

	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: tableaccessbus.TableAccess{
				ID:        ta[0].ID,
				RoleID:    ta[0].RoleID,
				TableName: ta[0].TableName,
				CanCreate: false,
				CanRead:   false,
				CanUpdate: false,
				CanDelete: false,
			},
			ExcFunc: func(ctx context.Context) any {
				uta := tableaccessbus.UpdateTableAccess{
					CanCreate: dbtest.BoolPointer(false),
					CanRead:   dbtest.BoolPointer(false),
					CanUpdate: dbtest.BoolPointer(false),
					CanDelete: dbtest.BoolPointer(false),
				}

				got, err := busdomain.TableAccess.Update(ctx, sd.TableAccesses[0], uta)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(tableaccessbus.TableAccess)
				if !exists {
					return fmt.Sprintf("expected tableaccessbus.TableAccess, got %T", got)
				}

				expResp := exp.(tableaccessbus.TableAccess)

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func delete(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	ta, err := busdomain.TableAccess.Query(context.Background(), tableaccessbus.QueryFilter{}, order.NewBy(tableaccessbus.OrderByID, order.ASC), page.MustParse("1", "1"))
	if err != nil {
		panic(err)
	}

	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busdomain.TableAccess.Delete(ctx, ta[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(exp, got any) string {
				if got != nil {
					return fmt.Sprintf("expected nil, got %v", got)
				}
				return ""
			},
		},
	}
}

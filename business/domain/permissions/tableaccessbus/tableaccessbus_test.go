package tableaccessbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
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
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 3, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	roles, err := rolebus.TestSeedRoles(ctx, busDomain.Role)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	tableAccesses, err := tableaccessbus.TestSeedTableAccesses(ctx, roleIDs[0], busDomain.TableAccess)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding table accesses : %w", err)
	}

	seedUsers := make([]unitest.User, len(usrs))
	for i, u := range usrs {
		seedUsers[i] = unitest.User{
			User: u,
		}
	}

	return unitest.SeedData{
		Users:         seedUsers,
		Roles:         roles,
		TableAccesses: tableAccesses,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	// create sorted copy of table accesses
	tableAccesses := make([]tableaccessbus.TableAccess, len(sd.TableAccesses))
	copy(tableAccesses, sd.TableAccesses)
	sort.Slice(tableAccesses, func(i, j int) bool {
		return tableAccesses[i].ID.String() < tableAccesses[j].ID.String()
	})

	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []tableaccessbus.TableAccess{
				tableAccesses[0],
				tableAccesses[1],
				tableAccesses[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.TableAccess.Query(ctx, tableaccessbus.QueryFilter{}, order.NewBy(tableaccessbus.OrderByID, order.ASC), page.MustParse("1", "3"))
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
				RoleID:    sd.TableAccesses[1].RoleID,
				TableName: "valid_assets",
				CanCreate: sd.TableAccesses[0].CanCreate,
				CanRead:   sd.TableAccesses[0].CanRead,
				CanUpdate: sd.TableAccesses[0].CanUpdate,
				CanDelete: sd.TableAccesses[0].CanDelete,
			},
			ExcFunc: func(ctx context.Context) any {
				nta := tableaccessbus.NewTableAccess{
					RoleID:    sd.TableAccesses[1].RoleID,
					TableName: "valid_assets",
					CanCreate: sd.TableAccesses[0].CanCreate,
					CanRead:   sd.TableAccesses[0].CanRead,
					CanUpdate: sd.TableAccesses[0].CanUpdate,
					CanDelete: sd.TableAccesses[0].CanDelete,
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
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: tableaccessbus.TableAccess{
				ID:        sd.TableAccesses[0].ID,
				RoleID:    sd.TableAccesses[0].RoleID,
				TableName: sd.TableAccesses[0].TableName,
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
	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busdomain.TableAccess.Delete(ctx, sd.TableAccesses[0])
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

package permissionsbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

// User roles
// users[0]: admin

func Test_Permissions(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Permissions")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, cacheRole(db.BusDomain, sd), "cacheRole")
	unitest.Run(t, cacheTableAccess(db.BusDomain, sd), "cacheTableAccess")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	users, err := busDomain.User.Query(ctx, userbus.QueryFilter{}, order.NewBy(userbus.OrderByUsername, order.ASC), page.MustParse("1", "100"))
	if err != nil {
		return unitest.SeedData{}, err
	}
	seedUsers := make([]unitest.User, len(users))
	for i, u := range users {
		seedUsers[i] = unitest.User{
			User: u,
		}
	}

	roles, err := rolebus.TestSeedRoles(ctx, len(users), busDomain.Role)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}
	userIDs := make(uuid.UUIDs, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	userRoles, err := userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user roles : %w", err)
	}

	tas, err := tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding table access : %w", err)
	}

	// CONSTRUCT SEED DATA
	return unitest.SeedData{
		Users:         seedUsers,
		Roles:         roles,
		UserRoles:     userRoles,
		TableAccesses: tas,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	testPerm := permissionsbus.UserPermissions{
		UserID: sd.Users[0].ID,
	}
	testUser := sd.Users[0]
	var testRoleID uuid.UUID
	tas := make(map[string]tableaccessbus.TableAccess)

	for _, ur := range sd.UserRoles {
		if ur.UserID == testUser.ID {
			testRoleID = ur.RoleID
			testPerm.Role = &ur
			break
		}
	}

	for _, r := range sd.Roles {
		if r.ID == testRoleID {
			testPerm.RoleName = r.Name
			break
		}
	}

	for _, ta := range sd.TableAccesses {
		if ta.RoleID == testRoleID {
			tas[ta.TableName] = ta
		}
	}
	testPerm.TableAccess = tas

	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: testPerm,
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Permissions.QueryUserPermissions(ctx, sd.Users[0].ID)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(permissionsbus.UserPermissions)
				if !exists {
					return "got is not a *permissionsbus.UserPermissions"
				}

				expResp, exists := exp.(permissionsbus.UserPermissions)
				if !exists {
					return "exp is not a *permissionsbus.UserPermissions"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func cacheRole(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	ctx := context.Background()

	userPerms, err := busDomain.Permissions.QueryUserPermissions(ctx, sd.Users[0].ID)
	if err != nil {
		panic(err)
	}
	roleName := userPerms.RoleName
	roleID := userPerms.Role.RoleID
	newRoleName := roleName + " updated"

	// make a copy of user perms for comparison into exp
	exp := userPerms
	exp.RoleName = newRoleName

	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				r, err := busDomain.Role.QueryByID(ctx, roleID)
				if err != nil {
					return fmt.Errorf("role not found")
				}
				ur := rolebus.UpdateRole{
					Name: &newRoleName,
				}

				_, err = busDomain.Role.Update(ctx, r, ur)
				if err != nil {
					return err
				}

				got, err := busDomain.Permissions.QueryUserPermissions(ctx, sd.Users[0].ID)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(permissionsbus.UserPermissions)
				if !exists {
					return "got is not a *permissionsbus.UserPermissions"
				}

				expResp, exists := exp.(permissionsbus.UserPermissions)
				if !exists {
					return "exp is not a *permissionsbus.UserPermissions"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func cacheTableAccess(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	ctx := context.Background()

	userPerms, err := busDomain.Permissions.QueryUserPermissions(ctx, sd.Users[0].ID)
	if err != nil {
		panic(err)
	}

	tas := userPerms.TableAccess
	ta := tas["user_assets"]

	tmp := ta
	tmp.CanCreate = !tmp.CanCreate
	tmp.CanRead = !tmp.CanRead
	tmp.CanUpdate = !tmp.CanUpdate
	tmp.CanDelete = !tmp.CanDelete
	exp := userPerms
	exp.TableAccess["user_assets"] = tmp

	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				uta := tableaccessbus.UpdateTableAccess{
					CanCreate: dbtest.BoolPointer(tmp.CanCreate),
					CanRead:   dbtest.BoolPointer(tmp.CanRead),
					CanUpdate: dbtest.BoolPointer(tmp.CanUpdate),
					CanDelete: dbtest.BoolPointer(tmp.CanDelete),
				}
				test, err := busDomain.TableAccess.Update(ctx, ta, uta)
				if err != nil {
					return err
				}
				_ = test
				got, err := busDomain.Permissions.QueryUserPermissions(ctx, sd.Users[0].ID)
				if err != nil {
					return err
				}

				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(permissionsbus.UserPermissions)
				if !exists {
					return "got is not a *permissionsbus.UserPermissions"
				}

				expResp, exists := exp.(permissionsbus.UserPermissions)
				if !exists {
					return "exp is not a *permissionsbus.UserPermissions"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

package tableaccess_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/permissions/tableaccessapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/permissions/roleapp"
	"github.com/timmaaaz/ichor/app/domain/permissions/tableaccessapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu2 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	// =========================================================================
	// Permissions stuff
	// =========================================================================
	roles, err := rolebus.TestSeedRoles(ctx, 12, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	// Assign roles to users - EXCLUDE the last role from both users
	// This ensures we have a role available for the create test
	userRoleIDs := roleIDs[:len(roleIDs)-1]

	// Include both users for permissions
	userIDs := make(uuid.UUIDs, 2)
	userIDs[0] = tu1.ID
	userIDs[1] = tu2.ID

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, userRoleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles : %w", err)
	}

	// Seed table access for ALL roles (including the unassigned one)
	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access : %w", err)
	}

	allTAs, err := busDomain.TableAccess.QueryAll(ctx)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying all table access : %w", err)
	}

	// Only modify tu1's permissions for the table_access table
	ur1, err := busDomain.UserRole.QueryByUserID(ctx, tu1.ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying user1 roles : %w", err)
	}

	usrRoleIDs := make(uuid.UUIDs, len(ur1))
	for i, r := range ur1 {
		usrRoleIDs[i] = r.RoleID
	}

	tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying table access : %w", err)
	}

	countCheck, err := busDomain.TableAccess.Count(ctx, tableaccessbus.QueryFilter{})
	if err != nil {
		fmt.Println("Error counting table access:", countCheck)
		return apitest.SeedData{}, fmt.Errorf("counting table access : %w", err)
	}

	// Update only tu1's role permissions - set CanRead to TRUE!
	for _, ta := range tas {
		if ta.TableName == tableaccessapi.RouteTable {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(false),
			}
			_, err := busDomain.TableAccess.Update(ctx, ta, update)
			if err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access : %w", err)
			}
		}
	}

	appRoles := roleapp.ToAppRoles(roles)

	return apitest.SeedData{
		Users:         []apitest.User{tu1},
		Admins:        []apitest.User{tu2},
		Roles:         appRoles,
		TableAccesses: tableaccessapp.ToAppTableAccesses(allTAs),
	}, nil
}

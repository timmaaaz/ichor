package scenarioapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/scenarios/scenarioapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// insertSeedData stages everything the scenarios integration tests need:
//   - one regular user whose inventory.scenarios table_access is downgraded
//     to 0 perms so 403 PermissionDenied tests fire reliably
//   - one admin user with full access
//   - three scenario rows for query/queryByID coverage
//
// Mirrors the labelapi seed_test pattern — the role downgrade loop is what
// makes 403 tests actually return 403 for the non-admin token.
func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(busDomain.User, ath, admins[0].Email.Address),
	}

	scenarios, err := scenariobus.TestSeedScenarios(ctx, 3, busDomain.Scenario)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding scenarios: %w", err)
	}

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make([]uuid.UUID, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	if _, err = userrolebus.TestSeedUserRoles(ctx, []uuid.UUID{tu1.ID, tu2.ID}, roleIDs, busDomain.UserRole); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	if _, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	ur1, err := busDomain.UserRole.QueryByUserID(ctx, tu1.ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying user1 roles: %w", err)
	}

	usrRoleIDs := make(uuid.UUIDs, len(ur1))
	for i, r := range ur1 {
		usrRoleIDs[i] = r.RoleID
	}

	tas, err := busDomain.TableAccess.QueryByRoleIDs(ctx, usrRoleIDs)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying table access: %w", err)
	}

	for _, ta := range tas {
		if ta.TableName == scenarioapi.RouteTable {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(false),
			}
			if _, err := busDomain.TableAccess.Update(ctx, ta, update); err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access: %w", err)
			}
		}
	}

	return apitest.SeedData{
		Admins:    []apitest.User{tu2},
		Users:     []apitest.User{tu1},
		Scenarios: scenarios,
	}, nil
}

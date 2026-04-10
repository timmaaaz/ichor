package userpreferencesapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Seed admin user
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin users: %w", err)
	}
	admin := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// Seed regular user
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding regular users: %w", err)
	}
	user := apitest.User{
		User:  users[0],
		Token: apitest.Token(db.BusDomain.User, ath, users[0].Email.Address),
	}

	// =========================================================================
	// Permissions setup
	// =========================================================================
	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make(uuid.UUIDs, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	userIDs := uuid.UUIDs{admin.ID, user.ID}

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	// =========================================================================
	// Seed user preferences
	// =========================================================================
	_, err = userpreferencesbus.TestSeedUserPreferences(ctx, uuid.UUIDs{admin.ID}, busDomain.UserPreferences)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user preferences: %w", err)
	}

	return apitest.SeedData{
		Admins: []apitest.User{admin},
		Users:  []apitest.User{user},
	}, nil
}

// toAppUserPreference is available if tests need to convert seeded bus
// preferences to app-layer expected values.
func toAppUserPreference(bus userpreferencesbus.UserPreference) userpreferencesapp.UserPreference {
	return userpreferencesapp.ToAppUserPreference(bus)
}

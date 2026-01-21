package currencyapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/currencyapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Seed admin user
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin users: %w", err)
	}
	admin := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
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

	userIDs := uuid.UUIDs{admin.ID}

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	// =========================================================================
	// Seed Currencies
	// =========================================================================
	currencies, err := currencybus.TestSeedCurrencies(ctx, 2, busDomain.Currency)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding currencies: %w", err)
	}

	return apitest.SeedData{
		Admins:     []apitest.User{admin},
		Currencies: currencies,
	}, nil
}

func toAppCurrency(currency currencybus.Currency) currencyapp.Currency {
	return currencyapp.ToAppCurrency(currency)
}

func toAppCurrencies(currencies []currencybus.Currency) []currencyapp.Currency {
	return currencyapp.ToAppCurrencies(currencies)
}

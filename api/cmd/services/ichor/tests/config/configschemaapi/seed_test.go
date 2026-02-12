package configschema_test

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// SchemaSeedData holds test data for config schema API tests.
type SchemaSeedData struct {
	apitest.SeedData
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (SchemaSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return SchemaSeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	return SchemaSeedData{
		SeedData: apitest.SeedData{
			Users: []apitest.User{tu1},
		},
	}, nil
}

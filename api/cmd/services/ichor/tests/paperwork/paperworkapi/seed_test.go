package paperworkapi_test

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// insertSeedData stages a single admin user with a valid bearer token. The
// paperwork B2 scaffold has no Authorize middleware and no DB-persisted
// entities, so this seed is intentionally minimal — no role downgrade, no
// per-domain rows. Phase 0g.B3 will expand this when handlers do real work.
func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admins: %w", err)
	}

	tu := apitest.User{
		User:  admins[0],
		Token: apitest.Token(busDomain.User, ath, admins[0].Email.Address),
	}

	return apitest.SeedData{
		Admins: []apitest.User{tu},
	}, nil
}

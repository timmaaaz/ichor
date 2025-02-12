package reportsto_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/reportstoapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/reportstobus"
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

	// ============= User Creation =================

	reporters, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding reporter : %w", err)
	}

	reporterIDs := make([]uuid.UUID, len(reporters))
	for i, r := range reporters {
		reporterIDs[i] = r.ID
	}

	bosses, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding reporter : %w", err)
	}

	bossIDs := make([]uuid.UUID, len(bosses))
	for i, b := range bosses {
		bossIDs[i] = b.ID
	}

	// ============= ReportsTo Creation =================

	reportsTo, err := reportstobus.TestSeedReportsTo(ctx, 20, reporterIDs, bossIDs, busDomain.ReportsTo)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding reportsto : %w", err)
	}

	return apitest.SeedData{
		Admins:    []apitest.User{tu2},
		Users:     []apitest.User{tu1},
		ReportsTo: reportstoapp.ToAppReportsTos(reportsTo),
	}, nil
}

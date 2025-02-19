package brandapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfoapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/brandapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
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

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin : %w", err)
	}

	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	contacts, err := contactinfobus.TestSeedContactInfo(ctx, 10, busDomain.ContactInfo)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 25, contactIDs, busDomain.Brand)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding brands : %w", err)
	}

	return apitest.SeedData{
		Admins:      []apitest.User{tu2},
		Users:       []apitest.User{tu1},
		ContactInfo: contactinfoapp.ToAppContactInfos(contacts),
		Brands:      brandapp.ToAppBrands(brands),
	}, nil
}

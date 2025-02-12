package validasset_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/assettypeapp"
	"github.com/timmaaaz/ichor/app/domain/validassetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"

	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/validassetbus"

	"github.com/timmaaaz/ichor/business/domain/userbus"
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

	ats, err := assettypebus.TestSeedAssetTypes(ctx, 10, busDomain.AssetType)
	if err != nil {
		return apitest.SeedData{}, err
	}
	atIDs := make([]uuid.UUID, 0, len(ats))
	for _, at := range ats {
		atIDs = append(atIDs, at.ID)
	}

	as, err := validassetbus.TestSeedValidAssets(ctx, 20, atIDs, tu1.ID, busDomain.ValidAsset)
	if err != nil {
		return apitest.SeedData{}, err
	}

	sd := apitest.SeedData{
		Users:       []apitest.User{tu1},
		Admins:      []apitest.User{tu2},
		ValidAssets: validassetapp.ToAppValidAssets(as),
		AssetTypes:  assettypeapp.ToAppAssetTypes(ats),
	}

	return sd, nil
}

package asset_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assetapp"
	"github.com/timmaaaz/ichor/app/domain/assetconditionapp"
	"github.com/timmaaaz/ichor/app/domain/assettypeapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"

	"github.com/timmaaaz/ichor/business/domain/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsers(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}
	usrs, err = userbus.TestSeedUsers(ctx, 1, userbus.Roles.Admin, busDomain.User)
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

	acs, err := assetconditionbus.TestSeedAssetConditions(ctx, 6, busDomain.AssetCondition)
	if err != nil {
		return apitest.SeedData{}, err
	}
	acIDs := make([]uuid.UUID, 0, len(acs))
	for _, ac := range acs {
		acIDs = append(acIDs, ac.ID)
	}

	as, err := assetbus.TestSeedAssets(ctx, 20, atIDs, acIDs, tu1.ID, busDomain.Asset)
	if err != nil {
		return apitest.SeedData{}, err
	}

	sd := apitest.SeedData{
		Users:           []apitest.User{tu1},
		Admins:          []apitest.User{tu2},
		Assets:          assetapp.ToAppAssets(as),
		AssetConditions: assetconditionapp.ToAppAssetConditions(acs),
		AssetTypes:      assettypeapp.ToAppAssetTypes(ats),
	}

	return sd, nil
}

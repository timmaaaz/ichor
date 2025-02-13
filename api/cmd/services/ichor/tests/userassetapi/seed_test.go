package userasset_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/userassetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"

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

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	types, err := assettypebus.TestSeedAssetTypes(ctx, 3, busDomain.AssetType)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding asset types : %w", err)
	}

	typeIDs := make([]uuid.UUID, 0, len(types))
	for _, t := range types {
		typeIDs = append(typeIDs, t.ID)
	}

	conditions, err := assetconditionbus.TestSeedAssetConditions(ctx, 5, busDomain.AssetCondition)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding asset conditions : %w", err)
	}

	conditionIDs := make([]uuid.UUID, 0, len(conditions))
	for _, c := range conditions {
		conditionIDs = append(conditionIDs, c.ID)
	}

	validAssets, err := validassetbus.TestSeedValidAssets(ctx, 25, typeIDs, admins[0].ID, busDomain.ValidAsset)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding assets : %w", err)
	}

	validAssetIDs := make([]uuid.UUID, len(validAssets))
	for i, a := range validAssets {
		validAssetIDs[i] = a.ID
	}

	assets, err := assetbus.TestSeedAssets(ctx, 15, validAssetIDs, conditionIDs, busDomain.Asset)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding assets : %w", err)
	}

	assetIDs := make([]uuid.UUID, len(assets))
	for i, a := range assets {
		assetIDs[i] = a.ID
	}

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 20, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding approved by : %w", err)
	}

	userIDs := make(uuid.UUIDs, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	approvalStatuses, err := approvalstatusbus.TestSeedApprovalStatus(ctx, 12, busDomain.ApprovalStatus)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding approval statuses : %w", err)
	}

	approvalStatusIDs := make([]uuid.UUID, len(approvalStatuses))
	for i, as := range approvalStatuses {
		approvalStatusIDs[i] = as.ID
	}

	fulfillmentStatuses, err := fulfillmentstatusbus.TestSeedFulfillmentStatus(ctx, 8, busDomain.FulfillmentStatus)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding fulfillment statuses : %w", err)
	}

	fulfillmentStatusIDs := make([]uuid.UUID, len(fulfillmentStatuses))
	for i, fs := range fulfillmentStatuses {
		fulfillmentStatusIDs[i] = fs.ID
	}

	userAssets, err := userassetbus.TestSeedUserAssets(ctx, 25, userIDs[:15], assetIDs, userIDs[15:], conditionIDs, approvalStatusIDs, fulfillmentStatusIDs, busDomain.UserAsset)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user assets : %w", err)
	}

	return apitest.SeedData{
		Users:      []apitest.User{tu1},
		Admins:     []apitest.User{tu2},
		UserAssets: userassetapp.ToAppUserAssets(userAssets),
	}, nil
}

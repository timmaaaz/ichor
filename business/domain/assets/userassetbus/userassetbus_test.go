package userassetbus_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_UserAsset(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_UserAsset")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	types, err := assettypebus.TestSeedAssetTypes(ctx, 3, busDomain.AssetType)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset types : %w", err)
	}

	typeIDs := make([]uuid.UUID, 0, len(types))
	for _, t := range types {
		typeIDs = append(typeIDs, t.ID)
	}

	conditions, err := assetconditionbus.TestSeedAssetConditions(ctx, 5, busDomain.AssetCondition)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset conditions : %w", err)
	}

	conditionIDs := make([]uuid.UUID, 0, len(conditions))
	for _, c := range conditions {
		conditionIDs = append(conditionIDs, c.ID)
	}

	validAssets, err := validassetbus.TestSeedValidAssets(ctx, 25, typeIDs, admins[0].ID, busDomain.ValidAsset)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding valid assets : %w", err)
	}

	validAssetIDs := make([]uuid.UUID, len(validAssets))
	for i, a := range validAssets {
		validAssetIDs[i] = a.ID
	}

	assets, err := assetbus.TestSeedAssets(ctx, 15, validAssetIDs, conditionIDs, busDomain.Asset)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding assets : %w", err)
	}

	assetIDs := make([]uuid.UUID, len(assets))
	for i, a := range assets {
		assetIDs[i] = a.ID
	}

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 20, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding approved by : %w", err)
	}

	userIDs := make(uuid.UUIDs, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	approvalStatuses, err := approvalstatusbus.TestSeedApprovalStatus(ctx, 12, busDomain.ApprovalStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding approval statuses : %w", err)
	}

	approvalStatusIDs := make([]uuid.UUID, len(approvalStatuses))
	for i, as := range approvalStatuses {
		approvalStatusIDs[i] = as.ID
	}

	fulfillmentStatuses, err := fulfillmentstatusbus.TestSeedFulfillmentStatus(ctx, 8, busDomain.FulfillmentStatus)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding fulfillment statuses : %w", err)
	}

	fulfillmentStatusIDs := make([]uuid.UUID, len(fulfillmentStatuses))
	for i, fs := range fulfillmentStatuses {
		fulfillmentStatusIDs[i] = fs.ID
	}

	userAssets, err := userassetbus.TestSeedUserAssets(ctx, 25, userIDs[:15], assetIDs, userIDs[15:], conditionIDs, approvalStatusIDs, fulfillmentStatusIDs, busDomain.UserAsset)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user assets : %w", err)
	}

	return unitest.SeedData{
		Admins:          []unitest.User{{User: admins[0]}},
		AssetTypes:      types,
		AssetConditions: conditions,
		ValidAssets:     validAssets,
		UserAssets:      userAssets,
		Assets:          assets,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []userassetbus.UserAsset{
				sd.UserAssets[0],
				sd.UserAssets[1],
				sd.UserAssets[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.UserAsset.Query(ctx, userassetbus.QueryFilter{}, userassetbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.([]userassetbus.UserAsset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]userassetbus.UserAsset)
				if !exists {
					return "error occurred"
				}

				for i := range gotResp {
					expResp[i].ID = gotResp[i].ID
				}

				return cmp.Diff(expResp, gotResp, cmpopts.EquateApproxTime(time.Second))
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Create",
			ExpResp: sd.UserAssets[0],
			ExcFunc: func(ctx context.Context) any {
				na := userassetbus.NewUserAsset{
					UserID:              sd.UserAssets[0].UserID,
					AssetID:             sd.UserAssets[0].AssetID,
					ApprovedBy:          sd.UserAssets[0].ApprovedBy,
					ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
					FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
					DateReceived:        sd.UserAssets[0].DateReceived,
					LastMaintenance:     sd.UserAssets[0].LastMaintenance,
				}

				got, err := busDomain.UserAsset.Create(ctx, na)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(userassetbus.UserAsset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(userassetbus.UserAsset)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: userassetbus.UserAsset{
				ID:                  sd.UserAssets[1].ID,
				UserID:              sd.UserAssets[0].UserID,
				AssetID:             sd.UserAssets[0].AssetID,
				ApprovedBy:          sd.UserAssets[0].ApprovedBy,
				ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[0].DateReceived,
				LastMaintenance:     sd.UserAssets[0].LastMaintenance,
			},
			ExcFunc: func(ctx context.Context) any {
				ua := userassetbus.UpdateUserAsset{
					UserID:              &sd.UserAssets[0].UserID,
					AssetID:             &sd.UserAssets[0].AssetID,
					ApprovedBy:          &sd.UserAssets[0].ApprovedBy,
					ApprovalStatusID:    &sd.UserAssets[0].ApprovalStatusID,
					FulfillmentStatusID: &sd.UserAssets[0].FulfillmentStatusID,
					DateReceived:        &sd.UserAssets[0].DateReceived,
					LastMaintenance:     &sd.UserAssets[0].LastMaintenance,
				}

				got, err := busDomain.UserAsset.Update(ctx, sd.UserAssets[1], ua)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(userassetbus.UserAsset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(userassetbus.UserAsset)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.UserAsset.Delete(ctx, sd.UserAssets[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

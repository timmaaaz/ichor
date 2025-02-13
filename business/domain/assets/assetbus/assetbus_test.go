package assetbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Asset(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Asset")

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

	assetTypes, err := assettypebus.TestSeedAssetTypes(ctx, 5, busDomain.AssetType)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset types : %w", err)
	}

	assetTypeIDs := make([]uuid.UUID, len(assetTypes))
	for i, at := range assetTypes {
		assetTypeIDs[i] = at.ID
	}

	validAssets, err := validassetbus.TestSeedValidAssets(ctx, 15, assetTypeIDs, admins[0].ID, busDomain.ValidAsset)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding valid assets : %w", err)
	}

	validAssetIDs := make([]uuid.UUID, len(validAssets))
	for i, va := range validAssets {
		validAssetIDs[i] = va.ID
	}

	conditions, err := assetconditionbus.TestSeedAssetConditions(ctx, 8, busDomain.AssetCondition)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset conditions : %w", err)
	}

	conditionIDs := make([]uuid.UUID, len(conditions))
	for i, c := range conditions {
		conditionIDs[i] = c.ID
	}

	assets, err := assetbus.TestSeedAssets(ctx, 15, validAssetIDs, conditionIDs, busDomain.Asset)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding assets : %w", err)
	}

	return unitest.SeedData{
		Admins: []unitest.User{{User: admins[0]}},
		Assets: assets,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []assetbus.Asset{
				sd.Assets[0],
				sd.Assets[1],
				sd.Assets[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Asset.Query(ctx, assetbus.QueryFilter{}, assetbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.([]assetbus.Asset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]assetbus.Asset)
				if !exists {
					return "error occurred"
				}

				for i := range gotResp {
					expResp[i].ID = gotResp[i].ID

				}

				return cmp.Diff(exp, got)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: assetbus.Asset{
				SerialNumber:     "123456",
				ValidAssetID:     sd.Assets[0].ValidAssetID,
				AssetConditionID: sd.Assets[0].AssetConditionID,
				LastMaintenance:  sd.Assets[0].LastMaintenance,
			},
			ExcFunc: func(ctx context.Context) any {
				na := assetbus.NewAsset{
					SerialNumber:     "123456",
					ValidAssetID:     sd.Assets[0].ValidAssetID,
					AssetConditionID: sd.Assets[0].AssetConditionID,
					LastMaintenance:  sd.Assets[0].LastMaintenance,
				}

				got, err := busDomain.Asset.Create(ctx, na)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(assetbus.Asset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(assetbus.Asset)
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
			ExpResp: assetbus.Asset{
				ID:               sd.Assets[0].ID,
				SerialNumber:     "654321",
				ValidAssetID:     sd.Assets[1].ValidAssetID,
				AssetConditionID: sd.Assets[1].AssetConditionID,
				LastMaintenance:  sd.Assets[1].LastMaintenance,
			},
			ExcFunc: func(ctx context.Context) any {
				ua := assetbus.UpdateAsset{
					ValidAssetID:     &sd.Assets[1].ValidAssetID,
					SerialNumber:     dbtest.StringPointer("654321"),
					AssetConditionID: &sd.Assets[1].AssetConditionID,
					LastMaintenance:  &sd.Assets[1].LastMaintenance,
				}

				got, err := busDomain.Asset.Update(ctx, sd.Assets[0], ua)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(assetbus.Asset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(assetbus.Asset)
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
				err := busDomain.Asset.Delete(ctx, sd.Assets[0])
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

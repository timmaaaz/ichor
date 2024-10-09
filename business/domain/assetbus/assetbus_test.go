package assetbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/userbus"
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

	admins, err := userbus.TestSeedUsers(ctx, 1, userbus.Roles.Admin, busDomain.User)
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

	conditions, err := assetconditionbus.TestSeedAssetConditions(ctx, 3, busDomain.AssetCondition)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset conditions : %w", err)
	}
	conditionIDs := make([]uuid.UUID, 0, len(conditions))
	for _, c := range conditions {
		conditionIDs = append(conditionIDs, c.ID)
	}

	assets, err := assetbus.TestSeedAssets(ctx, 10, typeIDs, conditionIDs, admins[0].ID, busDomain.Asset)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding assets : %w", err)
	}

	return unitest.SeedData{
		Admins:          []unitest.User{unitest.User{User: admins[0]}},
		AssetTypes:      types,
		AssetConditions: conditions,
		Assets:          assets,
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
					expResp[i].TypeID = gotResp[i].TypeID
					expResp[i].ConditionID = gotResp[i].ConditionID
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
				TypeID:      sd.AssetTypes[0].ID,
				ConditionID: sd.AssetConditions[0].ID,
				Name:        "Test Asset",
				ModelNumber: "654321",
				IsEnabled:   true,
				CreatedBy:   sd.Admins[0].ID,
				UpdatedBy:   sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				na := assetbus.NewAsset{
					TypeID:      sd.AssetTypes[0].ID,
					ConditionID: sd.AssetConditions[0].ID,
					Name:        "Test Asset",
					ModelNumber: "654321",
					IsEnabled:   true,
					CreatedBy:   sd.Admins[0].ID,
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
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

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
				ID:                  sd.Assets[1].ID,
				TypeID:              sd.Assets[1].TypeID,
				ConditionID:         sd.Assets[1].ConditionID,
				Name:                "Updated Asset",
				EstPrice:            sd.Assets[1].EstPrice,
				Price:               sd.Assets[1].Price,
				MaintenanceInterval: sd.Assets[1].MaintenanceInterval,
				LifeExpectancy:      sd.Assets[1].LifeExpectancy,
				ModelNumber:         "123456",
				IsEnabled:           false,
				CreatedBy:           sd.Admins[0].ID,
				UpdatedBy:           sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				ua := assetbus.UpdateAsset{
					Name:        dbtest.StringPointer("Updated Asset"),
					ModelNumber: dbtest.StringPointer("123456"),
					IsEnabled:   dbtest.BoolPointer(false),
				}

				got, err := busDomain.Asset.Update(ctx, sd.Assets[1], ua)
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

				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

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

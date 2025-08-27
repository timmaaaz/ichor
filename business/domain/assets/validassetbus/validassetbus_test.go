package validassetbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ValidAsset(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ValidAsset")

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

	assets, err := validassetbus.TestSeedValidAssets(ctx, 10, typeIDs, admins[0].ID, busDomain.ValidAsset)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding assets : %w", err)
	}

	return unitest.SeedData{
		Admins:      []unitest.User{{User: admins[0]}},
		AssetTypes:  types,
		ValidAssets: assets,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []validassetbus.ValidAsset{
				sd.ValidAssets[0],
				sd.ValidAssets[1],
				sd.ValidAssets[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.ValidAsset.Query(ctx, validassetbus.QueryFilter{}, validassetbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.([]validassetbus.ValidAsset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]validassetbus.ValidAsset)
				if !exists {
					return "error occurred"
				}

				for i := range gotResp {
					expResp[i].ID = gotResp[i].ID
					expResp[i].TypeID = gotResp[i].TypeID
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
			ExpResp: validassetbus.ValidAsset{
				TypeID:       sd.AssetTypes[0].ID,
				Name:         "Test Asset",
				SerialNumber: "123456",
				ModelNumber:  "654321",
				IsEnabled:    true,
				CreatedBy:    sd.Admins[0].ID,
				UpdatedBy:    sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				na := validassetbus.NewValidAsset{
					TypeID:       sd.AssetTypes[0].ID,
					Name:         "Test Asset",
					SerialNumber: "123456",
					ModelNumber:  "654321",
					IsEnabled:    true,
					CreatedBy:    sd.Admins[0].ID,
				}

				got, err := busDomain.ValidAsset.Create(ctx, na)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(validassetbus.ValidAsset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(validassetbus.ValidAsset)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: validassetbus.ValidAsset{
				ID:                  sd.ValidAssets[1].ID,
				TypeID:              sd.ValidAssets[1].TypeID,
				Name:                "Updated Asset",
				EstPrice:            sd.ValidAssets[1].EstPrice,
				Price:               sd.ValidAssets[1].Price,
				MaintenanceInterval: sd.ValidAssets[1].MaintenanceInterval,
				LifeExpectancy:      sd.ValidAssets[1].LifeExpectancy,
				SerialNumber:        "654321",
				ModelNumber:         "123456",
				IsEnabled:           false,
				CreatedBy:           sd.Admins[0].ID,
				UpdatedBy:           sd.Admins[0].ID,
			},
			ExcFunc: func(ctx context.Context) any {
				ua := validassetbus.UpdateValidAsset{
					Name:         dbtest.StringPointer("Updated Asset"),
					SerialNumber: dbtest.StringPointer("654321"),
					ModelNumber:  dbtest.StringPointer("123456"),
					IsEnabled:    dbtest.BoolPointer(false),
				}

				got, err := busDomain.ValidAsset.Update(ctx, sd.ValidAssets[1], ua)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(validassetbus.ValidAsset)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(validassetbus.ValidAsset)
				if !exists {
					return "error occurred"
				}

				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

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
				err := busDomain.ValidAsset.Delete(ctx, sd.ValidAssets[0])
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

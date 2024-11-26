package assettypebus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_AssetType(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_AssetType")

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

	at, err := assettypebus.TestSeedAssetType(ctx, 10, busDomain.AssetType)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset types : %w", err)
	}

	return unitest.SeedData{
		AssetType: at,
	}, nil
}

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []assettypebus.AssetType{
				{ID: sd.AssetType[0].ID, Name: sd.AssetType[0].Name},
				{ID: sd.AssetType[1].ID, Name: sd.AssetType[1].Name},
				{ID: sd.AssetType[2].ID, Name: sd.AssetType[2].Name},
				{ID: sd.AssetType[3].ID, Name: sd.AssetType[3].Name},
				{ID: sd.AssetType[4].ID, Name: sd.AssetType[4].Name},
			},
			ExcFunc: func(ctx context.Context) any {
				assetTypes, err := busdomain.AssetType.Query(ctx, assettypebus.QueryFilter{}, order.NewBy(assettypebus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return assetTypes
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, _ unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "create",
			ExpResp: assettypebus.AssetType{
				Name: "Test Approval Status",
			},
			ExcFunc: func(ctx context.Context) any {
				assetTypes, err := busDomain.AssetType.Create(ctx, assettypebus.NewAssetType{
					Name: "Test Approval Status",
				})
				if err != nil {
					return err
				}
				return assetTypes
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(assettypebus.AssetType)
				if !exists {
					return fmt.Sprintf("got is not an asset type %v", got)
				}

				expResp := exp.(assettypebus.AssetType)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: assettypebus.AssetType{
				ID:   sd.AssetType[0].ID,
				Name: "Updated Approval Status",
			},
			ExcFunc: func(ctx context.Context) any {
				assetType, err := busDomain.AssetType.Update(ctx, sd.AssetType[0], assettypebus.UpdateAssetType{
					Name: dbtest.StringPointer("Updated Approval Status"),
				})
				if err != nil {
					return err
				}
				return assetType
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(assettypebus.AssetType)
				if !exists {
					return fmt.Sprintf("got is not an asset type %v", got)
				}
				expResp := exp.(assettypebus.AssetType)
				return cmp.Diff(gotResp, expResp)
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
				err := busDomain.AssetType.Delete(ctx, sd.AssetType[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

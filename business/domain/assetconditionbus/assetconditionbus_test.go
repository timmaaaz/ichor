package assetconditionbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_AssetCondition(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_AssetCondition")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	ats, err := assetconditionbus.TestSeedAssetConditions(ctx, 10, busDomain.AssetCondition)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset conditions : %w", err)
	}

	return unitest.SeedData{
		AssetConditions: ats,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Query",
			ExpResp: []assetconditionbus.AssetCondition{
				sd.AssetConditions[0],
				sd.AssetConditions[1],
				sd.AssetConditions[2],
				sd.AssetConditions[3],
				sd.AssetConditions[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.AssetCondition.Query(ctx, assetconditionbus.QueryFilter{}, assetconditionbus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(exp, got)
			},
		},
		{
			Name:    "Query by id",
			ExpResp: sd.AssetConditions[0],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.AssetCondition.QueryByID(ctx, sd.AssetConditions[0].ID)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(exp, got)
			},
		},
	}
	return table
}

func create(busDomain dbtest.BusDomain) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Create",
			ExpResp: assetconditionbus.AssetCondition{
				Name:        "Test AssetCondition",
				Description: "Test AssetCondition Description",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.AssetCondition.Create(ctx, assetconditionbus.NewAssetCondition{
					Name:        "Test AssetCondition",
					Description: "Test AssetCondition Description",
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(assetconditionbus.AssetCondition)
				if !exists {
					return fmt.Sprintf("got is not an asset condition: %v", got)
				}

				expResp := exp.(assetconditionbus.AssetCondition)
				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Update",
			ExpResp: assetconditionbus.AssetCondition{
				ID:          sd.AssetConditions[0].ID,
				Name:        "Updated AssetCondition",
				Description: "Updated AssetCondition Description",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.AssetCondition.Update(ctx, sd.AssetConditions[0], assetconditionbus.UpdateAssetCondition{
					Name:        dbtest.StringPointer("Updated AssetCondition"),
					Description: dbtest.StringPointer("Updated AssetCondition Description"),
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(assetconditionbus.AssetCondition)
				if !exists {
					return fmt.Sprintf("got is not an asset condition: %v", got)
				}

				expResp := exp.(assetconditionbus.AssetCondition)
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.AssetCondition.Delete(ctx, sd.AssetConditions[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
	return table
}

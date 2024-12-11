package assettypebus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
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
	unitest.Run(t, create(db.BusDomain), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	ats, err := assettypebus.TestSeedAssetTypes(ctx, 10, busDomain.AssetType)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset types : %w", err)
	}

	return unitest.SeedData{
		AssetTypes: ats,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "Query",
			ExpResp: []assettypebus.AssetType{
				sd.AssetTypes[0],
				sd.AssetTypes[1],
				sd.AssetTypes[2],
				sd.AssetTypes[3],
				sd.AssetTypes[4],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.AssetType.Query(ctx, assettypebus.QueryFilter{}, assettypebus.DefaultOrderBy, page.MustParse("1", "5"))
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
			ExpResp: sd.AssetTypes[0],
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.AssetType.QueryByID(ctx, sd.AssetTypes[0].ID)
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
			ExpResp: assettypebus.AssetType{
				Name:        "Test AssetType",
				Description: "Test AssetType Description",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.AssetType.Create(ctx, assettypebus.NewAssetType{
					Name:        "Test AssetType",
					Description: "Test AssetType Description",
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(assettypebus.AssetType)
				if !exists {
					return fmt.Sprintf("got is not an asset type: %v", got)
				}

				expResp := exp.(assettypebus.AssetType)
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
			ExpResp: assettypebus.AssetType{
				ID:          sd.AssetTypes[0].ID,
				Name:        "Updated AssetType",
				Description: "Updated AssetType Description",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.AssetType.Update(ctx, sd.AssetTypes[0], assettypebus.UpdateAssetType{
					Name:        dbtest.StringPointer("Updated AssetType"),
					Description: dbtest.StringPointer("Updated AssetType Description"),
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(assettypebus.AssetType)
				if !exists {
					return fmt.Sprintf("got is not an asset type: %v", got)
				}

				expResp := exp.(assettypebus.AssetType)
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
				err := busDomain.AssetType.Delete(ctx, sd.AssetTypes[0])
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

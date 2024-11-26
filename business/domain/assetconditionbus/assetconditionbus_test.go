package assetconditionbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
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
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	as, err := assetconditionbus.TestSeedAssetCondition(ctx, 10, busDomain.AssetCondition)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset condition : %w", err)
	}

	return unitest.SeedData{
		AssetCondition: as,
	}, nil
}

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []assetconditionbus.AssetCondition{
				{ID: sd.AssetCondition[0].ID, Name: sd.AssetCondition[0].Name},
				{ID: sd.AssetCondition[1].ID, Name: sd.AssetCondition[1].Name},
				{ID: sd.AssetCondition[2].ID, Name: sd.AssetCondition[2].Name},
				{ID: sd.AssetCondition[3].ID, Name: sd.AssetCondition[3].Name},
				{ID: sd.AssetCondition[4].ID, Name: sd.AssetCondition[4].Name},
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatuses, err := busdomain.AssetCondition.Query(ctx, assetconditionbus.QueryFilter{}, order.NewBy(assetconditionbus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return aprvlStatuses
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
			ExpResp: assetconditionbus.AssetCondition{
				Name: "Test Asset Condition",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.AssetCondition.Create(ctx, assetconditionbus.NewAssetCondition{
					Name: "Test Asset Condition",
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(assetconditionbus.AssetCondition)
				if !exists {
					return fmt.Sprintf("got is not an asset condition %v", got)
				}

				expResp := exp.(assetconditionbus.AssetCondition)
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
			ExpResp: assetconditionbus.AssetCondition{
				ID:   sd.AssetCondition[0].ID,
				Name: "Updated Asset Condition",
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.AssetCondition.Update(ctx, sd.AssetCondition[0], assetconditionbus.UpdateAssetCondition{
					Name: dbtest.StringPointer("Updated Asset Condition"),
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(assetconditionbus.AssetCondition)
				if !exists {
					return fmt.Sprintf("got is not an asset condition %v", got)
				}
				expResp := exp.(assetconditionbus.AssetCondition)
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
				err := busDomain.AssetCondition.Delete(ctx, sd.AssetCondition[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

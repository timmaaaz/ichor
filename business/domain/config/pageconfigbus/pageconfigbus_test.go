package pageconfigbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_PageConfig(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_PageConfig")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, queryByID(db.BusDomain, sd), "queryByID")
	unitest.Run(t, queryByName(db.BusDomain, sd), "queryByName")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	configs, err := pageconfigbus.TestSeedPageConfigs(ctx, 10, busDomain.PageConfig)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding page configs : %w", err)
	}

	return unitest.SeedData{
		PageConfigs: configs,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []pageconfigbus.PageConfig{
				{ID: sd.PageConfigs[0].ID, Name: sd.PageConfigs[0].Name, IsDefault: sd.PageConfigs[0].IsDefault},
				{ID: sd.PageConfigs[1].ID, Name: sd.PageConfigs[1].Name, IsDefault: sd.PageConfigs[1].IsDefault},
				{ID: sd.PageConfigs[2].ID, Name: sd.PageConfigs[2].Name, IsDefault: sd.PageConfigs[2].IsDefault},
				{ID: sd.PageConfigs[3].ID, Name: sd.PageConfigs[3].Name, IsDefault: sd.PageConfigs[3].IsDefault},
				{ID: sd.PageConfigs[4].ID, Name: sd.PageConfigs[4].Name, IsDefault: sd.PageConfigs[4].IsDefault},
			},
			ExcFunc: func(ctx context.Context) any {
				configs, err := busDomain.PageConfig.Query(ctx, pageconfigbus.QueryFilter{}, order.NewBy(pageconfigbus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return configs
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByID(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "queryByID",
			ExpResp: sd.PageConfigs[0],
			ExcFunc: func(ctx context.Context) any {
				config, err := busDomain.PageConfig.QueryByID(ctx, sd.PageConfigs[0].ID)
				if err != nil {
					return err
				}
				return config
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByName(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "queryByName",
			ExpResp: sd.PageConfigs[0],
			ExcFunc: func(ctx context.Context) any {
				config, err := busDomain.PageConfig.QueryByName(ctx, sd.PageConfigs[0].Name)
				if err != nil {
					return err
				}
				return config
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "create",
			ExpResp: pageconfigbus.PageConfig{
				Name:      "Test Page Config",
				IsDefault: true,
			},
			ExcFunc: func(ctx context.Context) any {
				config, err := busDomain.PageConfig.Create(ctx, pageconfigbus.NewPageConfig{
					Name:      "Test Page Config",
					IsDefault: true,
				})
				if err != nil {
					return err
				}
				return config
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(pageconfigbus.PageConfig)
				if !exists {
					return fmt.Sprintf("got is not a page config %v", got)
				}

				expResp := exp.(pageconfigbus.PageConfig)
				expResp.ID = gotResp.ID
				expResp.UserID = gotResp.UserID

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
			ExpResp: pageconfigbus.PageConfig{
				ID:        sd.PageConfigs[0].ID,
				Name:      "Updated Page Config",
				IsDefault: sd.PageConfigs[0].IsDefault,
			},
			ExcFunc: func(ctx context.Context) any {
				config, err := busDomain.PageConfig.Update(ctx, pageconfigbus.UpdatePageConfig{
					Name: dbtest.StringPointer("Updated Page Config"),
				}, sd.PageConfigs[0].ID)
				if err != nil {
					return err
				}
				return config
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(pageconfigbus.PageConfig)
				if !exists {
					return fmt.Sprintf("got is not a page config %v", got)
				}
				expResp := exp.(pageconfigbus.PageConfig)
				expResp.UserID = gotResp.UserID
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
				err := busDomain.PageConfig.Delete(ctx, sd.PageConfigs[0].ID)
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

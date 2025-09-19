package inventoryadjustmentapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/inventory-adjustments?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[inventoryadjustmentapp.InventoryAdjustment]{},
			ExpResp: &query.Result[inventoryadjustmentapp.InventoryAdjustment]{
				Page:        1,
				RowsPerPage: 10,
				Total:       10,
				Items:       sd.InventoryAdjustments[:10],
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/inventory-adjustments/" + sd.InventoryAdjustments[0].InventoryAdjustmentID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &inventoryadjustmentapp.InventoryAdjustment{},
			ExpResp:    &sd.InventoryAdjustments[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

package inventorytransactionapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/inventory/inventorytransactionapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/movement/inventory-transactions?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[inventorytransactionapp.InventoryTransaction]{},
			ExpResp: &query.Result[inventorytransactionapp.InventoryTransaction]{
				Page:        1,
				RowsPerPage: 10,
				Total:       25,
				Items:       sd.InventoryTransactions[:10],
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
			URL:        "/v1/movement/inventory-transactions/" + sd.InventoryTransactions[0].InventoryTransactionID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &inventorytransactionapp.InventoryTransaction{},
			ExpResp:    &sd.InventoryTransactions[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

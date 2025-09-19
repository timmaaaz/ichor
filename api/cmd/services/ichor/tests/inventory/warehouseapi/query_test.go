package warehouse_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/warehouseapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	expItems := make([]warehouseapp.Warehouse, len(sd.Warehouses))
	copy(expItems, sd.Warehouses)

	sort.Slice(expItems, func(i, j int) bool {
		return expItems[i].ID < expItems[j].ID
	})

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/warehouses?page=1&rows=3",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[warehouseapp.Warehouse]{},
			ExpResp: &query.Result[warehouseapp.Warehouse]{
				Page:        1,
				RowsPerPage: 3,
				Total:       5,
				Items:       expItems[:3],
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*query.Result[warehouseapp.Warehouse])
				if !ok {
					return "error occurred"
				}
				expResp, ok := exp.(*query.Result[warehouseapp.Warehouse])
				if !ok {
					return "error occurred"
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/warehouses/" + sd.Warehouses[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &warehouseapp.Warehouse{},
			ExpResp:    &sd.Warehouses[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

package ordersapi_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	expItems := make([]ordersapp.Order, len(sd.Orders))
	copy(expItems, sd.Orders)

	sort.Slice(expItems, func(i, j int) bool {
		return expItems[i].Number < expItems[j].Number
	})

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/sales/orders?page=1&rows=3",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[ordersapp.Order]{},
			ExpResp: &query.Result[ordersapp.Order]{
				Page:        1,
				RowsPerPage: 3,
				Total:       5,
				Items:       expItems[:3],
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*query.Result[ordersapp.Order])
				if !ok {
					return "error occurred"
				}
				expResp, ok := exp.(*query.Result[ordersapp.Order])
				if !ok {
					return "error occurred"
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/sales/orders/" + sd.Orders[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ordersapp.Order{},
			ExpResp:    &sd.Orders[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

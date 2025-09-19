package orderlineitemapi_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	expItems := make([]orderlineitemsapp.OrderLineItem, len(sd.OrderLineItems))
	copy(expItems, sd.OrderLineItems)

	sort.Slice(expItems, func(i, j int) bool {
		return expItems[i].ID < expItems[j].ID
	})

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/sales/order-line-items?page=1&rows=3",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[orderlineitemsapp.OrderLineItem]{},
			ExpResp: &query.Result[orderlineitemsapp.OrderLineItem]{
				Page:        1,
				RowsPerPage: 3,
				Total:       5,
				Items:       expItems[:3],
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*query.Result[orderlineitemsapp.OrderLineItem])
				if !ok {
					return "error occurred"
				}
				expResp, ok := exp.(*query.Result[orderlineitemsapp.OrderLineItem])
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
			URL:        "/v1/sales/order-line-items/" + sd.OrderLineItems[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &orderlineitemsapp.OrderLineItem{},
			ExpResp:    &sd.OrderLineItems[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

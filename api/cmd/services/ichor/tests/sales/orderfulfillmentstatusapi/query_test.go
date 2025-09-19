package orderfulfillmentstatusapi_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/orderfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	expItems := make([]orderfulfillmentstatusapp.OrderFulfillmentStatus, len(sd.OrderFulfillmentStatuses))
	copy(expItems, sd.OrderFulfillmentStatuses)

	sort.Slice(expItems, func(i, j int) bool {
		return expItems[i].Name < expItems[j].Name
	})

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/order/order-fulfillment-statuses?page=1&rows=3",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[orderfulfillmentstatusapp.OrderFulfillmentStatus]{},
			ExpResp: &query.Result[orderfulfillmentstatusapp.OrderFulfillmentStatus]{
				Page:        1,
				RowsPerPage: 3,
				Total:       5,
				Items:       expItems[:3],
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*query.Result[orderfulfillmentstatusapp.OrderFulfillmentStatus])
				if !ok {
					return "error occurred"
				}
				expResp, ok := exp.(*query.Result[orderfulfillmentstatusapp.OrderFulfillmentStatus])
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
			URL:        "/v1/order/order-fulfillment-statuses/" + sd.OrderFulfillmentStatuses[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &orderfulfillmentstatusapp.OrderFulfillmentStatus{},
			ExpResp:    &sd.OrderFulfillmentStatuses[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

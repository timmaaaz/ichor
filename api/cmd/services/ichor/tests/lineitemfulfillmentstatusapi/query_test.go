package lineitemfulfillmentstatusapi_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/order/lineitemfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	expItems := make([]lineitemfulfillmentstatusapp.LineItemFulfillmentStatus, len(sd.LineItemFulfillmentStatuses))
	copy(expItems, sd.LineItemFulfillmentStatuses)

	sort.Slice(expItems, func(i, j int) bool {
		return expItems[i].Name < expItems[j].Name
	})

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/order/line-item-fulfillment-statuses?page=1&rows=3",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[lineitemfulfillmentstatusapp.LineItemFulfillmentStatus]{},
			ExpResp: &query.Result[lineitemfulfillmentstatusapp.LineItemFulfillmentStatus]{
				Page:        1,
				RowsPerPage: 3,
				Total:       6,
				Items:       expItems[:3],
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*query.Result[lineitemfulfillmentstatusapp.LineItemFulfillmentStatus])
				if !ok {
					return "error occurred"
				}
				expResp, ok := exp.(*query.Result[lineitemfulfillmentstatusapp.LineItemFulfillmentStatus])
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
			URL:        "/v1/order/line-item-fulfillment-statuses/" + sd.LineItemFulfillmentStatuses[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &lineitemfulfillmentstatusapp.LineItemFulfillmentStatus{},
			ExpResp:    &sd.LineItemFulfillmentStatuses[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

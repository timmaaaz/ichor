package fulfillmentstatus_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/fulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/fulfillmentstatus?page=1&rows=2",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[fulfillmentstatusapp.FulfillmentStatus]{},
			ExpResp: &query.Result[fulfillmentstatusapp.FulfillmentStatus]{
				Page:        1,
				RowsPerPage: 2,
				Total:       5,
				Items:       sd.FulfillmentStatuses[:2],
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
			URL:        "/v1/fulfillmentstatus/" + sd.FulfillmentStatuses[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &fulfillmentstatusapp.FulfillmentStatus{},
			ExpResp:    &sd.FulfillmentStatuses[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

package reportsto_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/users/reportstoapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users/reportsto?page=1&rows=10",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[reportstoapp.ReportsTo]{},
			ExpResp: &query.Result[reportstoapp.ReportsTo]{
				Page:        1,
				RowsPerPage: 10,
				Total:       20,
				Items:       sd.ReportsTo[:10],
			},
			CmpFunc: func(got any, exp any) string {
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
			URL:        "/v1/users/reportsto/" + sd.ReportsTo[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &reportstoapp.ReportsTo{},
			ExpResp:    &sd.ReportsTo[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

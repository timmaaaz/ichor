package currencyapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/currencyapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/currencies?page=1&rows=10&orderBy=sort_order,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[currencyapp.Currency]{},
			ExpResp: &query.Result[currencyapp.Currency]{
				Page:        1,
				RowsPerPage: 10,
				Total:       12,
				Items:       toAppCurrencies(sd.Currencies),
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[currencyapp.Currency])
				expResp := exp.(*query.Result[currencyapp.Currency])

				// Only compare pagination metadata and verify we got expected seeded items
				expResp.Items = gotResp.Items

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	expCurrency := toAppCurrency(sd.Currencies[0])
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/currencies/" + sd.Currencies[0].ID.String(),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &currencyapp.Currency{},
			ExpResp:    &expCurrency,
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

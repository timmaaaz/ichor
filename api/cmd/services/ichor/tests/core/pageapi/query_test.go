package page_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/pages?page=2&rows=5",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[pageapp.Page]{},
			ExpResp: &query.Result[pageapp.Page]{
				Page:        2,
				RowsPerPage: 5,
				Total:       12,
				Items:       sd.Pages,
			},
			CmpFunc: func(got any, exp any) string {
				// Sort by sortOrder
				items := exp.(*query.Result[pageapp.Page]).Items
				sort.Slice(items, func(i, j int) bool {
					return items[i].SortOrder < items[j].SortOrder
				})
				// Grab the items for page 2
				exp.(*query.Result[pageapp.Page]).Items = items[5:10]

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
			URL:        "/v1/core/pages/" + sd.Pages[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &pageapp.Page{},
			ExpResp:    &sd.Pages[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func query401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/pages?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &query.Result[pageapp.Page]{},
			ExpResp:    &query.Result[pageapp.Page]{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

package role_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/roleapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/permissions/roles?page=2&rows=5",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[roleapp.Role]{},
			ExpResp: &query.Result[roleapp.Role]{
				Page:        2,
				RowsPerPage: 5,
				Total:       13,
				Items:       sd.Roles,
			},
			CmpFunc: func(got any, exp any) string {

				// Sort by name
				items := exp.(*query.Result[roleapp.Role]).Items
				sort.Slice(items, func(i, j int) bool {
					return items[i].Name < items[j].Name
				})
				// Grab the first 10
				exp.(*query.Result[roleapp.Role]).Items = items[5:10]

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
			URL:        "/v1/permissions/roles/" + sd.Roles[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &roleapp.Role{},
			ExpResp:    &sd.Roles[0],
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
			URL:        "/v1/permissions/roles?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &query.Result[roleapp.Role]{},
			ExpResp:    &query.Result[roleapp.Role]{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

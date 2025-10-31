package rolepage_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/rolepageapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/role-pages?page=1&rows=10",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[rolepageapp.RolePage]{},
			ExpResp: &query.Result[rolepageapp.RolePage]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sd.RolePages),
				Items:       sd.RolePages,
			},
			CmpFunc: func(got any, exp any) string {
				// Sort by ID for consistent comparison
				gotResp := got.(*query.Result[rolepageapp.RolePage])
				expResp := exp.(*query.Result[rolepageapp.RolePage])

				sort.Slice(gotResp.Items, func(i, j int) bool {
					return gotResp.Items[i].ID < gotResp.Items[j].ID
				})
				sort.Slice(expResp.Items, func(i, j int) bool {
					return expResp.Items[i].ID < expResp.Items[j].ID
				})

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/role-pages/" + sd.RolePages[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &rolepageapp.RolePage{},
			ExpResp:    &sd.RolePages[0],
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
			URL:        "/v1/core/role-pages?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &query.Result[rolepageapp.RolePage]{},
			ExpResp:    &query.Result[rolepageapp.RolePage]{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

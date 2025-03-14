package userrole_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/permissions/userroleapp.go"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/permissions/user_roles?page=1&rows=10",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[userroleapp.UserRole]{},
			ExpResp: &query.Result[userroleapp.UserRole]{
				Page:        1,
				RowsPerPage: 10,
				Total:       2,
				Items:       sd.UserRoles,
			},
			CmpFunc: func(got any, exp any) string {
				// Create sorted versions of both got and exp for comparison
				gotResult := got.(*query.Result[userroleapp.UserRole])
				expResult := exp.(*query.Result[userroleapp.UserRole])

				// Sort got items by ID
				sort.Slice(gotResult.Items, func(i, j int) bool {
					return gotResult.Items[i].ID < gotResult.Items[j].ID
				})

				// Sort exp items by ID (although they should already be sorted now)
				sort.Slice(expResult.Items, func(i, j int) bool {
					return expResult.Items[i].ID < expResult.Items[j].ID
				})

				return cmp.Diff(gotResult, expResult)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/permissions/user_roles/" + sd.UserRoles[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userroleapp.UserRole{},
			ExpResp:    &sd.UserRoles[0],
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
			URL:        "/v1/permissions/user_roles?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &query.Result[userroleapp.UserRole]{},
			ExpResp:    &query.Result[userroleapp.UserRole]{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

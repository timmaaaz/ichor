package page_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
)

func queryByUserID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "user-with-access",
			URL:        "/v1/core/pages/user/" + sd.Users[0].ID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &pageapp.Pages{},
			ExpResp: &pageapp.Pages{
				Items: sd.Pages[3:6], // Only pages 3-5 have can_access=true for tu1
			},
			CmpFunc: func(got any, exp any) string {
				// Pages should be returned ordered by sort_order, name
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "admin-querying-user",
			URL:        "/v1/core/pages/user/" + sd.Users[0].ID.String(),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &pageapp.Pages{},
			ExpResp: &pageapp.Pages{
				Items: sd.Pages[3:6], // Admin can query any user's pages
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByUserID401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/core/pages/user/" + sd.Users[0].ID.String(),
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &pageapp.Pages{},
			ExpResp:    &pageapp.Pages{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByUserID400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid-uuid",
			URL:        "/v1/core/pages/user/invalid-uuid",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &pageapp.Pages{},
			ExpResp:    &pageapp.Pages{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

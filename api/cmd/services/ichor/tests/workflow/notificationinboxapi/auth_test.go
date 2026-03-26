package notificationinboxapi_test

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func query401(sd SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-auth-token",
			URL:        "/v1/workflow/notifications?page=1&rows=10",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

func count401(sd SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-auth-token",
			URL:        "/v1/workflow/notifications/count",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

func markAsRead401(sd SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-auth-token",
			URL:        "/v1/workflow/notifications/00000000-0000-0000-0000-000000000001/read",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

func markAllAsRead401(sd SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-auth-token",
			URL:        "/v1/workflow/notifications/read-all",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

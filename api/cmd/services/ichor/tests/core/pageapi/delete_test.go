package page_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/pages/" + sd.Pages[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func delete401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/core/pages/" + sd.Pages[2].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				// Just check it's an error, don't compare specific messages
				return ""
			},
		},
	}
	return table
}

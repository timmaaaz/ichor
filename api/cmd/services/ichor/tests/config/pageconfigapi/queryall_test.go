package pageconfigapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func queryAll200(sd apitest.SeedData) []apitest.Table {
	expected := pageconfigapp.PageConfigs(sd.PageConfigs)

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-configs/all",
			Method:     http.MethodGet,
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			GotResp:    &pageconfigapp.PageConfigs{},
			ExpResp:    &expected,
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryAll401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-configs/all",
			Method:     http.MethodGet,
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "expected authorization header format: Bearer <token>"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

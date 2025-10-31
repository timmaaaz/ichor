package page_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/pages",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &pageapp.NewPage{
				Path:      "/test/create",
				Name:      "Test Create Page",
				Module:    "testModule",
				Icon:      "test-icon",
				SortOrder: 100,
				IsActive:  true,
			},
			GotResp: &pageapp.Page{},
			ExpResp: &pageapp.Page{
				Path:      "/test/create",
				Name:      "Test Create Page",
				Module:    "testModule",
				Icon:      "test-icon",
				SortOrder: 100,
				IsActive:  true,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*pageapp.Page)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*pageapp.Page)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/pages",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &pageapp.NewPage{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"path\",\"error\":\"path is a required field\"},{\"field\":\"name\",\"error\":\"name is a required field\"},{\"field\":\"module\",\"error\":\"module is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/pages",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pageapp.NewPage{
				Path:      "/test/create",
				Name:      "Test Create Page",
				Module:    "testModule",
				Icon:      "test-icon",
				SortOrder: 100,
				IsActive:  true,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

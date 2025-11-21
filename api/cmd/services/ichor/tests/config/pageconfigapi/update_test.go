package pageconfigapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &pageconfigapp.UpdatePageConfig{
				Name: dbtest.StringPointer("Updated Name"),
			},
			GotResp: &pageconfigapp.PageConfig{},
			ExpResp: &pageconfigapp.PageConfig{
				ID:        sd.PageConfigs[0].ID,
				Name:      "Updated Name",
				IsDefault: sd.PageConfigs[0].IsDefault,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*pageconfigapp.PageConfig)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*pageconfigapp.PageConfig)
				expResp.UserID = gotResp.UserID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid-id",
			URL:        "/v1/config/page-configs/id/not-a-uuid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &pageconfigapp.UpdatePageConfig{
				Name: dbtest.StringPointer("Updated Name"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID,
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pageconfigapp.UpdatePageConfig{
				Name: dbtest.StringPointer("Updated Name"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID,
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pageconfigapp.UpdatePageConfig{
				Name: dbtest.StringPointer("Updated Name"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pageconfigapp.UpdatePageConfig{
				Name: dbtest.StringPointer("Updated Name"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update404(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "notfound",
			URL:        "/v1/config/page-configs/id/" + uuid.NewString(),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &pageconfigapp.UpdatePageConfig{
				Name: dbtest.StringPointer("Updated Name"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "page config not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

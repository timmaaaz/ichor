package role_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/permissions/roleapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/permissions/roles/%s", sd.Roles[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &roleapp.UpdateRole{
				Name:        dbtest.StringPointer("testUpdate"),
				Description: dbtest.StringPointer("testUpdate"),
			},
			GotResp: &roleapp.Role{},
			ExpResp: &roleapp.Role{
				ID:          sd.Roles[1].ID,
				Name:        "testUpdate",
				Description: "testUpdate",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*roleapp.Role)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*roleapp.Role)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "the name is empty",
			URL:        fmt.Sprintf("/v1/permissions/roles/%s", sd.Roles[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &roleapp.UpdateRole{
				Name: dbtest.StringPointer("123456789012345678901234567890123456789012345678901"),
			},
			GotResp: &roleapp.Role{},
			ExpResp: &roleapp.Role{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty token",
			URL:        fmt.Sprintf("/v1/permissions/roles/%s", sd.Roles[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        fmt.Sprintf("/v1/permissions/roles/%s", sd.Roles[0].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "role admin only",
			URL:        fmt.Sprintf("/v1/permissions/roles/%s", sd.Roles[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

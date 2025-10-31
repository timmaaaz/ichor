package userrole_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/userroleapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/user-roles",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &userroleapp.NewUserRole{
				UserID: sd.Users[0].ID.String(),
				RoleID: sd.Roles[2].ID,
			},
			GotResp: &userroleapp.UserRole{},
			ExpResp: &userroleapp.UserRole{
				UserID: sd.Users[0].ID.String(),
				RoleID: sd.Roles[2].ID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*userroleapp.UserRole)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*userroleapp.UserRole)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing role id",
			URL:        "/v1/core/user-roles",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userroleapp.NewUserRole{
				UserID: sd.Users[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"role_id\",\"error\":\"role_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*errs.Error)

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "missing user id",
			URL:        "/v1/core/user-roles",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userroleapp.NewUserRole{
				RoleID: sd.Roles[2].ID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"user_id\",\"error\":\"user_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*errs.Error)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/core/user-roles",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        "/v1/core/user-roles",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        "/v1/core/user-roles",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/core/user-roles",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

package tableaccess_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/permissions/tableaccessapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/permissions/table-access/%s", sd.TableAccesses[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &tableaccessapp.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(false),
			},
			GotResp: &tableaccessapp.TableAccess{},
			ExpResp: &tableaccessapp.TableAccess{
				ID:        sd.TableAccesses[1].ID,
				RoleID:    sd.TableAccesses[1].RoleID,
				TableName: sd.TableAccesses[1].TableName,
				CanCreate: false,
				CanRead:   false,
				CanUpdate: false,
				CanDelete: false,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*tableaccessapp.TableAccess)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*tableaccessapp.TableAccess)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	// TODO: Test role not found
	return []apitest.Table{
		{
			Name:       "the name is empty",
			URL:        fmt.Sprintf("/v1/permissions/table-access/%s", sd.TableAccesses[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &tableaccessapp.UpdateTableAccess{
				RoleID: dbtest.StringPointer("not a uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"role_id","error":"role_id must be a valid UUID"}]`),
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
			URL:        fmt.Sprintf("/v1/permissions/table-access/%s", sd.TableAccesses[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &tableaccessapp.UpdateTableAccess{
				TableName: dbtest.StringPointer("123456789012345678901234567890123456789012345678901"),
			},
			GotResp: &tableaccessapp.TableAccess{},
			ExpResp: &tableaccessapp.TableAccess{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        fmt.Sprintf("/v1/permissions/table-access/%s", sd.TableAccesses[0].ID),
			Token:      sd.Admins[0].Token + "bad",
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
			URL:        fmt.Sprintf("/v1/permissions/table-access/%s", sd.TableAccesses[0].ID),
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

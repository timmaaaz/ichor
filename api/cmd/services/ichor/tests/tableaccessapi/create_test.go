package tableaccess_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/permissions/tableaccessapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/permissions/table-access",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &tableaccessapp.NewTableAccess{
				RoleID:    sd.Roles[len(sd.Roles)-1].ID,
				TableName: "allocation_results",
				CanCreate: true,
				CanRead:   true,
				CanUpdate: true,
				CanDelete: true,
			},
			GotResp: &tableaccessapp.TableAccess{},
			ExpResp: &tableaccessapp.TableAccess{
				RoleID:    sd.Roles[len(sd.Roles)-1].ID,
				TableName: "allocation_results",
				CanCreate: true,
				CanRead:   true,
				CanUpdate: true,
				CanDelete: true,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*tableaccessapp.TableAccess)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*tableaccessapp.TableAccess)
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
			URL:        "/v1/permissions/table-access",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &tableaccessapp.NewTableAccess{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, `validate: [{"field":"role_id","error":"role_id is a required field"},{"field":"table_name","error":"table_name is a required field"}]`),
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
			URL:        "/v1/permissions/table-access",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &tableaccessapp.NewTableAccess{
				RoleID:    sd.Roles[len(sd.Roles)-1].ID,
				TableName: "table_access",
				CanCreate: true,
				CanRead:   true,
				CanUpdate: true,
				CanDelete: true,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

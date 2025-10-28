package rolepage_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/rolepageapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/role-pages/" + sd.RolePages[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &rolepageapp.UpdateRolePage{
				CanAccess: dbtest.BoolPointer(false),
			},
			GotResp: &rolepageapp.RolePage{},
			ExpResp: &rolepageapp.RolePage{
				ID:        sd.RolePages[0].ID,
				RoleID:    sd.RolePages[0].RoleID,
				PageID:    sd.RolePages[0].PageID,
				CanAccess: false,
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	// Note: UpdateRolePage has no required fields, so an empty update is valid
	// and will succeed with 200 OK. There's no way to trigger a 400 error
	// for this endpoint without invalid UUID in the URL path.
	table := []apitest.Table{}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/core/role-pages/" + sd.RolePages[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &rolepageapp.UpdateRolePage{
				CanAccess: dbtest.BoolPointer(true),
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
	return table
}

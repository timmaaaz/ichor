package rolepage_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/rolepageapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/role-pages",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &rolepageapp.NewRolePage{
				RoleID:    sd.RolePages[0].RoleID,
				PageID:    sd.Pages[5].ID, // Use a different page to avoid unique constraint
				CanAccess: true,
			},
			GotResp: &rolepageapp.RolePage{},
			ExpResp: &rolepageapp.RolePage{
				RoleID:    sd.RolePages[0].RoleID,
				PageID:    sd.Pages[5].ID,
				CanAccess: true,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*rolepageapp.RolePage)
				expResp := exp.(*rolepageapp.RolePage)

				// Don't compare ID as it will be generated
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing-roleid",
			URL:        "/v1/core/role-pages",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &rolepageapp.NewRolePage{
				PageID:    sd.Pages[0].ID,
				CanAccess: true,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
		{
			Name:       "missing-pageid",
			URL:        "/v1/core/role-pages",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &rolepageapp.NewRolePage{
				RoleID:    sd.RolePages[0].RoleID,
				CanAccess: true,
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

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/core/role-pages",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &rolepageapp.NewRolePage{
				RoleID:    sd.RolePages[0].RoleID,
				PageID:    sd.Pages[0].ID,
				CanAccess: true,
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

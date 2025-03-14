package userrole_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/permissions/userroleapp.go"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/permissions/user_roles",
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

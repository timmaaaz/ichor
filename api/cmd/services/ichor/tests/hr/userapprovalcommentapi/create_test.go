package userapprovalcomment_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/hr/commentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/hr/comments",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &commentapp.NewUserApprovalComment{
				Comment:     "Please change your username and password",
				UserID:      sd.Users[0].ID.String(),
				CommenterID: sd.Users[1].ID.String(),
			},
			GotResp: &commentapp.UserApprovalComment{},
			ExpResp: &commentapp.UserApprovalComment{
				Comment:     "Please change your username and password",
				UserID:      sd.Users[0].ID.String(),
				CommenterID: sd.Users[1].ID.String(),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*commentapp.UserApprovalComment)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*commentapp.UserApprovalComment)
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {

	table := []apitest.Table{
		{
			Name:       "missing comment",
			URL:        "/v1/hr/comments",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &commentapp.NewUserApprovalComment{
				UserID:      sd.Users[0].ID.String(),
				CommenterID: sd.Users[1].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"comment\",\"error\":\"comment is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing user id",
			URL:        "/v1/hr/comments",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &commentapp.NewUserApprovalComment{
				Comment:     "Please change your username and password",
				CommenterID: sd.Users[1].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"user_id\",\"error\":\"user_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing commenter id",
			URL:        "/v1/hr/comments",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &commentapp.NewUserApprovalComment{
				Comment: "Please change your username and password",
				UserID:  sd.Users[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"commenter_id\",\"error\":\"commenter_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

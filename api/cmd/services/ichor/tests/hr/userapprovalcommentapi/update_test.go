package userapprovalcomment_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/hr/commentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/hr/user-approval-comments/" + sd.UserApprovalComments[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &commentapp.UpdateUserApprovalComment{
				Comment: dbtest.StringPointer("New Comment"),
			},
			GotResp: &commentapp.UserApprovalComment{},
			ExpResp: &sd.UserApprovalComments[0],
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*commentapp.UserApprovalComment)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*commentapp.UserApprovalComment)
				expResp.Comment = "New Comment"

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-id",
			URL:        "/v1/hr/user-approval-comments/abc",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: commentapp.UpdateUserApprovalComment{
				Comment: dbtest.StringPointer("new comment"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 3"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no comment",
			URL:        "/v1/hr/user-approval-comments/" + sd.UserApprovalComments[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: commentapp.UpdateUserApprovalComment{
				Comment: nil,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"comment","error":"comment is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

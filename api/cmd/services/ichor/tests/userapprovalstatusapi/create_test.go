package userapprovalstatus_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/users/status/approvalapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	newUUID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users/status/approvals",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &approvalapp.NewUserApprovalStatus{
				IconID: newUUID.String(),
				Name:   "TestUserApprovalStatus",
			},
			GotResp: &approvalapp.UserApprovalStatus{},
			ExpResp: &approvalapp.UserApprovalStatus{
				IconID: newUUID.String(),
				Name:   "TestUserApprovalStatus",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*approvalapp.UserApprovalStatus)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*approvalapp.UserApprovalStatus)
				expResp.ID = gotResp.ID

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {

	newUUID, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	table := []apitest.Table{
		{
			Name:       "missing icon id",
			URL:        "/v1/users/status/approvals",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &approvalapp.NewUserApprovalStatus{
				Name: "missing icon id",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"icon_id\",\"error\":\"iconID is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing name",
			URL:        "/v1/users/status/approvals",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &approvalapp.NewUserApprovalStatus{
				IconID: newUUID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

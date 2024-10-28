package approvalstatus_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/approvalstatusapp"
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
			URL:        "/v1/approvalstatus",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &approvalstatusapp.NewApprovalStatus{
				IconID: newUUID.String(),
				Name:   "TestApprovalStatus",
			},
			GotResp: &approvalstatusapp.ApprovalStatus{},
			ExpResp: &approvalstatusapp.ApprovalStatus{
				IconID: newUUID.String(),
				Name:   "TestApprovalStatus",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*approvalstatusapp.ApprovalStatus)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*approvalstatusapp.ApprovalStatus)
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
			URL:        "/v1/approvalstatus",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &approvalstatusapp.NewApprovalStatus{
				Name: "missing icon id",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"iconID\",\"error\":\"iconID is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing name",
			URL:        "/v1/approvalstatus",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &approvalstatusapp.NewApprovalStatus{
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

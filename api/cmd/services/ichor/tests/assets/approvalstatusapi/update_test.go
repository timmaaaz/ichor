package approvalstatus_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/approvalstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {

	newUUID := uuid.NewString()

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets/approval-status/" + sd.ApprovalStatuses[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &approvalstatusapp.UpdateApprovalStatus{
				Name:   dbtest.StringPointer("UpdatedApprovalStatus"),
				IconID: dbtest.StringPointer(newUUID),
			},
			GotResp: &approvalstatusapp.ApprovalStatus{},
			ExpResp: &approvalstatusapp.ApprovalStatus{Name: "UpdatedApprovalStatus", IconID: newUUID},
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

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-id",
			URL:        "/v1/assets/approval-status/abc",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: approvalstatusapp.UpdateApprovalStatus{
				Name:   dbtest.StringPointer("UpdatedApprovalStatus"),
				IconID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 3"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

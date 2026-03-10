package approvalapi_test

import (
	"net/http"
	"testing"

	"github.com/timmaaaz/ichor/api/domain/http/workflow/approvalapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_ApprovalResolve(t *testing.T) {
	at := apitest.StartTest(t, "Test_ApprovalResolve")

	sd, err := insertSeedData(at.DB, at.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	at.Run(t, resolveTests(sd), "resolve")
}

func resolveTests(sd ApproveSeedData) []apitest.Table {
	adminToken := sd.Admins[0].Token
	approvalURL := "/v1/workflow/approvals/" + sd.ApprovalID.String() + "/resolve"
	approvalWithTokenURL := "/v1/workflow/approvals/" + sd.ApprovalWithTokenID.String() + "/resolve"

	resolveBody := approvalapi.ResolveRequest{
		Resolution: "approved",
		Reason:     "looks good",
	}

	return []apitest.Table{
		{
			Name:       "first-resolve-returns-200",
			URL:        approvalURL,
			Token:      adminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      resolveBody,
			GotResp:    &approvalapi.Approval{},
			ExpResp: &approvalapi.Approval{
				Status: "approved",
			},
			CmpFunc: func(got, exp any) string {
				gotApp := got.(*approvalapi.Approval)
				expApp := exp.(*approvalapi.Approval)
				if gotApp.Status != expApp.Status {
					return "expected status=approved, got " + gotApp.Status
				}
				return ""
			},
		},
		{
			Name:       "double-resolve-returns-200-not-412",
			URL:        approvalURL, // same approval as above — already resolved
			Token:      adminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      resolveBody,
			GotResp:    &approvalapi.Approval{},
			ExpResp: &approvalapi.Approval{
				Status: "approved",
			},
			CmpFunc: func(got, exp any) string {
				gotApp := got.(*approvalapi.Approval)
				expApp := exp.(*approvalapi.Approval)
				if gotApp.Status != expApp.Status {
					return "expected status=approved on retry, got " + gotApp.Status
				}
				return ""
			},
		},
		{
			Name:       "resolve-with-task-token-returns-200",
			URL:        approvalWithTokenURL,
			Token:      adminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      resolveBody,
			GotResp:    &approvalapi.Approval{},
			ExpResp: &approvalapi.Approval{
				Status: "approved",
			},
			CmpFunc: func(got, exp any) string {
				gotApp := got.(*approvalapi.Approval)
				expApp := exp.(*approvalapi.Approval)
				if gotApp.Status != expApp.Status {
					return "expected status=approved, got " + gotApp.Status
				}
				return ""
			},
		},
		{
			Name:       "double-resolve-with-token-returns-200",
			URL:        approvalWithTokenURL, // already resolved above
			Token:      adminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      resolveBody,
			GotResp:    &approvalapi.Approval{},
			ExpResp: &approvalapi.Approval{
				Status: "approved",
			},
			CmpFunc: func(got, exp any) string {
				gotApp := got.(*approvalapi.Approval)
				expApp := exp.(*approvalapi.Approval)
				if gotApp.Status != expApp.Status {
					return "expected status=approved on retry with token, got " + gotApp.Status
				}
				return ""
			},
		},
	}
}

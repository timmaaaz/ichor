package action_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func getExecutionStatus401(sd ActionSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s", sd.CompletedExecutionID),
			Token:      "&nbsp;",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func getExecutionStatus404(sd ActionSeedData) []apitest.Table {
	nonExistentID := uuid.New()

	return []apitest.Table{
		{
			Name:       "execution-not-found",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s", nonExistentID),
			Token:      sd.AdminUser.Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				// Verify we get a not found error
				if gotErr.Code.Value() != errs.NotFound.Value() {
					return cmp.Diff(gotErr.Code.Value(), errs.NotFound.Value())
				}
				return ""
			},
		},
	}
}

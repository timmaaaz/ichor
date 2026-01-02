package alert_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/alertapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func dismissSelected200(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "dismiss-single-alert",
			URL:        "/v1/workflow/alerts/dismiss-selected",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &alertapi.BulkSelectedRequest{
				IDs: []string{sd.HighSeverityID.String()},
			},
			GotResp: &alertapi.BulkActionResult{},
			ExpResp: &alertapi.BulkActionResult{
				Count:   1,
				Skipped: 0,
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func dismissAll200(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "dismiss-all-active",
			URL:        "/v1/workflow/alerts/dismiss-all",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			GotResp:    &alertapi.BulkActionResult{},
			ExpResp: &alertapi.BulkActionResult{
				Skipped: 0,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*alertapi.BulkActionResult)

				// Count should be >= 0 (some may have been acknowledged/dismissed in previous tests)
				if gotResp.Count < 0 {
					return "count should be non-negative"
				}
				if gotResp.Skipped != 0 {
					return "skipped should be 0 for dismiss-all"
				}

				return ""
			},
		},
	}
}

func dismissAll401(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-auth-token",
			URL:        "/v1/workflow/alerts/dismiss-all",
			Token:      "", // No token
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

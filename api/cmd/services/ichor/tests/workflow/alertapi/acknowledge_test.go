package alert_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/alertapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func acknowledgeSelected200(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "acknowledge-single-alert",
			URL:        "/v1/workflow/alerts/acknowledge-selected",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &alertapi.BulkSelectedRequest{
				IDs:   []string{sd.LowSeverityID.String()},
				Notes: "Bulk acknowledge test",
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

func acknowledgeSelectedPartialSkip200(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "partial-skip-non-recipient",
			URL:        "/v1/workflow/alerts/acknowledge-selected",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &alertapi.BulkSelectedRequest{
				IDs: []string{
					sd.MediumSeverityID.String(), // User is recipient
					sd.NonRecipientID.String(),   // User is NOT a recipient
				},
				Notes: "Partial skip test",
			},
			GotResp: &alertapi.BulkActionResult{},
			ExpResp: &alertapi.BulkActionResult{
				Count:   1,
				Skipped: 1,
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func acknowledgeSelected400(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-uuid",
			URL:        "/v1/workflow/alerts/acknowledge-selected",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: &alertapi.BulkSelectedRequest{
				IDs: []string{"not-a-uuid"},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				if gotErr.Code.Value() != errs.InvalidArgument.Value() {
					return cmp.Diff(gotErr.Code.Value(), errs.InvalidArgument.Value())
				}
				return ""
			},
		},
		{
			Name:       "empty-ids",
			URL:        "/v1/workflow/alerts/acknowledge-selected",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: &alertapi.BulkSelectedRequest{
				IDs: []string{},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				if gotErr.Code.Value() != errs.InvalidArgument.Value() {
					return cmp.Diff(gotErr.Code.Value(), errs.InvalidArgument.Value())
				}
				return ""
			},
		},
	}
}

func acknowledgeSelected401(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-auth-token",
			URL:        "/v1/workflow/alerts/acknowledge-selected",
			Token:      "", // No token
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			Input: &alertapi.BulkSelectedRequest{
				IDs: []string{sd.LowSeverityID.String()},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

func acknowledgeAll200(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "acknowledge-all-active",
			URL:        "/v1/workflow/alerts/acknowledge-all",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &alertapi.BulkAllRequest{
				Notes: "Acknowledge all test",
			},
			GotResp: &alertapi.BulkActionResult{},
			ExpResp: &alertapi.BulkActionResult{
				Skipped: 0,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*alertapi.BulkActionResult)

				// Count should be >= 0 (some may have been acknowledged in previous tests)
				if gotResp.Count < 0 {
					return "count should be non-negative"
				}
				if gotResp.Skipped != 0 {
					return "skipped should be 0 for acknowledge-all"
				}

				return ""
			},
		},
	}
}

func acknowledgeAll401(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-auth-token",
			URL:        "/v1/workflow/alerts/acknowledge-all",
			Token:      "", // No token
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			Input:      &alertapi.BulkAllRequest{},
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

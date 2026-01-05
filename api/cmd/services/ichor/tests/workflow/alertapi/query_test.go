package alert_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/alertapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func queryMine200(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/alerts/mine?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[alertapi.Alert]{},
			ExpResp: &query.Result[alertapi.Alert]{
				Page:        1,
				RowsPerPage: 10,
				Total:       3,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*query.Result[alertapi.Alert])
				expResp := exp.(*query.Result[alertapi.Alert])

				// Verify pagination metadata
				if gotResp.Page != expResp.Page {
					return "page mismatch"
				}
				if gotResp.Total != expResp.Total {
					return "total count mismatch: expected 3 alerts"
				}

				return ""
			},
		},
	}
}

func queryMineWithSeverityFilter200(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "single-severity",
			URL:        "/v1/workflow/alerts/mine?page=1&rows=10&severity=low",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[alertapi.Alert]{},
			ExpResp: &query.Result[alertapi.Alert]{
				Page:        1,
				RowsPerPage: 10,
				Total:       1, // Only 1 low severity alert
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*query.Result[alertapi.Alert])
				expResp := exp.(*query.Result[alertapi.Alert])

				if gotResp.Total != expResp.Total {
					return "expected 1 low severity alert"
				}
				if len(gotResp.Items) != 1 {
					return "expected exactly 1 item"
				}
				if gotResp.Items[0].Severity != "low" {
					return "expected severity to be low"
				}

				return ""
			},
		},
		{
			Name:       "multi-severity",
			URL:        "/v1/workflow/alerts/mine?page=1&rows=10&severity=low,medium",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[alertapi.Alert]{},
			ExpResp: &query.Result[alertapi.Alert]{
				Page:        1,
				RowsPerPage: 10,
				Total:       2, // low + medium
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*query.Result[alertapi.Alert])
				expResp := exp.(*query.Result[alertapi.Alert])

				if gotResp.Total != expResp.Total {
					return "expected 2 alerts (low + medium)"
				}

				// Verify all returned alerts are either low or medium severity
				for _, item := range gotResp.Items {
					if item.Severity != "low" && item.Severity != "medium" {
						return "unexpected severity: " + item.Severity
					}
				}

				return ""
			},
		},
		{
			Name:       "all-severities",
			URL:        "/v1/workflow/alerts/mine?page=1&rows=10&severity=low,medium,high",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[alertapi.Alert]{},
			ExpResp: &query.Result[alertapi.Alert]{
				Page:        1,
				RowsPerPage: 10,
				Total:       3, // low + medium + high
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*query.Result[alertapi.Alert])
				expResp := exp.(*query.Result[alertapi.Alert])

				if gotResp.Total != expResp.Total {
					return "expected 3 alerts (low + medium + high)"
				}

				return ""
			},
		},
	}
}

func queryMineWithInvalidSeverity400(sd AlertSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-severity-value",
			URL:        "/v1/workflow/alerts/mine?page=1&rows=10&severity=invalid",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
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

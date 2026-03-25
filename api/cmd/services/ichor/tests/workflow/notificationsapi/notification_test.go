package notificationsapi_test

import (
	"net/http"
	"testing"

	"github.com/timmaaaz/ichor/api/domain/http/workflow/notificationsapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func Test_NotificationsAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_NotificationsAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, summary200(sd), "summary-200")
	test.Run(t, summary401(sd), "summary-401")
}

// -------------------------------------------------------------------------

func summary200(sd NotificationSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "authenticated",
			URL:        "/v1/workflow/notifications/summary",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &notificationsapi.NotificationSummary{},
			ExpResp: &notificationsapi.NotificationSummary{
				Alerts: notificationsapi.AlertSummary{
					TotalActive: sd.ExpectedTotal,
					Low:         sd.ExpectedLow,
					Medium:      sd.ExpectedMedium,
					High:        sd.ExpectedHigh,
					Critical:    sd.ExpectedCritical,
				},
				Approvals: notificationsapi.ApprovalSummary{
					PendingCount: sd.ExpectedPending,
				},
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*notificationsapi.NotificationSummary)
				expResp := exp.(*notificationsapi.NotificationSummary)

				// Verify alert severity counts.
				if gotResp.Alerts.TotalActive != expResp.Alerts.TotalActive {
					return "alert TotalActive mismatch"
				}
				if gotResp.Alerts.Low != expResp.Alerts.Low {
					return "alert Low count mismatch"
				}
				if gotResp.Alerts.Medium != expResp.Alerts.Medium {
					return "alert Medium count mismatch"
				}
				if gotResp.Alerts.High != expResp.Alerts.High {
					return "alert High count mismatch"
				}
				if gotResp.Alerts.Critical != expResp.Alerts.Critical {
					return "alert Critical count mismatch"
				}

				// Verify pending approval count.
				if gotResp.Approvals.PendingCount != expResp.Approvals.PendingCount {
					return "approval PendingCount mismatch"
				}

				return ""
			},
		},
	}
}

func summary401(sd NotificationSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-auth-token",
			URL:        "/v1/workflow/notifications/summary",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

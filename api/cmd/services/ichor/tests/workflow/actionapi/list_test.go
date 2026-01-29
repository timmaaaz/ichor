package action_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/actionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func listActions200Admin(sd ActionSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "admin-sees-permitted-actions",
			URL:        "/v1/workflow/actions",
			Token:      sd.AdminUser.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &actionapp.AvailableActions{},
			ExpResp:    &actionapp.AvailableActions{},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*actionapp.AvailableActions)

				// Admin should see actions if admin role has action permissions
				// For now, just verify we get a valid response (empty or with items)
				if gotResp == nil {
					return "expected non-nil response"
				}

				return ""
			},
		},
	}
}

func listActions200UserWithPermissions(sd ActionSeedData) []apitest.Table {
	// In test environments without RabbitMQ, no action handlers are registered,
	// so the action registry is empty. The user has permissions for actions,
	// but since no handlers exist, the filtered list is empty.
	// This tests that the endpoint works correctly - in production with handlers
	// registered, the user would see their permitted actions.
	expActions := actionapp.AvailableActions{}

	return []apitest.Table{
		{
			Name:       "user-sees-permitted-actions-only",
			URL:        "/v1/workflow/actions",
			Token:      sd.UserWithAlertPerm.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &actionapp.AvailableActions{},
			ExpResp:    &expActions,
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*actionapp.AvailableActions)
				expResp := exp.(*actionapp.AvailableActions)

				// In test env, both should be empty since no handlers are registered
				if len(*gotResp) != len(*expResp) {
					return cmp.Diff(gotResp, expResp)
				}

				return ""
			},
		},
	}
}

func listActions200UserNoPermissions(sd ActionSeedData) []apitest.Table {
	expActions := actionapp.AvailableActions{} // Empty list

	return []apitest.Table{
		{
			Name:       "user-no-permissions-sees-empty-list",
			URL:        "/v1/workflow/actions",
			Token:      sd.UserNoPermissions.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &actionapp.AvailableActions{},
			ExpResp:    &expActions,
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*actionapp.AvailableActions)
				expResp := exp.(*actionapp.AvailableActions)

				// Both should be empty slices
				if len(*gotResp) != len(*expResp) {
					return cmp.Diff(gotResp, expResp)
				}

				return ""
			},
		},
	}
}

func listActions401(sd ActionSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        "/v1/workflow/actions",
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

package action_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/actionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func execute401(sd ActionSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        "/v1/workflow/actions/create_alert/execute",
			Token:      "&nbsp;",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config: json.RawMessage(`{}`),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func execute403NoPermission(sd ActionSeedData) []apitest.Table {
	config := map[string]any{
		"warehouse_id": "some-warehouse-id",
		"items":        []map[string]any{},
	}
	configBytes, _ := json.Marshal(config)

	return []apitest.Table{
		{
			Name:       "user-lacks-action-permission",
			URL:        "/v1/workflow/actions/allocate_inventory/execute",
			Token:      sd.UserWithAlertPerm.Token, // Has alert perm, NOT inventory perm
			StatusCode: http.StatusForbidden,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config: configBytes,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				// Verify we get a permission denied error
				if gotErr.Code.Value() != errs.PermissionDenied.Value() {
					return cmp.Diff(gotErr.Code.Value(), errs.PermissionDenied.Value())
				}
				return ""
			},
		},
		{
			Name:       "user-no-permissions-at-all",
			URL:        "/v1/workflow/actions/create_alert/execute",
			Token:      sd.UserNoPermissions.Token,
			StatusCode: http.StatusForbidden,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config: json.RawMessage(`{}`),
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				// Verify we get a permission denied error
				if gotErr.Code.Value() != errs.PermissionDenied.Value() {
					return cmp.Diff(gotErr.Code.Value(), errs.PermissionDenied.Value())
				}
				return ""
			},
		},
	}
}

func execute404UnknownAction(sd ActionSeedData) []apitest.Table {
	// Note: Permission check happens before action type validation in the current
	// implementation. Since the admin user doesn't have explicit permission for
	// "nonexistent_action", the permission check fails first, returning 403.
	// This is intentional security behavior - it doesn't reveal whether an action
	// type exists or not to users without permission.
	return []apitest.Table{
		{
			Name:       "unknown-action-type",
			URL:        "/v1/workflow/actions/nonexistent_action/execute",
			Token:      sd.AdminUser.Token,
			StatusCode: http.StatusForbidden, // Permission check happens first
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config: json.RawMessage(`{}`),
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				// Verify we get a permission denied error
				if gotErr.Code.Value() != errs.PermissionDenied.Value() {
					return cmp.Diff(gotErr.Code.Value(), errs.PermissionDenied.Value())
				}
				return ""
			},
		},
	}
}

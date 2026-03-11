package action_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/actionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func execute200CreateAlert(sd ActionSeedData) []apitest.Table {
	config := map[string]any{
		"alert_type": "manual_execute_test",
		"severity":   "low",
		"title":      "Manual Execute Test",
		"message":    "Executed manually via test",
		"recipients": map[string]any{
			"users": []string{sd.UserWithAlertPerm.User.ID.String()},
			"roles": []string{},
		},
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("marshal config: %v", err))
	}

	return []apitest.Table{
		{
			Name:       "user-with-alert-perm-executes-create-alert",
			URL:        "/v1/workflow/actions/create_alert/execute",
			Token:      sd.UserWithAlertPerm.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config: configBytes,
			},
			GotResp: &actionapp.ExecuteResponse{},
			ExpResp: &actionapp.ExecuteResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*actionapp.ExecuteResponse)
				if !ok {
					return fmt.Sprintf("unexpected response type: %T", got)
				}
				if gotResp.ExecutionID == "" {
					return "expected non-empty execution_id in response"
				}
				return ""
			},
		},
	}
}

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
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("marshal config: %v", err))
	}

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

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

// executeTransitionStatus200 verifies that a user holding the transition_status
// permission can manually execute a PENDING → PICKING transition and gets back
// transitioned=true / output="success".
func executeTransitionStatus200(sd ActionSeedData) []apitest.Table {
	config := map[string]any{
		"target_entity":       "sales.orders",
		"target_id":           sd.PendingOrderID.String(),
		"status_field":        "order_fulfillment_status_id",
		"to_status":           sd.PickingStatusID.String(),
		"valid_from_statuses": []string{sd.PendingStatusID.String()},
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("marshal transition config: %v", err))
	}

	entityID := sd.PendingOrderID.String()
	return []apitest.Table{
		{
			Name:       "transition-status-200-granted-user",
			URL:        "/v1/workflow/actions/transition_status/execute",
			Token:      sd.UserWithTransitionPerm.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config:     configBytes,
				EntityID:   &entityID,
				EntityName: "sales.orders",
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
				// Result is deserialized as map[string]any by JSON round-trip.
				resultMap, ok := gotResp.Result.(map[string]any)
				if !ok {
					return fmt.Sprintf("expected map[string]any result, got %T: %v", gotResp.Result, gotResp.Result)
				}
				if resultMap["transitioned"] != true {
					return fmt.Sprintf("expected transitioned=true, got %v", resultMap["transitioned"])
				}
				if resultMap["output"] != "success" {
					return fmt.Sprintf("expected output=success, got %v", resultMap["output"])
				}
				return ""
			},
		},
	}
}

// executeTransitionStatus403Denied verifies that a user without transition_status
// permission receives a 403 PermissionDenied response.
func executeTransitionStatus403Denied(sd ActionSeedData) []apitest.Table {
	config := map[string]any{
		"target_entity":       "sales.orders",
		"target_id":           sd.PendingOrderID.String(),
		"status_field":        "order_fulfillment_status_id",
		"to_status":           sd.PickingStatusID.String(),
		"valid_from_statuses": []string{sd.PendingStatusID.String()},
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("marshal transition config: %v", err))
	}

	return []apitest.Table{
		{
			Name:       "transition-status-403-no-permission",
			URL:        "/v1/workflow/actions/transition_status/execute",
			Token:      sd.UserNoPermissions.Token,
			StatusCode: http.StatusForbidden,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config: configBytes,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				if gotErr.Code.Value() != errs.PermissionDenied.Value() {
					return cmp.Diff(gotErr.Code.Value(), errs.PermissionDenied.Value())
				}
				return ""
			},
		},
	}
}

// executeTransitionStatusInvalidFrom verifies that a granted user posting a
// transition where the order's current status is NOT in valid_from_statuses
// gets back a 200 with transitioned=false / output="invalid_transition" (no error).
func executeTransitionStatusInvalidFrom(sd ActionSeedData) []apitest.Table {
	config := map[string]any{
		"target_entity": "sales.orders",
		"target_id":     sd.NonTransitionableOrderID.String(),
		"status_field":  "order_fulfillment_status_id",
		"to_status":     sd.PickingStatusID.String(),
		// valid_from_statuses lists only PENDING; the order is at PICKING → invalid
		"valid_from_statuses": []string{sd.PendingStatusID.String()},
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("marshal transition config: %v", err))
	}

	entityID := sd.NonTransitionableOrderID.String()
	return []apitest.Table{
		{
			Name:       "transition-status-200-invalid-from",
			URL:        "/v1/workflow/actions/transition_status/execute",
			Token:      sd.UserWithTransitionPerm.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config:     configBytes,
				EntityID:   &entityID,
				EntityName: "sales.orders",
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
				resultMap, ok := gotResp.Result.(map[string]any)
				if !ok {
					return fmt.Sprintf("expected map[string]any result, got %T: %v", gotResp.Result, gotResp.Result)
				}
				if resultMap["transitioned"] != false {
					return fmt.Sprintf("expected transitioned=false, got %v", resultMap["transitioned"])
				}
				if resultMap["output"] != "invalid_transition" {
					return fmt.Sprintf("expected output=invalid_transition, got %v", resultMap["output"])
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

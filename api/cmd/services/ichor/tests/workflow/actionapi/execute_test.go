package action_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

// executeTransitionStatusProtected400 verifies that the protected-list (P3) rejects a manual
// transition_status on an invariant-bearing status field via the HTTP manual-execute path.
// order_fulfillment_status_id is recomputed by the picking flow, so a generic write must be
// blocked (400 InvalidArgument) with a clear, actionable message — the backend-authoritative
// rejection the FE error toast surfaces (Path A). This replaces the former "200 success"
// expectation, which P3 intentionally invalidated by protecting this field; the transition
// success / invalid-from mechanics remain covered at the handler level (Test_TransitionStatusAction).
func executeTransitionStatusProtected400(sd ActionSeedData) []apitest.Table {
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
			Name:       "transition-status-protected-400",
			URL:        "/v1/workflow/actions/transition_status/execute",
			Token:      sd.UserWithTransitionPerm.Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config:     configBytes,
				EntityID:   &entityID,
				EntityName: "sales.orders",
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return fmt.Sprintf("unexpected response type: %T", got)
				}
				if !gotErr.Code.Equal(errs.InvalidArgument) {
					return "expected InvalidArgument, got " + gotErr.Code.String()
				}
				for _, want := range []string{"order_fulfillment_status_id", "sales.orders", "protected"} {
					if !strings.Contains(gotErr.Message, want) {
						return fmt.Sprintf("error message %q missing %q (the FE toast must name the field + reason)", gotErr.Message, want)
					}
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

// executeCreateEntityProtected400 verifies the protected-list rejects a manual create_entity into
// a whole-table-protected entity via the HTTP manual-execute path (the create arm of Path A).
// inventory.inventory_transactions is an append-only ledger; a generic create must be blocked
// (400 InvalidArgument). create_entity supports manual execution, so the protected check is
// reached (unlike update_field — see executeUpdateFieldNotManuallyExecutable).
func executeCreateEntityProtected400(sd ActionSeedData) []apitest.Table {
	config := map[string]any{
		"target_entity": "inventory.inventory_transactions",
		"fields":        map[string]any{"quantity": 1},
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("marshal create_entity config: %v", err))
	}

	return []apitest.Table{
		{
			Name:       "create-entity-protected-400",
			URL:        "/v1/workflow/actions/create_entity/execute",
			Token:      sd.UserWithTransitionPerm.Token, // role grants create_entity
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config: configBytes,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return fmt.Sprintf("unexpected response type: %T", got)
				}
				if !gotErr.Code.Equal(errs.InvalidArgument) {
					return "expected InvalidArgument, got " + gotErr.Code.String()
				}
				for _, want := range []string{"inventory.inventory_transactions", "protected"} {
					if !strings.Contains(gotErr.Message, want) {
						return fmt.Sprintf("error message %q missing %q", gotErr.Message, want)
					}
				}
				return ""
			},
		},
	}
}

// executeUpdateFieldNotManuallyExecutable documents that update_field cannot be triggered via the
// manual-execute HTTP path at all (SupportsManualExecution=false — raw field writes are
// automation-only), so its protected-list is enforced only on the cascade/worker path (covered by
// data/protected_enforcement_test.go), never here. A permitted user still gets FailedPrecondition,
// not a protected 400. This is the honest answer to "does a manual update_field on a protected
// field return 400?" — it never reaches the protected check.
func executeUpdateFieldNotManuallyExecutable(sd ActionSeedData) []apitest.Table {
	config := map[string]any{
		"target_entity": "sales.orders",
		"target_field":  "order_fulfillment_status_id",
		"new_value":     sd.PickingStatusID.String(),
		"conditions": []map[string]any{
			{"field_name": "id", "operator": "equals", "value": sd.PendingOrderID.String()},
		},
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("marshal update_field config: %v", err))
	}

	return []apitest.Table{
		{
			Name:       "update-field-not-manually-executable",
			URL:        "/v1/workflow/actions/update_field/execute",
			Token:      sd.UserWithTransitionPerm.Token, // role grants update_field (so we pass the perm gate)
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: &actionapp.ExecuteRequest{
				Config: configBytes,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return fmt.Sprintf("unexpected response type: %T", got)
				}
				if !gotErr.Code.Equal(errs.FailedPrecondition) {
					return "expected FailedPrecondition (update_field is automation-only), got " + gotErr.Code.String()
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

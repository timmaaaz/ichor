package ordersapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// bindContainer200 binds CONTAINER-A to Orders[3]. Orders[3] is used
// exclusively by binding tests — no other apitest scenario references it.
// CONTAINER-A is sd.Labels[0] (free at start of test). After this scenario
// Orders[3] has 1 active binding (CONTAINER-A); queryBindings200 verifies.
func bindContainer200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &ordersapp.NewOrderContainerBinding{
				ContainerLabelID: sd.Labels[0].ID,
			},
			GotResp: &ordersapp.OrderContainerBinding{},
			ExpResp: &ordersapp.OrderContainerBinding{
				OrderID:          sd.Orders[3].ID,
				ContainerLabelID: sd.Labels[0].ID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*ordersapp.OrderContainerBinding)
				if !ok {
					return fmt.Sprintf("expected *ordersapp.OrderContainerBinding, got %T", got)
				}
				expResp := exp.(*ordersapp.OrderContainerBinding)
				// ID and BoundAt are server-generated.
				expResp.ID = gotResp.ID
				expResp.BoundAt = gotResp.BoundAt
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

// bindContainer409 attempts to bind CONTAINER-C (pre-bound to Orders[4] in
// setup) to Orders[3]. The EXCLUDE constraint one_active_binding_per_container
// rejects: one container, one active binding at a time. Maps to errs.Aborted
// → 409 Conflict per app/sdk/errs/codes.go:179.
func bindContainer409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "exclude-on-second-active-bind",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &ordersapp.NewOrderContainerBinding{
				ContainerLabelID: sd.Labels[2].ID, // CONTAINER-C, pre-bound to Orders[4]
			},
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*errs.Error)
				if !ok {
					return fmt.Sprintf("expected *errs.Error, got %T", got)
				}
				if gotResp.Code.String() != "aborted" {
					return fmt.Sprintf("expected code 'aborted', got %q", gotResp.Code.String())
				}
				return ""
			},
		},
	}
}

// bindContainer400 verifies validator + UUID-parse error paths.
func bindContainer400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-container-label-id",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &ordersapp.NewOrderContainerBinding{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"container_label_id\",\"error\":\"container_label_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-container-label-id",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.NewOrderContainerBinding{
				ContainerLabelID: "not-a-uuid",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"container_label_id\",\"error\":\"container_label_id must be a valid UUID\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// bindContainer404 verifies unknown-order ID surfaces as 404 (app pre-checks
// order existence via QueryByID before attempting the bind).
func bindContainer404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unknown-order",
			URL:        "/v1/sales/orders/00000000-0000-0000-0000-000000000000/bindings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			Input: &ordersapp.NewOrderContainerBinding{
				ContainerLabelID: sd.Labels[1].ID,
			},
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*errs.Error)
				if !ok {
					return fmt.Sprintf("expected *errs.Error, got %T", got)
				}
				if gotResp.Code.String() != "not_found" {
					return fmt.Sprintf("expected code 'not_found', got %q", gotResp.Code.String())
				}
				return ""
			},
		},
	}
}

// bindContainer401 verifies authentication failures (empty token, bad
// signature) and the read-only-user authorization failure (sd.Users[0] has
// CanRead but not CanUpdate on sales.orders).
func bindContainer401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &ordersapp.NewOrderContainerBinding{
				ContainerLabelID: sd.Labels[1].ID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &ordersapp.NewOrderContainerBinding{
				ContainerLabelID: sd.Labels[1].ID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "readonly-user-no-update",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusForbidden,
			Input: &ordersapp.NewOrderContainerBinding{
				ContainerLabelID: sd.Labels[1].ID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.PermissionDenied, "user does not have permission UPDATE for table: sales.orders"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// unbindContainer200 releases the pre-seeded binding (CONTAINER-C ↔ Orders[4])
// from setup. After this scenario, Orders[4] has 0 active bindings and
// CONTAINER-C is free to be rebound by future tests.
func unbindContainer200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/sales/order-container-bindings/%s/unbind", sd.OrderContainerBindings[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNoContent,
		},
	}
}

// unbindContainerIdempotent verifies the bus-layer idempotency contract:
// calling Unbind on an already-unbound binding is a silent no-op (returns
// 204, not 404). Source of truth: ordersbus.UnbindContainer (ErrAlreadyUnbound
// is caught inside the bus and swallowed).
func unbindContainerIdempotent(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "double-unbind-still-204",
			URL:        fmt.Sprintf("/v1/sales/order-container-bindings/%s/unbind", sd.OrderContainerBindings[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNoContent,
		},
	}
}

// unbindContainer404 verifies unknown-binding-id surfaces as 404.
func unbindContainer404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unknown-binding",
			URL:        "/v1/sales/order-container-bindings/00000000-0000-0000-0000-000000000000/unbind",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*errs.Error)
				if !ok {
					return fmt.Sprintf("expected *errs.Error, got %T", got)
				}
				if gotResp.Code.String() != "not_found" {
					return fmt.Sprintf("expected code 'not_found', got %q", gotResp.Code.String())
				}
				return ""
			},
		},
	}
}

// unbindContainer401 — auth failures and read-only-user 403.
func unbindContainer401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/sales/order-container-bindings/%s/unbind", sd.OrderContainerBindings[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/sales/order-container-bindings/%s/unbind", sd.OrderContainerBindings[0].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "readonly-user-no-update",
			URL:        fmt.Sprintf("/v1/sales/order-container-bindings/%s/unbind", sd.OrderContainerBindings[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusForbidden,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.PermissionDenied, "user does not have permission UPDATE for table: sales.orders"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// queryBindings200 — Orders[3] has 1 active binding after bindContainer200.
func queryBindings200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "one-active-binding",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &ordersapp.OrderContainerBindings{},
			ExpResp:    &ordersapp.OrderContainerBindings{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*ordersapp.OrderContainerBindings)
				if !ok {
					return fmt.Sprintf("expected *OrderContainerBindings, got %T", got)
				}
				if len(*gotResp) != 1 {
					return fmt.Sprintf("expected 1 active binding, got %d", len(*gotResp))
				}
				if (*gotResp)[0].ContainerLabelID != sd.Labels[0].ID {
					return fmt.Sprintf("expected ContainerLabelID %s, got %s", sd.Labels[0].ID, (*gotResp)[0].ContainerLabelID)
				}
				return ""
			},
		},
	}
}

// queryBindingsEmpty — Orders[1] has zero active bindings.
func queryBindingsEmpty(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-active-bindings",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &ordersapp.OrderContainerBindings{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*ordersapp.OrderContainerBindings)
				if !ok {
					return fmt.Sprintf("expected *OrderContainerBindings, got %T", got)
				}
				if len(*gotResp) != 0 {
					return fmt.Sprintf("expected 0 active bindings, got %d", len(*gotResp))
				}
				return ""
			},
		},
	}
}

// queryBindings401 — auth failure paths.
func queryBindings401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      "&nbsp;",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[3].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

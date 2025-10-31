package purchaseorderstatusapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-statuses/" + sd.PurchaseOrderStatuses[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}

func delete400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-id",
			URL:        "/v1/procurement/purchase-order-statuses/invalid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusBadRequest,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid UUID length: 7"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func delete401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/procurement/purchase-order-statuses/" + sd.PurchaseOrderStatuses[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission DELETE for table: purchase_order_statuses"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func delete404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/procurement/purchase-order-statuses/" + uuid.NewString(),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "querybyid: db: purchase order status not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

package purchaseorderstatusapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-statuses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &purchaseorderstatusapp.NewPurchaseOrderStatus{
				Name:        "Pending Review",
				Description: "Order is pending review",
				SortOrder:   500,
			},
			GotResp: &purchaseorderstatusapp.PurchaseOrderStatus{},
			ExpResp: &purchaseorderstatusapp.PurchaseOrderStatus{
				Name:        "Pending Review",
				Description: "Order is pending review",
				SortOrder:   500,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*purchaseorderstatusapp.PurchaseOrderStatus)
				expResp := exp.(*purchaseorderstatusapp.PurchaseOrderStatus)
				expResp.ID = gotResp.ID
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-name",
			URL:        "/v1/procurement/purchase-order-statuses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &purchaseorderstatusapp.NewPurchaseOrderStatus{
				Description: "Missing name",
				SortOrder:   500,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/procurement/purchase-order-statuses",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &purchaseorderstatusapp.NewPurchaseOrderStatus{
				Name:        "Test",
				Description: "Test",
				SortOrder:   100,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: procurement.purchase_order_statuses"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

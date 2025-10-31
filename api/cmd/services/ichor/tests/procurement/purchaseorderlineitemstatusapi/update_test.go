package purchaseorderlineitemstatusapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-item-statuses/" + sd.PurchaseOrderLineItemStatuses[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &purchaseorderlineitemstatusapp.UpdatePurchaseOrderLineItemStatus{
				Name: dbtest.StringPointer("Updated Name"),
			},
			GotResp: &purchaseorderlineitemstatusapp.PurchaseOrderLineItemStatus{},
			ExpResp: &purchaseorderlineitemstatusapp.PurchaseOrderLineItemStatus{
				ID:          sd.PurchaseOrderLineItemStatuses[0].ID,
				Name:        "Updated Name",
				Description: sd.PurchaseOrderLineItemStatuses[0].Description,
				SortOrder:   sd.PurchaseOrderLineItemStatuses[0].SortOrder,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-id",
			URL:        "/v1/procurement/purchase-order-line-item-statuses/invalid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &purchaseorderlineitemstatusapp.UpdatePurchaseOrderLineItemStatus{
				Name: dbtest.StringPointer("Updated"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 7"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/procurement/purchase-order-line-item-statuses/" + sd.PurchaseOrderLineItemStatuses[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &purchaseorderlineitemstatusapp.UpdatePurchaseOrderLineItemStatus{
				Name: dbtest.StringPointer("Updated"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: purchase_order_line_item_statuses"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/procurement/purchase-order-line-item-statuses/" + uuid.NewString(),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &purchaseorderlineitemstatusapp.UpdatePurchaseOrderLineItemStatus{
				Name: dbtest.StringPointer("Updated"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "querybyid: db: purchase order line item status not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

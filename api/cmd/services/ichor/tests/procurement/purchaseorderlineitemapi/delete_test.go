package purchaseorderlineitemapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-items/" + sd.PurchaseOrderLineItems[19].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: 204,
			Method:     "DELETE",
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func delete401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/procurement/purchase-order-line-items/" + sd.PurchaseOrderLineItems[18].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 401,
			Method:     "DELETE",
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission DELETE for table: purchase_order_line_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func delete404(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/procurement/purchase-order-line-items/00000000-0000-0000-0000-000000000000",
			Token:      sd.Admins[0].Token,
			StatusCode: 404,
			Method:     "DELETE",
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "querybyid: db: purchase order line item not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

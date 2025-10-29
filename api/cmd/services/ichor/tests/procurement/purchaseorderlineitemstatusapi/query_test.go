package purchaseorderlineitemstatusapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-item-statuses?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[purchaseorderlineitemstatusapp.PurchaseOrderLineItemStatus]{},
			ExpResp: &query.Result[purchaseorderlineitemstatusapp.PurchaseOrderLineItemStatus]{
				Page:        1,
				RowsPerPage: 10,
				Total:       10,
				Items:       sd.PurchaseOrderLineItemStatuses[:10],
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-item-statuses/" + sd.PurchaseOrderLineItemStatuses[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &purchaseorderlineitemstatusapp.PurchaseOrderLineItemStatus{},
			ExpResp:    &sd.PurchaseOrderLineItemStatuses[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

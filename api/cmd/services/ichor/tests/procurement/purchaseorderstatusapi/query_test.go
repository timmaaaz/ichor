package purchaseorderstatusapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-statuses?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[purchaseorderstatusapp.PurchaseOrderStatus]{},
			ExpResp: &query.Result[purchaseorderstatusapp.PurchaseOrderStatus]{
				Page:        1,
				RowsPerPage: 10,
				Total:       10,
				Items:       sd.PurchaseOrderStatuses[:10],
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
			URL:        "/v1/procurement/purchase-order-statuses/" + sd.PurchaseOrderStatuses[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &purchaseorderstatusapp.PurchaseOrderStatus{},
			ExpResp:    &sd.PurchaseOrderStatuses[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

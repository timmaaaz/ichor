package purchaseorderapi_test

import (
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-orders?page=1&rows=10&orderBy=orderNumber,ASC",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[purchaseorderapp.PurchaseOrder]{},
			ExpResp: &query.Result[purchaseorderapp.PurchaseOrder]{
				Page:        1,
				RowsPerPage: 10,
				Total:       10,
				Items:       sd.PurchaseOrders[:10],
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[purchaseorderapp.PurchaseOrder])
				expResp := exp.(*query.Result[purchaseorderapp.PurchaseOrder])

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				// Sort both by OrderNumber for comparison (database sorts lexicographically)
				sort.Slice(gotResp.Items, func(i, j int) bool {
					return gotResp.Items[i].OrderNumber < gotResp.Items[j].OrderNumber
				})
				sort.Slice(expResp.Items, func(i, j int) bool {
					return expResp.Items[i].OrderNumber < expResp.Items[j].OrderNumber
				})

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
			URL:        "/v1/procurement/purchase-orders/" + sd.PurchaseOrders[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &purchaseorderapp.PurchaseOrder{},
			ExpResp:    &sd.PurchaseOrders[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryIsUndelivered200(sd apitest.SeedData) []apitest.Table {
	// seed_test marks POs[8] and POs[9] as delivered — 8 are undelivered, 2 are delivered.
	table := []apitest.Table{
		{
			Name:       "is-undelivered-true",
			URL:        "/v1/procurement/purchase-orders?page=1&rows=10&orderBy=orderNumber,ASC&isUndelivered=true",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[purchaseorderapp.PurchaseOrder]{},
			ExpResp: &query.Result[purchaseorderapp.PurchaseOrder]{
				Page:        1,
				RowsPerPage: 10,
				Total:       8,
				Items:       sd.PurchaseOrders[:8],
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[purchaseorderapp.PurchaseOrder])
				expResp := exp.(*query.Result[purchaseorderapp.PurchaseOrder])

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				sort.Slice(gotResp.Items, func(i, j int) bool {
					return gotResp.Items[i].OrderNumber < gotResp.Items[j].OrderNumber
				})
				sort.Slice(expResp.Items, func(i, j int) bool {
					return expResp.Items[i].OrderNumber < expResp.Items[j].OrderNumber
				})

				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "is-undelivered-false",
			URL:        "/v1/procurement/purchase-orders?page=1&rows=10&orderBy=orderNumber,ASC&isUndelivered=false",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[purchaseorderapp.PurchaseOrder]{},
			ExpResp: &query.Result[purchaseorderapp.PurchaseOrder]{
				Page:        1,
				RowsPerPage: 10,
				Total:       2,
				Items:       sd.PurchaseOrders[8:],
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[purchaseorderapp.PurchaseOrder])
				expResp := exp.(*query.Result[purchaseorderapp.PurchaseOrder])

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				sort.Slice(gotResp.Items, func(i, j int) bool {
					return gotResp.Items[i].OrderNumber < gotResp.Items[j].OrderNumber
				})
				sort.Slice(expResp.Items, func(i, j int) bool {
					return expResp.Items[i].OrderNumber < expResp.Items[j].OrderNumber
				})

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByDeliveryDateRange200(sd apitest.SeedData) []apitest.Table {
	// POs[8] and POs[9] have actual_delivery_date = 2024-06-01.
	table := []apitest.Table{
		{
			Name:       "within-range",
			URL:        "/v1/procurement/purchase-orders?page=1&rows=10&orderBy=orderNumber,ASC&startActualDeliveryDate=2024-01-01+00%3A00%3A00+%2B0000+UTC&endActualDeliveryDate=2024-12-31+00%3A00%3A00+%2B0000+UTC",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[purchaseorderapp.PurchaseOrder]{},
			ExpResp: &query.Result[purchaseorderapp.PurchaseOrder]{
				Page:        1,
				RowsPerPage: 10,
				Total:       2,
				Items:       sd.PurchaseOrders[8:],
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[purchaseorderapp.PurchaseOrder])
				expResp := exp.(*query.Result[purchaseorderapp.PurchaseOrder])

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				sort.Slice(gotResp.Items, func(i, j int) bool {
					return gotResp.Items[i].OrderNumber < gotResp.Items[j].OrderNumber
				})
				sort.Slice(expResp.Items, func(i, j int) bool {
					return expResp.Items[i].OrderNumber < expResp.Items[j].OrderNumber
				})

				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "after-range",
			URL:        "/v1/procurement/purchase-orders?page=1&rows=10&orderBy=orderNumber,ASC&startActualDeliveryDate=2030-01-01+00%3A00%3A00+%2B0000+UTC",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[purchaseorderapp.PurchaseOrder]{},
			ExpResp: &query.Result[purchaseorderapp.PurchaseOrder]{
				Page:        1,
				RowsPerPage: 10,
				Total:       0,
				Items:       []purchaseorderapp.PurchaseOrder{},
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByIDs200(sd apitest.SeedData) []apitest.Table {
	ids := []string{sd.PurchaseOrders[0].ID, sd.PurchaseOrders[1].ID}
	expected := purchaseorderapp.PurchaseOrders{sd.PurchaseOrders[0], sd.PurchaseOrders[1]}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-orders/batch",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "POST",
			Input: &purchaseorderapp.QueryByIDsRequest{
				IDs: ids,
			},
			GotResp: &purchaseorderapp.PurchaseOrders{},
			ExpResp: &expected,
			CmpFunc: func(got, exp any) string {

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

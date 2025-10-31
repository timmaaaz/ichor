package purchaseorderlineitemapi_test

import (
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func query200(sd apitest.SeedData) []apitest.Table {
	// Sort all items by ID to match database ordering
	allItems := make([]purchaseorderlineitemapp.PurchaseOrderLineItem, len(sd.PurchaseOrderLineItems))
	copy(allItems, sd.PurchaseOrderLineItems)
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].ID < allItems[j].ID
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-items?page=1&rows=10&orderBy=id,ASC",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[purchaseorderlineitemapp.PurchaseOrderLineItem]{},
			ExpResp: &query.Result[purchaseorderlineitemapp.PurchaseOrderLineItem]{
				Page:        1,
				RowsPerPage: 10,
				Total:       20,
				Items:       allItems[:10],
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[purchaseorderlineitemapp.PurchaseOrderLineItem])
				expResp := exp.(*query.Result[purchaseorderlineitemapp.PurchaseOrderLineItem])

				dbtest.NormalizeJSONFields(gotResp, &expResp)

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
			URL:        "/v1/procurement/purchase-order-line-items/" + sd.PurchaseOrderLineItems[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &purchaseorderlineitemapp.PurchaseOrderLineItem{},
			ExpResp:    &sd.PurchaseOrderLineItems[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByIDs200(sd apitest.SeedData) []apitest.Table {
	ids := []string{sd.PurchaseOrderLineItems[0].ID, sd.PurchaseOrderLineItems[1].ID}
	expected := purchaseorderlineitemapp.PurchaseOrderLineItems{sd.PurchaseOrderLineItems[0], sd.PurchaseOrderLineItems[1]}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-items/batch",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "POST",
			Input: &purchaseorderlineitemapp.QueryByIDsRequest{
				IDs: ids,
			},
			GotResp: &purchaseorderlineitemapp.PurchaseOrderLineItems{},
			ExpResp: &expected,
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByPurchaseOrderID200(sd apitest.SeedData) []apitest.Table {
	// Find all line items for the first purchase order
	var expectedItems []purchaseorderlineitemapp.PurchaseOrderLineItem
	for _, item := range sd.PurchaseOrderLineItems {
		if item.PurchaseOrderID == sd.PurchaseOrders[0].ID {
			expectedItems = append(expectedItems, item)
		}
	}

	expected := purchaseorderlineitemapp.PurchaseOrderLineItems(expectedItems)

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/purchase-order-line-items/purchase-order/" + sd.PurchaseOrders[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &purchaseorderlineitemapp.PurchaseOrderLineItems{},
			ExpResp:    &expected,
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*purchaseorderlineitemapp.PurchaseOrderLineItems)
				expResp := exp.(*purchaseorderlineitemapp.PurchaseOrderLineItems)

				// Sort both by ID for comparison
				sort.Slice(*gotResp, func(i, j int) bool {
					return (*gotResp)[i].ID < (*gotResp)[j].ID
				})
				sort.Slice(*expResp, func(i, j int) bool {
					return (*expResp)[i].ID < (*expResp)[j].ID
				})

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

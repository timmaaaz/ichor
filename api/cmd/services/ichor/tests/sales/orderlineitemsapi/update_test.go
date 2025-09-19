package orderlineitemapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/order/order-line-items/%s", sd.OrderLineItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &orderlineitemsapp.UpdateOrderLineItem{
				Quantity: dbtest.StringPointer("2"),
			},
			GotResp: &orderlineitemsapp.OrderLineItem{},
			ExpResp: &orderlineitemsapp.OrderLineItem{
				ID:                            sd.OrderLineItems[0].ID,
				OrderID:                       sd.OrderLineItems[0].OrderID,
				ProductID:                     sd.OrderLineItems[0].ProductID,
				Quantity:                      "2",
				Discount:                      sd.OrderLineItems[0].Discount,
				LineItemFulfillmentStatusesID: sd.OrderLineItems[0].LineItemFulfillmentStatusesID,
				CreatedBy:                     sd.OrderLineItems[0].CreatedBy,
				CreatedDate:                   sd.OrderLineItems[0].CreatedDate,
				UpdatedBy:                     sd.OrderLineItems[0].UpdatedBy,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*orderlineitemsapp.OrderLineItem)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*orderlineitemsapp.OrderLineItem)

				gotResp.UpdatedDate = expResp.UpdatedDate
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid order id",
			URL:        fmt.Sprintf("/v1/order/order-line-items/%s", sd.OrderLineItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.UpdateOrderLineItem{
				OrderID: dbtest.StringPointer("invalid-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"order_id\",\"error\":\"order_id must be a valid version 4 UUID\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
		{
			Name:       "invalid product id",
			URL:        fmt.Sprintf("/v1/order/order-line-items/%s", sd.OrderLineItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.UpdateOrderLineItem{
				ProductID: dbtest.StringPointer("invalid-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id must be a valid version 4 UUID\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
		{
			Name:       "invalid quantity",
			URL:        fmt.Sprintf("/v1/order/order-line-items/%s", sd.OrderLineItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.UpdateOrderLineItem{
				Quantity: dbtest.StringPointer("invalid-quantity"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quantity\",\"error\":\"quantity must be a valid numeric value\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
		{
			Name:       "invalid line item fulfillment status id",
			URL:        fmt.Sprintf("/v1/order/order-line-items/%s", sd.OrderLineItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.UpdateOrderLineItem{
				LineItemFulfillmentStatusesID: dbtest.StringPointer("invalid-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"line_item_fulfillment_statuses_id\",\"error\":\"line_item_fulfillment_statuses_id must be a valid version 4 UUID\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/order/order-line-items/%s", sd.OrderLineItems[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/order/order-line-items/%s", sd.OrderLineItems[0].ID),
			Token:      sd.Admins[0].Token + "bad",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        fmt.Sprintf("/v1/order/order-line-items/%s", sd.OrderLineItems[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: order_line_items"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

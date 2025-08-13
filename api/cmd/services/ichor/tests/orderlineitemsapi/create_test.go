package orderlineitemapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/order/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &orderlineitemsapp.NewOrderLineItem{
				OrderID:                       sd.Orders[0].ID,
				ProductID:                     sd.Products[0].ProductID,
				Quantity:                      "1",
				Discount:                      "0",
				LineItemFulfillmentStatusesID: sd.LineItemFulfillmentStatuses[0].ID,
				CreatedBy:                     sd.Admins[0].ID.String(),
			},
			GotResp: &orderlineitemsapp.OrderLineItem{},
			ExpResp: &orderlineitemsapp.OrderLineItem{
				OrderID:                       sd.Orders[0].ID,
				ProductID:                     sd.Products[0].ProductID,
				Quantity:                      "1",
				Discount:                      "0.00",
				LineItemFulfillmentStatusesID: sd.LineItemFulfillmentStatuses[0].ID,
				CreatedBy:                     sd.Admins[0].ID.String(),
				UpdatedBy:                     sd.Admins[0].ID.String(),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*orderlineitemsapp.OrderLineItem)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*orderlineitemsapp.OrderLineItem)
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing order id",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.NewOrderLineItem{
				ProductID:                     sd.Products[0].ProductID,
				Quantity:                      "1",
				Discount:                      "0",
				LineItemFulfillmentStatusesID: sd.LineItemFulfillmentStatuses[0].ID,
				CreatedBy:                     sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"order_id\",\"error\":\"order_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing product id",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.NewOrderLineItem{
				OrderID:                       sd.Orders[0].ID,
				Quantity:                      "1",
				Discount:                      "0",
				LineItemFulfillmentStatusesID: sd.LineItemFulfillmentStatuses[0].ID,
				CreatedBy:                     sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing quantity",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.NewOrderLineItem{
				OrderID:                       sd.Orders[0].ID,
				ProductID:                     sd.Products[0].ProductID,
				Discount:                      "0",
				LineItemFulfillmentStatusesID: sd.LineItemFulfillmentStatuses[0].ID,
				CreatedBy:                     sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quantity\",\"error\":\"quantity is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing line item fulfillment status id",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.NewOrderLineItem{
				OrderID:   sd.Orders[0].ID,
				ProductID: sd.Products[0].ProductID,
				Quantity:  "1",
				Discount:  "0",
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"line_item_fulfillment_statuses_id\",\"error\":\"line_item_fulfillment_statuses_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing created by",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &orderlineitemsapp.NewOrderLineItem{
				OrderID:                       sd.Orders[0].ID,
				ProductID:                     sd.Products[0].ProductID,
				Quantity:                      "1",
				Discount:                      "0",
				LineItemFulfillmentStatusesID: sd.LineItemFulfillmentStatuses[0].ID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"created_by\",\"error\":\"created_by is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
	}
	return table
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        "/v1/order/orderlineitems",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badtoken",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "wronguser",
			URL:        "/v1/order/orderlineitems",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: order_line_items"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

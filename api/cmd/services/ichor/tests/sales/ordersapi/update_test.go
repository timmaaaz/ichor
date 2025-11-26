package ordersapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/sales/orders/%s", sd.Orders[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &ordersapp.UpdateOrder{
				CustomerID: &sd.Customers[1].ID,
			},
			GotResp: &ordersapp.Order{},
			ExpResp: &ordersapp.Order{
				ID:                  sd.Orders[0].ID,
				Number:              sd.Orders[0].Number,
				CustomerID:          sd.Customers[1].ID,
				FulfillmentStatusID: sd.Orders[0].FulfillmentStatusID,
				CreatedBy:           sd.Orders[0].CreatedBy,
				UpdatedBy:           sd.Orders[0].UpdatedBy,
				DueDate:             sd.Orders[0].DueDate,
				CreatedDate:         sd.Orders[0].CreatedDate,
				UpdatedDate:         sd.Orders[0].UpdatedDate,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ordersapp.Order)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*ordersapp.Order)
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid customer id",
			URL:        fmt.Sprintf("/v1/sales/orders/%s", sd.Orders[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.UpdateOrder{
				CustomerID: dbtest.StringPointer("invalid-id"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"customer_id\",\"error\":\"customer_id must be a valid version 4 UUID\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
		{
			Name:       "invalid fulfillment status id",
			URL:        fmt.Sprintf("/v1/sales/orders/%s", sd.Orders[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.UpdateOrder{
				FulfillmentStatusID: dbtest.StringPointer("invalid-id"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"fulfillment_status_id\",\"error\":\"fulfillment_status_id must be a valid version 4 UUID\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
		{
			Name:       "invalid updated by",
			URL:        fmt.Sprintf("/v1/sales/orders/%s", sd.Orders[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.UpdateOrder{
				UpdatedBy: dbtest.StringPointer("invalid-id"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"updated_by\",\"error\":\"updated_by must be a valid version 4 UUID\"}]"),
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
			URL:        fmt.Sprintf("/v1/sales/orders/%s", sd.Orders[0].ID),
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
			URL:        fmt.Sprintf("/v1/sales/orders/%s", sd.Orders[0].ID),
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
			URL:        fmt.Sprintf("/v1/sales/orders/%s", sd.Orders[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: sales.orders"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

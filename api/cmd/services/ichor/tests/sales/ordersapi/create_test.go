package ordersapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/sales/orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &ordersapp.NewOrder{
				Number:              "ORD-12345",
				CustomerID:          sd.Customers[0].ID,
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CurrencyID:          sd.Currencies[0].ID.String(),
				CreatedBy:           sd.Admins[0].ID.String(),
				DueDate:             time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
				OrderDate:           time.Now().Format("2006-01-02"),
			},
			GotResp: &ordersapp.Order{},
			ExpResp: &ordersapp.Order{
				Number:              "ORD-12345",
				CustomerID:          sd.Customers[0].ID,
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CurrencyID:          sd.Currencies[0].ID.String(),
				CreatedBy:           sd.Admins[0].ID.String(),
				UpdatedBy:           sd.Admins[0].ID.String(),
				DueDate:             time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
				OrderDate:           time.Now().Format("2006-01-02"),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ordersapp.Order)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*ordersapp.Order)
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
			Name:       "missing number",
			URL:        "/v1/sales/orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.NewOrder{
				CustomerID:          sd.Customers[0].ID,
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CurrencyID:          sd.Currencies[0].ID.String(),
				CreatedBy:           sd.Admins[0].ID.String(),
				DueDate:             time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
				OrderDate:           time.Now().Format("2006-01-02"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"number\",\"error\":\"number is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing customer id",
			URL:        "/v1/sales/orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.NewOrder{
				Number:              "ORD-12345",
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CurrencyID:          sd.Currencies[0].ID.String(),
				CreatedBy:           sd.Admins[0].ID.String(),
				DueDate:             time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
				OrderDate:           time.Now().Format("2006-01-02"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"customer_id\",\"error\":\"customer_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing due date",
			URL:        "/v1/sales/orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.NewOrder{
				Number:              "ORD-12345",
				CustomerID:          sd.Customers[0].ID,
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CurrencyID:          sd.Currencies[0].ID.String(),
				CreatedBy:           sd.Admins[0].ID.String(),
				OrderDate:           time.Now().Format("2006-01-02"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"due_date\",\"error\":\"due_date is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing fulfillment status id",
			URL:        "/v1/sales/orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.NewOrder{
				Number:     "ORD-12345",
				CustomerID: sd.Customers[0].ID,
				CurrencyID: sd.Currencies[0].ID.String(),
				CreatedBy:  sd.Admins[0].ID.String(),
				DueDate:    time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
				OrderDate:  time.Now().Format("2006-01-02"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"order_fulfillment_status_id\",\"error\":\"order_fulfillment_status_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing currency id",
			URL:        "/v1/sales/orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.NewOrder{
				Number:              "ORD-12345",
				CustomerID:          sd.Customers[0].ID,
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CreatedBy:           sd.Admins[0].ID.String(),
				DueDate:             time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
				OrderDate:           time.Now().Format("2006-01-02"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"currency_id\",\"error\":\"currency_id is a required field\"}]"),
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
			URL:        "/v1/sales/orders",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &ordersapp.NewOrder{
				Number:              "ORD-12345",
				CustomerID:          sd.Customers[0].ID,
				FulfillmentStatusID: sd.OrderFulfillmentStatuses[0].ID,
				CurrencyID:          sd.Currencies[0].ID.String(),
				DueDate:             time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02"),
				OrderDate:           time.Now().Format("2006-01-02"),
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
			URL:        "/v1/sales/orders",
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
			URL:        "/v1/sales/orders",
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
			URL:        "/v1/sales/orders",
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
			URL:        "/v1/sales/orders",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: sales.orders"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

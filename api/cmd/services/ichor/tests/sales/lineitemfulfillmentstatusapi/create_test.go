package lineitemfulfillmentstatusapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/lineitemfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/sales/line-item-fulfillment-statuses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &lineitemfulfillmentstatusapp.NewLineItemFulfillmentStatus{
				Name:        "NEW STATUS",
				Description: "NEW STATUS DESCRIPTION",
			},
			GotResp: &lineitemfulfillmentstatusapp.LineItemFulfillmentStatus{},
			ExpResp: &lineitemfulfillmentstatusapp.LineItemFulfillmentStatus{
				Name:        "NEW STATUS",
				Description: "NEW STATUS DESCRIPTION",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*lineitemfulfillmentstatusapp.LineItemFulfillmentStatus)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*lineitemfulfillmentstatusapp.LineItemFulfillmentStatus)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "with colors and icon",
			URL:        "/v1/sales/line-item-fulfillment-statuses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &lineitemfulfillmentstatusapp.NewLineItemFulfillmentStatus{
				Name:           "COLORED STATUS",
				Description:    "STATUS WITH COLORS",
				PrimaryColor:   "#FF5733",
				SecondaryColor: "#33FF57",
				Icon:           "check-circle",
			},
			GotResp: &lineitemfulfillmentstatusapp.LineItemFulfillmentStatus{},
			ExpResp: &lineitemfulfillmentstatusapp.LineItemFulfillmentStatus{
				Name:           "COLORED STATUS",
				Description:    "STATUS WITH COLORS",
				PrimaryColor:   "#FF5733",
				SecondaryColor: "#33FF57",
				Icon:           "check-circle",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*lineitemfulfillmentstatusapp.LineItemFulfillmentStatus)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*lineitemfulfillmentstatusapp.LineItemFulfillmentStatus)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing name",
			URL:        "/v1/sales/line-item-fulfillment-statuses",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lineitemfulfillmentstatusapp.NewLineItemFulfillmentStatus{
				Description: "Missing name field",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
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
			URL:        "/v1/sales/line-item-fulfillment-statuses",
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
			URL:        "/v1/sales/line-item-fulfillment-statuses",
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
			URL:        "/v1/sales/line-item-fulfillment-statuses",
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
			URL:        "/v1/sales/line-item-fulfillment-statuses",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: sales.line_item_fulfillment_statuses"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

package customersapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/customersapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/customers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &customersapp.NewCustomers{
				Name:              "Test Create Customer",
				ContactID:         sd.ContactInfos[0].ID,
				DeliveryAddressID: sd.Streets[0].ID,
				Notes:             "Testing input notes",
				CreatedBy:         sd.Admins[0].ID.String(),
			},
			GotResp: &customersapp.Customers{},
			ExpResp: &customersapp.Customers{
				Name:              "Test Create Customer",
				ContactID:         sd.ContactInfos[0].ID,
				DeliveryAddressID: sd.Streets[0].ID,
				Notes:             "Testing input notes",
				CreatedBy:         sd.Admins[0].ID.String(),
				UpdatedBy:         sd.Admins[0].ID.String(),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*customersapp.Customers)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*customersapp.Customers)
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
			Name:       "missing name",
			URL:        "/v1/core/customers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &customersapp.Customers{
				ContactID:         sd.ContactInfos[0].ID,
				DeliveryAddressID: sd.Streets[0].ID,
				Notes:             "Testing input notes",
				CreatedBy:         sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing contact id",
			URL:        "/v1/core/customers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &customersapp.Customers{
				Name:              "Test Create Customer",
				DeliveryAddressID: sd.Streets[0].ID,
				Notes:             "Testing input notes",
				CreatedBy:         sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"contact_id\",\"error\":\"contact_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing delivery address",
			URL:        "/v1/core/customers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &customersapp.Customers{
				Name:      "Test Create Customer",
				ContactID: sd.ContactInfos[0].ID,
				Notes:     "Testing input notes",
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"delivery_address_id\",\"error\":\"delivery_address_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing created by",
			URL:        "/v1/core/customers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &customersapp.Customers{
				Name:              "Test Create Customer",
				ContactID:         sd.ContactInfos[0].ID,
				DeliveryAddressID: sd.Streets[0].ID,
				Notes:             "Testing input notes",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"created_by\",\"error\":\"created_by is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/core/customers",
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
			Name:       "bad token",
			URL:        "/v1/core/customers",
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
			Name:       "bad sig",
			URL:        "/v1/core/customers",
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
			Name:       "roleadminonly",
			URL:        "/v1/core/customers",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: customers"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

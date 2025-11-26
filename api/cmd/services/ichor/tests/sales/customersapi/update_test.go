package customersapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/customersapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/sales/customers/%s", sd.Customers[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &customersapp.UpdateCustomers{
				Name: dbtest.StringPointer("Updated Customer Name"),
			},
			GotResp: &customersapp.Customers{},
			ExpResp: &customersapp.Customers{
				ID:                sd.Customers[1].ID,
				Name:              "Updated Customer Name",
				ContactID:         sd.Customers[1].ContactID,
				DeliveryAddressID: sd.Customers[1].DeliveryAddressID,
				Notes:             sd.Customers[1].Notes,
				CreatedBy:         sd.Admins[0].ID.String(),
				UpdatedBy:         sd.Admins[0].ID.String(),
				CreatedDate:       sd.Customers[1].CreatedDate,
				UpdatedDate:       sd.Customers[1].UpdatedDate,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*customersapp.Customers)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*customersapp.Customers)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	// generate 501 character string
	longString := "aa"
	for i := 0; i < 500; i++ {
		longString += "a"
	}

	table := []apitest.Table{
		{
			Name:       "bad-name",
			URL:        "/v1/sales/customers/abc",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &customersapp.UpdateCustomers{
				Name: dbtest.StringPointer("a"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"name","error":"name must be at least 3 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-contact-id",
			URL:        "/v1/sales/customers/" + sd.Customers[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &customersapp.UpdateCustomers{
				ContactID: dbtest.StringPointer("abc"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"contact_id","error":"contact_id must be a valid version 4 UUID"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-delivery-address-id",
			URL:        "/v1/sales/customers/" + sd.Customers[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &customersapp.UpdateCustomers{
				DeliveryAddressID: dbtest.StringPointer("abc"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"delivery_address_id","error":"delivery_address_id must be a valid version 4 UUID"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-notes",
			URL:        "/v1/sales/customers/" + sd.Customers[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &customersapp.UpdateCustomers{
				Notes: dbtest.StringPointer(longString),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"notes","error":"notes must be a maximum of 500 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/sales/customers/%s", sd.Customers[0].ID),
			Token:      "&nbsp",
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
			URL:        fmt.Sprintf("/v1/sales/customers/%s", sd.Customers[0].ID),
			Token:      sd.Users[0].Token + "A",
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
			URL:        fmt.Sprintf("/v1/sales/customers/%s", sd.Customers[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: sales.customers"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

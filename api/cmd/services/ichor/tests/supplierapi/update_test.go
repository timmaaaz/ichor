package supplierapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/supplier/supplierapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/suppliers/%s", sd.Suppliers[0].SupplierID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &supplierapp.UpdateSupplier{
				ContactInfoID: &sd.ContactInfo[0].ID,
				Name:          dbtest.StringPointer("UpdateSupplier"),
				PaymentTerms:  dbtest.StringPointer("UpdatePaymentTerms"),
				LeadTimeDays:  dbtest.StringPointer("10"),
				Rating:        dbtest.StringPointer("2.22"),
				IsActive:      dbtest.StringPointer("true"),
			},
			GotResp: &supplierapp.Supplier{},
			ExpResp: &supplierapp.Supplier{
				ContactInfoID: sd.ContactInfo[0].ID,
				Name:          "UpdateSupplier",
				PaymentTerms:  "UpdatePaymentTerms",
				LeadTimeDays:  "10",
				Rating:        "2.22",
				IsActive:      "true",
				CreatedDate:   sd.Suppliers[0].CreatedDate,
				SupplierID:    sd.Suppliers[0].SupplierID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*supplierapp.Supplier)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*supplierapp.Supplier)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-contact-uuid",
			URL:        fmt.Sprintf("/v1/suppliers/%s", sd.Suppliers[0].SupplierID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.UpdateSupplier{
				ContactInfoID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"contact_infos_id","error":"contact_infos_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-cost-uuid",
			URL:        fmt.Sprintf("/v1/suppliers/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.UpdateSupplier{
				ContactInfoID: &sd.Suppliers[0].ContactInfoID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/suppliers/%s", sd.Suppliers[0].SupplierID),
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
			URL:        fmt.Sprintf("/v1/suppliers/%s", sd.Suppliers[0].SupplierID),
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
			URL:        fmt.Sprintf("/v1/suppliers/%s", sd.Suppliers[0].SupplierID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: suppliers"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "supplier-dne",
			URL:        fmt.Sprintf("/v1/suppliers/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &supplierapp.UpdateSupplier{
				ContactInfoID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "supplier not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "contact-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/suppliers/%s", sd.Suppliers[0].SupplierID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &supplierapp.UpdateSupplier{
				ContactInfoID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

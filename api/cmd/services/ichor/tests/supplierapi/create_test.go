package supplierapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/supplier/supplierapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &supplierapp.NewSupplier{
				ContactInfoID: sd.ContactInfo[0].ID,
				Name:          "NewName",
				PaymentTerms:  "NewPaymentTerms",
				LeadTimeDays:  "8",
				Rating:        "4.6",
				IsActive:      "true",
			},
			GotResp: &supplierapp.Supplier{},
			ExpResp: &supplierapp.Supplier{
				ContactInfoID: sd.ContactInfo[0].ID,
				Name:          "NewName",
				PaymentTerms:  "NewPaymentTerms",
				LeadTimeDays:  "8",
				Rating:        "4.6",
				IsActive:      "true",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*supplierapp.Supplier)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*supplierapp.Supplier)
				expResp.SupplierID = gotResp.SupplierID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-contact-id",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.NewSupplier{
				Name:         "NewName",
				PaymentTerms: "NewPaymentTerms",
				LeadTimeDays: "8",
				Rating:       "4.6",
				IsActive:     "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"contact_info_id\",\"error\":\"contact_info_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-name",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.NewSupplier{
				ContactInfoID: sd.ContactInfo[0].ID,
				PaymentTerms:  "NewPaymentTerms",
				LeadTimeDays:  "8",
				Rating:        "4.6",
				IsActive:      "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-payment-terms",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.NewSupplier{
				ContactInfoID: sd.ContactInfo[0].ID,
				Name:          "NewName",
				LeadTimeDays:  "8",
				Rating:        "4.6",
				IsActive:      "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"payment_terms\",\"error\":\"payment_terms is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-lead-time-days",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.NewSupplier{
				ContactInfoID: sd.ContactInfo[0].ID,
				Name:          "NewName",
				PaymentTerms:  "NewPaymentTerms",
				Rating:        "4.6",
				IsActive:      "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lead_time_days\",\"error\":\"lead_time_days is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-rating",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.NewSupplier{
				ContactInfoID: sd.ContactInfo[0].ID,
				Name:          "NewName",
				PaymentTerms:  "NewPaymentTerms",
				LeadTimeDays:  "8",
				IsActive:      "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"rating\",\"error\":\"rating is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-is-active",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.NewSupplier{
				ContactInfoID: sd.ContactInfo[0].ID,
				Name:          "NewName",
				PaymentTerms:  "NewPaymentTerms",
				LeadTimeDays:  "8",
				Rating:        "4.6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"is_active\",\"error\":\"is_active is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-id",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierapp.Supplier{
				ContactInfoID: "not-a-uuid",
				Name:          "NewName",
				PaymentTerms:  "NewPaymentTerms",
				LeadTimeDays:  "8",
				Rating:        "4.6",
				IsActive:      "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"contact_info_id\",\"error\":\"contact_info_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "contact-info-not-valid-fk",
			URL:        "/v1/suppliers",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &supplierapp.NewSupplier{
				ContactInfoID: uuid.New().String(),
				Name:          "NewName",
				PaymentTerms:  "NewPaymentTerms",
				LeadTimeDays:  "8",
				Rating:        "4.6",
				IsActive:      "true",
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/suppliers",
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
			URL:        "/v1/suppliers",
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
			URL:        "/v1/suppliers",
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
			URL:        "/v1/suppliers",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: suppliers"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

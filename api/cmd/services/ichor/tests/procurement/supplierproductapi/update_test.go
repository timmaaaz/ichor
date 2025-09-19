package supplierapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/procurement/supplierproductapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", sd.SupplierProducts[0].SupplierProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &supplierproductapp.UpdateSupplierProduct{
				SupplierID:         &sd.Suppliers[2].SupplierID,
				ProductID:          &sd.Products[2].ProductID,
				SupplierPartNumber: dbtest.StringPointer("UpdateSupplierPartNumber"),
				MinOrderQuantity:   dbtest.StringPointer("3"),
				MaxOrderQuantity:   dbtest.StringPointer("7"),
				LeadTimeDays:       dbtest.StringPointer("4"),
				UnitCost:           dbtest.StringPointer("10000.93"),
				IsPrimarySupplier:  dbtest.StringPointer("true"),
			},
			GotResp: &supplierproductapp.SupplierProduct{},
			ExpResp: &supplierproductapp.SupplierProduct{
				SupplierID:         sd.Suppliers[2].SupplierID,
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "UpdateSupplierPartNumber",
				MinOrderQuantity:   "3",
				MaxOrderQuantity:   "7",
				LeadTimeDays:       "4",
				UnitCost:           "10000.93",
				IsPrimarySupplier:  "true",
				CreatedDate:        sd.SupplierProducts[0].CreatedDate,
				SupplierProductID:  sd.SupplierProducts[0].SupplierProductID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*supplierproductapp.SupplierProduct)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*supplierproductapp.SupplierProduct)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "supplier-uuid",
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", sd.SupplierProducts[0].SupplierProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.UpdateSupplierProduct{
				SupplierID: dbtest.StringPointer("not-a-uuid"),
				ProductID:  &sd.Products[0].ProductID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"supplier_id","error":"supplier_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-uuid",
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", sd.SupplierProducts[0].SupplierProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.UpdateSupplierProduct{
				ProductID:  dbtest.StringPointer("not-a-uuid"),
				SupplierID: &sd.Suppliers[0].SupplierID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-supplier-product-id-uuid",
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.UpdateSupplierProduct{
				SupplierID: &sd.SupplierProducts[0].SupplierID,
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
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", sd.SupplierProducts[0].SupplierProductID),
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
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", sd.SupplierProducts[0].SupplierProductID),
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
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", sd.SupplierProducts[0].SupplierProductID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: supplier_products"),
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
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &supplierproductapp.UpdateSupplierProduct{
				SupplierID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "supplierProduct not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "supplier-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", sd.SupplierProducts[0].SupplierProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &supplierproductapp.UpdateSupplierProduct{
				SupplierID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/procurement/supplier-products/%s", sd.SupplierProducts[0].SupplierProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &supplierproductapp.UpdateSupplierProduct{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

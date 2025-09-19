package inventoryproductapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/products/productapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/products/products/%s", sd.Products[1].ProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &productapp.UpdateProduct{
				Name:                 dbtest.StringPointer("Test Product"),
				SKU:                  dbtest.StringPointer("sku123"),
				BrandID:              dbtest.StringPointer(sd.Brands[0].ID),
				ProductCategoryID:    dbtest.StringPointer(sd.ProductCategories[0].ID),
				Description:          dbtest.StringPointer("test description"),
				ModelNumber:          dbtest.StringPointer("test model number"),
				UpcCode:              dbtest.StringPointer("test upc code"),
				Status:               dbtest.StringPointer("test status"),
				IsActive:             dbtest.StringPointer("true"),
				IsPerishable:         dbtest.StringPointer("false"),
				HandlingInstructions: dbtest.StringPointer("test handling instructions"),
				UnitsPerCase:         dbtest.StringPointer("20"),
			},
			GotResp: &productapp.Product{},
			ExpResp: &productapp.Product{
				ProductID:            sd.Products[1].ProductID,
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				ProductCategoryID:    sd.ProductCategories[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
				CreatedDate:          sd.Products[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*productapp.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productapp.Product)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-brand-uuid",
			URL:        fmt.Sprintf("/v1/products/products/%s", sd.Products[0].ProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.UpdateProduct{
				BrandID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"brand_id","error":"brand_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-category-uuid",
			URL:        fmt.Sprintf("/v1/products/products/%s", sd.Products[0].ProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.UpdateProduct{
				ProductCategoryID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_category_id","error":"product_category_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-brand-uuid",
			URL:        fmt.Sprintf("/v1/products/products/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input:      &productapp.UpdateProduct{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
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
			URL:        fmt.Sprintf("/v1/products/products/%s", sd.Products[0].ProductID),
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
			URL:        fmt.Sprintf("/v1/products/products/%s", sd.Products[0].ProductID),
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
			URL:        fmt.Sprintf("/v1/products/products/%s", sd.Products[0].ProductID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: products"),
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
			Name:       "brand-dne",
			URL:        fmt.Sprintf("/v1/products/products/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input:      &productapp.UpdateProduct{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "product not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "brand-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/products/products/%s", sd.Products[0].ProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &productapp.UpdateProduct{
				BrandID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-category-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/products/products/%s", sd.Products[0].ProductID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &productapp.UpdateProduct{
				ProductCategoryID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

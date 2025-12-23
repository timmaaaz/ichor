package inventoryproductapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/products/productapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &productapp.NewProduct{
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
			},
			GotResp: &productapp.Product{},
			ExpResp: &productapp.Product{
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
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*productapp.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productapp.Product)
				expResp.ProductID = gotResp.ProductID
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
			Name:       "missing-name",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-sku",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"sku\",\"error\":\"sku is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-brand-id",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				ProductCategoryID:    sd.ProductCategories[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"brand_id\",\"error\":\"brand_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-brand-id",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              "not a uuid",
				ProductCategoryID:    sd.ProductCategories[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"brand_id","error":"brand_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-product-category",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_category_id\",\"error\":\"product_category_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-category",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				ProductCategoryID:    "not-a-uuid",
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_category_id","error":"product_category_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-description",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				ProductCategoryID:    sd.ProductCategories[0].ID,
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"description\",\"error\":\"description is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-upc-code",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				ProductCategoryID:    sd.ProductCategories[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"upc_code\",\"error\":\"upc_code is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-status",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				ProductCategoryID:    sd.ProductCategories[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"status\",\"error\":\"status is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-is-active",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				ProductCategoryID:    sd.ProductCategories[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"is_active\",\"error\":\"is_active is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-is-perishable",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				ProductCategoryID:    sd.ProductCategories[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"is_perishable\",\"error\":\"is_perishable is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-handling-units-per-case",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productapp.NewProduct{
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"units_per_case\",\"error\":\"units_per_case is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "brand-is-not-valid-fk",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              uuid.NewString(),
				ProductCategoryID:    sd.ProductCategories[0].ID,
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-category-id-is-not-valid-fk",
			URL:        "/v1/products/products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &productapp.NewProduct{
				Name:                 "Test Product",
				SKU:                  "sku123",
				BrandID:              sd.Brands[0].ID,
				ProductCategoryID:    uuid.NewString(),
				Description:          "test description",
				ModelNumber:          "test model number",
				UpcCode:              "test upc code",
				Status:               "test status",
				IsActive:             "true",
				IsPerishable:         "false",
				HandlingInstructions: "test handling instructions",
				UnitsPerCase:         "20",
			},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
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
			URL:        "/v1/products/products",
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
			URL:        "/v1/products/products",
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
			URL:        "/v1/products/products",
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
			URL:        "/v1/products/products",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: products.products"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

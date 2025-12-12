package supplierapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/procurement/supplierproductapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			GotResp: &supplierproductapp.SupplierProduct{},
			ExpResp: &supplierproductapp.SupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*supplierproductapp.SupplierProduct)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*supplierproductapp.SupplierProduct)
				expResp.SupplierProductID = gotResp.SupplierProductID
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
			Name:       "missing-supplier-id",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"supplier_id\",\"error\":\"supplier_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-product-id",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-supplier-part-number",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:        sd.Suppliers[0].SupplierID,
				ProductID:         sd.Products[2].ProductID,
				MinOrderQuantity:  "10",
				MaxOrderQuantity:  "25",
				LeadTimeDays:      "7",
				UnitCost:          "12.90",
				IsPrimarySupplier: "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"supplier_part_number\",\"error\":\"supplier_part_number is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-min-order-quantity",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"min_order_quantity\",\"error\":\"min_order_quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-max-order-quantity",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"max_order_quantity\",\"error\":\"max_order_quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-lead-time-days",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lead_time_days\",\"error\":\"lead_time_days is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-unit-cost",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				IsPrimarySupplier:  "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"unit_cost\",\"error\":\"unit_cost is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-is-primary-supplier",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"is_primary_supplier\",\"error\":\"is_primary_supplier is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-supplier-id",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         "not-a-uuid",
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"supplier_id\",\"error\":\"supplier_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-id",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          "not-a-uuid",
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "supplier-id-not-valid-fk",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         uuid.New().String(),
				ProductID:          sd.Products[2].ProductID,
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
			},
			ExpResp: errs.Newf(errs.Aborted, "foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-id-not-valid-fk",
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &supplierproductapp.NewSupplierProduct{
				SupplierID:         sd.Suppliers[0].SupplierID,
				ProductID:          uuid.New().String(),
				SupplierPartNumber: "SupplierPartNumber",
				MinOrderQuantity:   "10",
				MaxOrderQuantity:   "25",
				LeadTimeDays:       "7",
				UnitCost:           "12.90",
				IsPrimarySupplier:  "true",
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
			URL:        "/v1/procurement/supplier-products",
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
			URL:        "/v1/procurement/supplier-products",
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
			URL:        "/v1/procurement/supplier-products",
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
			URL:        "/v1/procurement/supplier-products",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: procurement.supplier_products"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

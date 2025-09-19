package productcostapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/products/productcostapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func create200(sd apitest.SeedData) []apitest.Table {

	now := time.Now().Format(timeutil.FORMAT)

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &productcostapp.ProductCost{},
			ExpResp: &productcostapp.ProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*productcostapp.ProductCost)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productcostapp.ProductCost)
				expResp.ID = gotResp.ID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	now := time.Now().Format(timeutil.FORMAT)
	return []apitest.Table{
		{
			Name:       "missing-product-id",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-purchase-cost",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"purchase_cost\",\"error\":\"purchase_cost is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-selling-price",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"selling_price\",\"error\":\"selling_price is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-currency",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"currency\",\"error\":\"currency is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-msrp",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"msrp\",\"error\":\"msrp is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-markup-percentage",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"markup_percentage\",\"error\":\"markup_percentage is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-landed-cost",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"landed_cost\",\"error\":\"landed_cost is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-carrying-cost",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"carrying_cost\",\"error\":\"carrying_cost is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-abc-classification",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"abc_classification\",\"error\":\"abc_classification is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-depreciation-value",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"depreciation_value\",\"error\":\"depreciation_value is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-insurance-value",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				EffectiveDate:     now,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"insurance_value\",\"error\":\"insurance_value is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-effective-date",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         sd.Products[0].ProductID,
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"effective_date\",\"error\":\"effective_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-id",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.NewProductCost{
				ProductID:         "not a uuid",
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     now,
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
			Name:       "contact-info-not-valid-fk",
			URL:        "/v1/products/product-costs",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &productcostapp.NewProductCost{
				ProductID:         uuid.NewString(),
				PurchaseCost:      "20.16",
				SellingPrice:      "30.76",
				Currency:          "USD",
				MSRP:              "44.44",
				MarkupPercentage:  "12.678",
				LandedCost:        "220.6",
				CarryingCost:      "5.4",
				ABCClassification: "B",
				DepreciationValue: "1.23",
				InsuranceValue:    "77.77",
				EffectiveDate:     time.Now().Format(timeutil.FORMAT),
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
			URL:        "/v1/products/product-costs",
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
			URL:        "/v1/products/product-costs",
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
			URL:        "/v1/products/product-costs",
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
			URL:        "/v1/products/product-costs",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: product_costs"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

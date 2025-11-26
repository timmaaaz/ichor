package productcostapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/products/productcostapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/products/product-costs/%s", sd.ProductCosts[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &productcostapp.UpdateProductCost{
				ProductID:         &sd.ProductCosts[1].ProductID,
				PurchaseCost:      &sd.ProductCosts[2].PurchaseCost,
				SellingPrice:      &sd.ProductCosts[3].SellingPrice,
				Currency:          &sd.ProductCosts[4].Currency,
				MSRP:              &sd.ProductCosts[5].MSRP,
				MarkupPercentage:  &sd.ProductCosts[6].MarkupPercentage,
				LandedCost:        &sd.ProductCosts[7].LandedCost,
				CarryingCost:      &sd.ProductCosts[8].CarryingCost,
				ABCClassification: &sd.ProductCosts[9].ABCClassification,
				DepreciationValue: &sd.ProductCosts[10].DepreciationValue,
				InsuranceValue:    &sd.ProductCosts[11].InsuranceValue,
				EffectiveDate:     &sd.ProductCosts[12].EffectiveDate,
			},
			GotResp: &productcostapp.ProductCost{},
			ExpResp: &productcostapp.ProductCost{
				ProductID:         sd.ProductCosts[1].ProductID,
				PurchaseCost:      sd.ProductCosts[2].PurchaseCost,
				SellingPrice:      sd.ProductCosts[3].SellingPrice,
				Currency:          sd.ProductCosts[4].Currency,
				MSRP:              sd.ProductCosts[5].MSRP,
				MarkupPercentage:  sd.ProductCosts[6].MarkupPercentage,
				LandedCost:        sd.ProductCosts[7].LandedCost,
				CarryingCost:      sd.ProductCosts[8].CarryingCost,
				ABCClassification: sd.ProductCosts[9].ABCClassification,
				DepreciationValue: sd.ProductCosts[10].DepreciationValue,
				InsuranceValue:    sd.ProductCosts[11].InsuranceValue,
				EffectiveDate:     sd.ProductCosts[12].EffectiveDate,
				CreatedDate:       sd.ProductCosts[0].CreatedDate,
				ID:                sd.ProductCosts[0].ID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*productcostapp.ProductCost)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*productcostapp.ProductCost)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-product-uuid",
			URL:        fmt.Sprintf("/v1/products/product-costs/%s", sd.ProductCosts[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.UpdateProductCost{
				ProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-cost-uuid",
			URL:        fmt.Sprintf("/v1/products/product-costs/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &productcostapp.UpdateProductCost{
				ProductID: &sd.ProductCosts[0].ID,
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
			URL:        fmt.Sprintf("/v1/products/product-costs/%s", sd.ProductCosts[0].ID),
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
			URL:        fmt.Sprintf("/v1/products/product-costs/%s", sd.ProductCosts[0].ID),
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
			URL:        fmt.Sprintf("/v1/products/product-costs/%s", sd.ProductCosts[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: products.product_costs"),
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
			Name:       "product-dne",
			URL:        fmt.Sprintf("/v1/products/product-costs/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &productcostapp.UpdateProductCost{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "productCost not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/products/product-costs/%s", sd.ProductCosts[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &productcostapp.UpdateProductCost{
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

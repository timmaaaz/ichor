package productcostapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/finance/productcostapp"
)

func parseQueryParams(r *http.Request) (productcostapp.QueryParams, error) {
	values := r.URL.Query()

	filter := productcostapp.QueryParams{
		Page:              values.Get("page"),
		Rows:              values.Get("rows"),
		OrderBy:           values.Get("orderBy"),
		ID:                values.Get("product_cost_id"),
		ProductID:         values.Get("product_id"),
		PurchaseCost:      values.Get("purchase_cost"),
		CreatedDate:       values.Get("created_date"),
		UpdatedDate:       values.Get("updated_date"),
		Currency:          values.Get("currency"),
		SellingPrice:      values.Get("selling_price"),
		MSRP:              values.Get("msrp"),
		MarkupPercentage:  values.Get("markup_percentage"),
		LandedCost:        values.Get("landed_cost"),
		CarryingCost:      values.Get("carrying_cost"),
		ABCClassification: values.Get("abc_classification"),
		DepreciationValue: values.Get("depreciation_value"),
		InsuranceValue:    values.Get("insurance_value"),
		EffectiveDate:     values.Get("effective_date"),
	}

	return filter, nil
}

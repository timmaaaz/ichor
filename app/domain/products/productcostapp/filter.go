package productcostapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (productcostbus.QueryFilter, error) {
	var filter productcostbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}
	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id
	}

	if qp.CurrencyID != "" {
		id, err := uuid.Parse(qp.CurrencyID)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("currency_id", err)
		}
		filter.CurrencyID = &id
	}

	if qp.EffectiveDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EffectiveDate)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("effective_date", err)
		}
		filter.EffectiveDate = &t
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &t
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &t
	}

	if qp.ABCClassification != "" {
		filter.ABCClassification = &qp.ABCClassification
	}

	if qp.PurchaseCost != "" {
		purchaseCost, err := types.ParseMoney(qp.PurchaseCost)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("purchase_cost", err)
		}
		filter.PurchaseCost = &purchaseCost
	}

	if qp.SellingPrice != "" {
		sellingPrice, err := types.ParseMoney(qp.SellingPrice)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("selling_price", err)
		}
		filter.SellingPrice = &sellingPrice
	}

	if qp.MSRP != "" {
		msrp, err := types.ParseMoney(qp.MSRP)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("msrp", err)
		}
		filter.MSRP = &msrp
	}

	if qp.MarkupPercentage != "" {
		markupPercentage, err := types.ParseRoundedFloat(qp.MarkupPercentage)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("markup_percentage", err)
		}
		filter.MarkupPercentage = &markupPercentage
	}

	if qp.LandedCost != "" {
		landedCost, err := types.ParseMoney(qp.LandedCost)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("landed_cost", err)
		}

		filter.LandedCost = &landedCost
	}

	if qp.CarryingCost != "" {
		carryingCost, err := types.ParseMoney(qp.CarryingCost)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("carrying_cost", err)
		}
		filter.CarryingCost = &carryingCost
	}

	if qp.DepreciationValue != "" {
		depreciationValue, err := types.ParseRoundedFloat(qp.DepreciationValue)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("depreciation_value", err)
		}
		filter.DepreciationValue = &depreciationValue
	}

	if qp.InsuranceValue != "" {
		insuranceValue, err := types.ParseMoney(qp.InsuranceValue)
		if err != nil {
			return productcostbus.QueryFilter{}, errs.NewFieldsError("insuranceValue", err)
		}

		filter.InsuranceValue = &insuranceValue
	}

	return filter, nil
}

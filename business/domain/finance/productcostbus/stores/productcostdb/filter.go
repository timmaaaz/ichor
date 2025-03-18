package productcostdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/finance/productcostbus"
)

func applyFilter(filter productcostbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["cost_id"] = *filter.ID
		wc = append(wc, "cost_id = :cost_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.PurchaseCost != nil {
		data["purchase_cost"] = *filter.PurchaseCost
		wc = append(wc, "purchase_cost = :purchase_cost")
	}

	if filter.SellingPrice != nil {
		data["selling_price"] = *filter.SellingPrice
		wc = append(wc, "selling_price = :selling_price")
	}

	if filter.Currency != nil {
		data["currency"] = *filter.Currency
		wc = append(wc, "currency = :currency")
	}

	if filter.MSRP != nil {
		data["msrp"] = *filter.MSRP
		wc = append(wc, "msrp = :msrp")
	}

	if filter.MarkupPercentage != nil {
		data["markup_percentage"] = *filter.MarkupPercentage
		wc = append(wc, "markup_percentage = :markup_percentage")
	}

	if filter.LandedCost != nil {
		data["landed_cost"] = *filter.LandedCost
		wc = append(wc, "landed_cost = :landed_cost")
	}

	if filter.CarryingCost != nil {
		data["carrying_cost"] = *filter.CarryingCost
		wc = append(wc, "carrying_cost = :carrying_cost")
	}

	if filter.ABCClassification != nil {
		data["abc_classification"] = *filter.ABCClassification
		wc = append(wc, "abc_classification = :abc_classification")
	}

	if filter.DepreciationValue != nil {
		data["depreciation_value"] = *filter.DepreciationValue
		wc = append(wc, "depreciation_value = :depreciation_value")
	}

	if filter.InsuranceValue != nil {
		data["insurance_value"] = *filter.InsuranceValue
		wc = append(wc, "insurance_value = :insurance_value")
	}
	if filter.EffectiveDate != nil {
		data["effective_date"] = *filter.EffectiveDate
		wc = append(wc, "effective_date = :effective_date")
	}
	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}
	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}
	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}

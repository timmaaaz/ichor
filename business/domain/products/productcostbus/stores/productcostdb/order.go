package productcostdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	productcostbus.OrderByID:                "id",
	productcostbus.OrderByProductID:         "product_id",
	productcostbus.OrderByPurchaseCost:      "purchase_cost",
	productcostbus.OrderBySellingPrice:      "selling_price",
	productcostbus.OrderByCurrencyID:        "currency_id",
	productcostbus.OrderByMSRP:              "msrp",
	productcostbus.OrderByMarkupPercentage:  "markup_percentage",
	productcostbus.OrderByLandedCost:        "landed_cost",
	productcostbus.OrderByCarryingCost:      "carrying_cost",
	productcostbus.OrderByABCClassification: "abc_classification",
	productcostbus.OrderByDepreciationValue: "depreciation_value",
	productcostbus.OrderByInsuranceValue:    "insurance_value",
	productcostbus.OrderByEffectiveDate:     "effective_date",
	productcostbus.OrderByCreatedDate:       "created_date",
	productcostbus.OrderByUpdatedDate:       "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

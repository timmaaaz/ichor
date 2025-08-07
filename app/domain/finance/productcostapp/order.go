package productcostapp

import (
	"github.com/timmaaaz/ichor/business/domain/finance/productcostbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("product_id", order.ASC)

var orderByFields = map[string]string{
	"id":                 productcostbus.OrderByID,
	"product_id":         productcostbus.OrderByProductID,
	"purchase_cost":      productcostbus.OrderByPurchaseCost,
	"selling_price":      productcostbus.OrderBySellingPrice,
	"currency":           productcostbus.OrderByCurrency,
	"msrp":               productcostbus.OrderByMSRP,
	"markup_percentage":  productcostbus.OrderByMarkupPercentage,
	"landed_cost":        productcostbus.OrderByLandedCost,
	"carrying_cost":      productcostbus.OrderByCarryingCost,
	"abc_classification": productcostbus.OrderByABCClassification,
	"depreciation_value": productcostbus.OrderByDepreciationValue,
	"insurance_value":    productcostbus.OrderByInsuranceValue,
	"effective_date":     productcostbus.OrderByEffectiveDate,
	"created_date":       productcostbus.OrderByCreatedDate,
	"updated_date":       productcostbus.OrderByUpdatedDate,
}

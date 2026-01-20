package productcostbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByProductID, order.ASC)

const (
	OrderByID                = "id"
	OrderByProductID         = "product_id"
	OrderByPurchaseCost      = "purchase_cost"
	OrderBySellingPrice      = "selling_price"
	OrderByCurrencyID        = "currency_id"
	OrderByMSRP              = "msrp"
	OrderByMarkupPercentage  = "markup_percentage"
	OrderByLandedCost        = "landed_cost"
	OrderByCarryingCost      = "carrying_cost"
	OrderByABCClassification = "abc_classification"
	OrderByDepreciationValue = "depreciation_value"
	OrderByInsuranceValue    = "insurance_value"
	OrderByEffectiveDate     = "effective_date"
	OrderByCreatedDate       = "created_date"
	OrderByUpdatedDate       = "updated_date"
)

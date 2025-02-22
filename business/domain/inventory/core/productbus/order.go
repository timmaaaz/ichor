package productbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderByProductID            = "product_id"
	OrderBySKU                  = "sku"
	OrderByBrandID              = "brand_id"
	OrderByProductCategoryID    = "category_id"
	OrderByName                 = "name"
	OrderByDescription          = "description"
	OrderByModelNumber          = "model_number"
	OrderByUpcCode              = "upc_code"
	OrderByStatus               = "status"
	OrderByIsActive             = "is_active"
	OrderByIsPerishable         = "is_perishable"
	OrderByHandlingInstructions = "handling_instructions"
	OrderByUnitsPerCase         = "units_per_case"
	OrderByCreatedDate          = "created_date"
	OrderByUpdatedDate          = "updated_date"
)

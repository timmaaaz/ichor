package productapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":                    productbus.OrderByProductID,
	"sku":                   productbus.OrderBySKU,
	"brand_id":              productbus.OrderByBrandID,
	"category_id":           productbus.OrderByProductCategoryID,
	"name":                  productbus.OrderByName,
	"description":           productbus.OrderByDescription,
	"model_number":          productbus.OrderByModelNumber,
	"upc_code":              productbus.OrderByUpcCode,
	"status":                productbus.OrderByStatus,
	"is_active":             productbus.OrderByIsActive,
	"is_perishable":         productbus.OrderByIsPerishable,
	"handling_instructions": productbus.OrderByHandlingInstructions,
	"units_per_case":        productbus.OrderByUnitsPerCase,
	"created_date":          productbus.OrderByCreatedDate,
	"updated_date":          productbus.OrderByUpdatedDate,
}

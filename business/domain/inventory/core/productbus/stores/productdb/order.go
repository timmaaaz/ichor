package productdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	productbus.OrderByProductID:            "id",
	productbus.OrderBySKU:                  "sku",
	productbus.OrderByBrandID:              "brand_id",
	productbus.OrderByProductCategoryID:    "category_id",
	productbus.OrderByName:                 "name",
	productbus.OrderByDescription:          "description",
	productbus.OrderByModelNumber:          "model_number",
	productbus.OrderByUpcCode:              "upc_code",
	productbus.OrderByStatus:               "status",
	productbus.OrderByIsActive:             "is_active",
	productbus.OrderByIsPerishable:         "is_perishable",
	productbus.OrderByHandlingInstructions: "handling_instructions",
	productbus.OrderByUnitsPerCase:         "units_per_case",
	productbus.OrderByCreatedDate:          "created_date",
	productbus.OrderByUpdatedDate:          "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

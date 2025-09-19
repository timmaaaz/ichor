package supplierproductdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	supplierproductbus.OrderBySupplierProductID:  "id",
	supplierproductbus.OrderBySupplierID:         "supplier_id",
	supplierproductbus.OrderByProductID:          "product_id",
	supplierproductbus.OrderBySupplierPartNumber: "supplier_part_number",
	supplierproductbus.OrderByMinOrderQuantity:   "min_order_quantity",
	supplierproductbus.OrderByMaxOrderQuantity:   "max_order_quantity",
	supplierproductbus.OrderByLeadTimeDays:       "lead_time_days",
	supplierproductbus.OrderByUnitCost:           "unit_cost",
	supplierproductbus.OrderByIsPrimarySupplier:  "is_primary_supplier",
	supplierproductbus.OrderByCreatedDate:        "created_date",
	supplierproductbus.OrderByUpdatedDate:        "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

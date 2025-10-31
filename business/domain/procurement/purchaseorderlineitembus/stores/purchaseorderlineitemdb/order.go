package purchaseorderlineitemdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	purchaseorderlineitembus.OrderByID:                   "id",
	purchaseorderlineitembus.OrderByPurchaseOrderID:      "purchase_order_id",
	purchaseorderlineitembus.OrderBySupplierProductID:    "supplier_product_id",
	purchaseorderlineitembus.OrderByQuantityOrdered:      "quantity_ordered",
	purchaseorderlineitembus.OrderByQuantityReceived:     "quantity_received",
	purchaseorderlineitembus.OrderByLineTotal:            "line_total",
	purchaseorderlineitembus.OrderByExpectedDeliveryDate: "expected_delivery_date",
	purchaseorderlineitembus.OrderByActualDeliveryDate:   "actual_delivery_date",
	purchaseorderlineitembus.OrderByCreatedDate:          "created_date",
	purchaseorderlineitembus.OrderByUpdatedDate:          "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
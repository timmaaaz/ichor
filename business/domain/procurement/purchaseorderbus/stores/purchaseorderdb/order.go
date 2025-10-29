package purchaseorderdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	purchaseorderbus.OrderByID:                   "id",
	purchaseorderbus.OrderByOrderNumber:          "order_number",
	purchaseorderbus.OrderBySupplierID:           "supplier_id",
	purchaseorderbus.OrderByOrderDate:            "order_date",
	purchaseorderbus.OrderByExpectedDeliveryDate: "expected_delivery_date",
	purchaseorderbus.OrderByActualDeliveryDate:   "actual_delivery_date",
	purchaseorderbus.OrderByTotalAmount:          "total_amount",
	purchaseorderbus.OrderByRequestedBy:          "requested_by",
	purchaseorderbus.OrderByApprovedBy:           "approved_by",
	purchaseorderbus.OrderByCreatedDate:          "created_date",
	purchaseorderbus.OrderByUpdatedDate:          "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
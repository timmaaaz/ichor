package ordersdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	ordersbus.OrderByID:                         "id",
	ordersbus.OrderByNumber:                     "number",
	ordersbus.OrderByCustomerID:                 "customer_id",
	ordersbus.OrderByDueDate:                    "due_date",
	ordersbus.OrderByOrderByFulfillmentStatusID: "order_fulfillment_status_id",
	ordersbus.OrderByOrderDate:                  "order_date",
	ordersbus.OrderBySubtotal:                   "subtotal",
	ordersbus.OrderByTotalAmount:                "total_amount",
	ordersbus.OrderByCurrency:                   "currency",
	ordersbus.OrderByCreatedBy:                  "created_by",
	ordersbus.OrderByUpdatedBy:                  "updated_by",
	ordersbus.OrderByCreatedDate:                "created_date",
	ordersbus.OrderByUpdatedDate:                "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

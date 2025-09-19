package orderlineitemsdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	orderlineitemsbus.OrderByID:                            "id",
	orderlineitemsbus.OrderByOrderID:                       "order_id",
	orderlineitemsbus.OrderByProductID:                     "product_id",
	orderlineitemsbus.OrderByQuantity:                      "quantity",
	orderlineitemsbus.OrderByDiscount:                      "discount",
	orderlineitemsbus.OrderByLineItemFulfillmentStatusesID: "line_item_fulfillment_statuses_id",
	orderlineitemsbus.OrderByCreatedBy:                     "created_by",
	orderlineitemsbus.OrderByCreatedDate:                   "created_date",
	orderlineitemsbus.OrderByUpdatedBy:                     "updated_by",
	orderlineitemsbus.OrderByUpdatedDate:                   "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package picktaskapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	picktaskbus.OrderByID:             "id",
	picktaskbus.OrderBySalesOrderID:   "sales_order_id",
	picktaskbus.OrderByProductID:      "product_id",
	picktaskbus.OrderByLocationID:     "location_id",
	picktaskbus.OrderByQuantityToPick: "quantity_to_pick",
	picktaskbus.OrderByStatus:         "status",
	picktaskbus.OrderByAssignedTo:     "assigned_to",
	picktaskbus.OrderByCreatedBy:      "created_by",
	picktaskbus.OrderByCreatedDate:    "created_date",
	picktaskbus.OrderByUpdatedDate:    "updated_date",
}

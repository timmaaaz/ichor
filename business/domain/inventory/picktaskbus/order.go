package picktaskbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for pick task queries.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID             = "id"
	OrderBySalesOrderID   = "sales_order_id"
	OrderByProductID      = "product_id"
	OrderByLocationID     = "location_id"
	OrderByQuantityToPick = "quantity_to_pick"
	OrderByStatus         = "status"
	OrderByAssignedTo     = "assigned_to"
	OrderByCreatedBy      = "created_by"
	OrderByCreatedDate    = "created_date"
	OrderByUpdatedDate    = "updated_date"
)

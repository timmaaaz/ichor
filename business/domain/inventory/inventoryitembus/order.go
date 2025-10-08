package inventoryitembus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID                    = "id"
	OrderByProductID             = "product_id"
	OrderByLocationID            = "location_id"
	OrderByQuantity              = "quantity"
	OrderByReservedQuantity      = "reserved_quantity"
	OrderByAllocatedQuantity     = "allocated_quantity"
	OrderByMinimumStock          = "minimum_stock"
	OrderByMaximumStock          = "maximum_stock"
	OrderByReorderPoint          = "reorder_point"
	OrderByEconomicOrderQuantity = "economic_order_quantity"
	OrderBySafetyStock           = "safety_stock"
	OrderByAvgDailyUsage         = "avg_daily_usage"
	OrderByCreatedDate           = "created_date"
	OrderByUpdatedDate           = "updated_date"
)

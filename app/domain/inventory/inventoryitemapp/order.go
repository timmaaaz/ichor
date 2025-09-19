package inventoryitemapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":                      inventoryitembus.OrderByItemID,
	"product_id":              inventoryitembus.OrderByProductID,
	"location_id":             inventoryitembus.OrderByLocationID,
	"quantity":                inventoryitembus.OrderByQuantity,
	"reserved_quantity":       inventoryitembus.OrderByReservedQuantity,
	"allocated_quantity":      inventoryitembus.OrderByAllocatedQuantity,
	"minimum_stock":           inventoryitembus.OrderByMinimumStock,
	"maximum_stock":           inventoryitembus.OrderByMaximumStock,
	"reorder_point":           inventoryitembus.OrderByReorderPoint,
	"economic_order_quantity": inventoryitembus.OrderByEconomicOrderQuantity,
	"safety_stock":            inventoryitembus.OrderBySafetyStock,
	"avg_daily_usage":         inventoryitembus.OrderByAvgDailyUsage,
	"created_date":            inventoryitembus.OrderByCreatedDate,
	"updated_date":            inventoryitembus.OrderByUpdatedDate,
}

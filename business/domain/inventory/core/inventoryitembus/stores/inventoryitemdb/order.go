package inventoryitemdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	inventoryitembus.OrderByItemID:                "id",
	inventoryitembus.OrderByProductID:             "product_id",
	inventoryitembus.OrderByLocationID:            "location_id",
	inventoryitembus.OrderByReservedQuantity:      "reserved_quantity",
	inventoryitembus.OrderByAllocatedQuantity:     "allocated_quantity",
	inventoryitembus.OrderByMinimumStock:          "minimum_stock",
	inventoryitembus.OrderByMaximumStock:          "maximum_stock",
	inventoryitembus.OrderByReorderPoint:          "reorder_point",
	inventoryitembus.OrderByEconomicOrderQuantity: "economic_order_quantity",
	inventoryitembus.OrderBySafetyStock:           "safety_stock",
	inventoryitembus.OrderByAvgDailyUsage:         "avg_daily_usage",
	inventoryitembus.OrderByQuantity:              "quantity",
	inventoryitembus.OrderByCreatedDate:           "created_date",
	inventoryitembus.OrderByUpdatedDate:           "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

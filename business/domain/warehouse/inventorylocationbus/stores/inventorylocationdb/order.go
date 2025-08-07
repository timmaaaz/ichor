package inventorylocationdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	inventorylocationbus.OrderByLocationID:         "id",
	inventorylocationbus.OrderByWarehouseID:        "zone_id",
	inventorylocationbus.OrderByZoneID:             "warehouse_id",
	inventorylocationbus.OrderByAisle:              "aisle",
	inventorylocationbus.OrderByRack:               "rack",
	inventorylocationbus.OrderByShelf:              "shelf",
	inventorylocationbus.OrderByBin:                "bin",
	inventorylocationbus.OrderByIsPickLocation:     "is_pick_location",
	inventorylocationbus.OrderByIsReserveLocation:  "is_reserve_location",
	inventorylocationbus.OrderByMaxCapacity:        "max_capacity",
	inventorylocationbus.OrderByCurrentUtilization: "current_utilization",
	inventorylocationbus.OrderByCreatedDate:        "created_date",
	inventorylocationbus.OrderByUpdatedDate:        "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package zonedb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	zonebus.OrderByID:          "zone_id",
	zonebus.OrderByWarehouseID: "warehouse_id",
	zonebus.OrderByName:        "name",
	zonebus.OrderByIsActive:    "is_active",
	zonebus.OrderByDateCreated: "date_created",
	zonebus.OrderByDateUpdated: "date_updated",
	zonebus.OrderByCreatedBy:   "created_by",
	zonebus.OrderByUpdatedBy:   "updated_by",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

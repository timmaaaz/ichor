package warehousedb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	warehousebus.OrderByID:          "warehouse_id",
	warehousebus.OrderByStreetID:    "street_id",
	warehousebus.OrderByName:        "name",
	warehousebus.OrderByIsActive:    "is_active",
	warehousebus.OrderByDateCreated: "date_created",
	warehousebus.OrderByDateUpdated: "date_updated",
	warehousebus.OrderByCreatedBy:   "created_by",
	warehousebus.OrderByUpdatedBy:   "updated_by",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

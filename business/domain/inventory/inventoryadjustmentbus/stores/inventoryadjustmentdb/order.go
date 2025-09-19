package inventoryadjustmentdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	inventoryadjustmentbus.OrderByInventoryAdjustmentID: "id",
	inventoryadjustmentbus.OrderByProductID:             "product_id",
	inventoryadjustmentbus.OrderByLocationID:            "location_id",
	inventoryadjustmentbus.OrderByAdjustedBy:            "adjusted_by",
	inventoryadjustmentbus.OrderByApprovedBy:            "approved_by",
	inventoryadjustmentbus.OrderByQuantityChange:        "quantity_change",
	inventoryadjustmentbus.OrderByReasonCode:            "reason_code",
	inventoryadjustmentbus.OrderByNotes:                 "notes",
	inventoryadjustmentbus.OrderByAdjustmentDate:        "adjustment_date",
	inventoryadjustmentbus.OrderByCreatedDate:           "created_date",
	inventoryadjustmentbus.OrderByUpdatedDate:           "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package serialnumberdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	serialnumberbus.OrderBySerialID:     "id",
	serialnumberbus.OrderByLotID:        "lot_id",
	serialnumberbus.OrderByProductID:    "product_id",
	serialnumberbus.OrderByLocationID:   "location_id",
	serialnumberbus.OrderBySerialNumber: "serial_number",
	serialnumberbus.OrderByStatus:       "status",
	serialnumberbus.OrderByCreatedDate:  "created_date",
	serialnumberbus.OrderByUpdatedDate:  "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

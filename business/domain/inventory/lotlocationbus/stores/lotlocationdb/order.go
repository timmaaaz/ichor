package lotlocationdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	lotlocationbus.OrderByID:          "id",
	lotlocationbus.OrderByLotID:       "lot_id",
	lotlocationbus.OrderByLocationID:  "location_id",
	lotlocationbus.OrderByQuantity:    "quantity",
	lotlocationbus.OrderByCreatedDate: "created_date",
	lotlocationbus.OrderByUpdatedDate: "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

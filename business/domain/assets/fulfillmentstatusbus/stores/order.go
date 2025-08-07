package fulfillmentstatusdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	fulfillmentstatusbus.OrderByID:     "id",
	fulfillmentstatusbus.OrderByIconID: "icon_id",
	fulfillmentstatusbus.OrderByName:   "name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

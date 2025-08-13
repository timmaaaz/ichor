package orderfulfillmentstatusdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/order/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	orderfulfillmentstatusbus.OrderByID:          "id",
	orderfulfillmentstatusbus.OrderByName:        "name",
	orderfulfillmentstatusbus.OrderByDescription: "description",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

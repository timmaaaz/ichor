package purchaseorderlineitemstatusdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	purchaseorderlineitemstatusbus.OrderByID:          "id",
	purchaseorderlineitemstatusbus.OrderByName:        "name",
	purchaseorderlineitemstatusbus.OrderByDescription: "description",
	purchaseorderlineitemstatusbus.OrderBySortOrder:   "sort_order",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
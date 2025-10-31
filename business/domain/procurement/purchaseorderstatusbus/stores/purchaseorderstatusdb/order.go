package purchaseorderstatusdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	purchaseorderstatusbus.OrderByID:          "id",
	purchaseorderstatusbus.OrderByName:        "name",
	purchaseorderstatusbus.OrderByDescription: "description",
	purchaseorderstatusbus.OrderBySortOrder:   "sort_order",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

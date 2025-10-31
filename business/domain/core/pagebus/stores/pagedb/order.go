package pagedb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	pagebus.OrderByID:        "id",
	pagebus.OrderByPath:      "path",
	pagebus.OrderByName:      "name",
	pagebus.OrderByModule:    "module",
	pagebus.OrderBySortOrder: "sort_order",
	pagebus.OrderByIsActive:  "is_active",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

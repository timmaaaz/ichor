package currencydb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	currencybus.OrderByID:            "id",
	currencybus.OrderByCode:          "code",
	currencybus.OrderByName:          "name",
	currencybus.OrderBySortOrder:     "sort_order",
	currencybus.OrderByIsActive:      "is_active",
	currencybus.OrderByDecimalPlaces: "decimal_places",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

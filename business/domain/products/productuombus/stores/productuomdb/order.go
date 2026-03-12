package productuomdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	productuombus.OrderByID:          "id",
	productuombus.OrderByProductID:   "product_id",
	productuombus.OrderByName:        "name",
	productuombus.OrderByCreatedDate: "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

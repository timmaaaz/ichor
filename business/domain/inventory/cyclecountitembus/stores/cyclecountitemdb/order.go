package cyclecountitemdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	cyclecountitembus.OrderByID:             "id",
	cyclecountitembus.OrderBySessionID:      "session_id",
	cyclecountitembus.OrderByProductID:      "product_id",
	cyclecountitembus.OrderByLocationID:     "location_id",
	cyclecountitembus.OrderBySystemQuantity: "system_quantity",
	cyclecountitembus.OrderByStatus:         "status",
	cyclecountitembus.OrderByCreatedDate:    "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

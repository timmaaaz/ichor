package cyclecountsessiondb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	cyclecountsessionbus.OrderByID:          "id",
	cyclecountsessionbus.OrderByName:        "name",
	cyclecountsessionbus.OrderByStatus:      "status",
	cyclecountsessionbus.OrderByCreatedBy:   "created_by",
	cyclecountsessionbus.OrderByCreatedDate: "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

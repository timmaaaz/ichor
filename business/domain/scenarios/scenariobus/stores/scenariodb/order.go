package scenariodb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	scenariobus.OrderByID:          "id",
	scenariobus.OrderByName:        "name",
	scenariobus.OrderByCreatedDate: "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

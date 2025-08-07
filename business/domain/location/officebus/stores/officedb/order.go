package officedb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/location/officebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	officebus.OrderByID:       "id",
	officebus.OrderByName:     "name",
	officebus.OrderByStreetID: "street_id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

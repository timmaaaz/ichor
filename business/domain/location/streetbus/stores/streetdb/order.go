package streetdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	streetbus.OrderByID:         "id",
	streetbus.OrderByCityID:     "city_id",
	streetbus.OrderByLine1:      "line_1",
	streetbus.OrderByLine2:      "line_2",
	streetbus.OrderByPostalCode: "postal_code",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

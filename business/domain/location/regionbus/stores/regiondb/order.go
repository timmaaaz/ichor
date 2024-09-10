package regiondb

import (
	"fmt"

	"bitbucket.org/superiortechnologies/ichor/business/domain/location/regionbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	regionbus.OrderByID:        "region_id",
	regionbus.OrderByName:      "name",
	regionbus.OrderByCode:      "code",
	regionbus.OrderByCountryID: "country_id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package countrydb

import (
	"fmt"

	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	countrybus.OrderByID:     "country_id",
	countrybus.OrderByNumber: "number",
	countrybus.OrderByName:   "name",
	countrybus.OrderByAlpha2: "alpha_2",
	countrybus.OrderByAlpha3: "alpha_3",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package titledb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/users/titlebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	titlebus.OrderByID:          "title_id",
	titlebus.OrderByName:        "name",
	titlebus.OrderByDescription: "description",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

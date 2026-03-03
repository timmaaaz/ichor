package settingsdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	settingsbus.OrderByKey:         "key",
	settingsbus.OrderByCreatedDate: "created_date",
	settingsbus.OrderByUpdatedDate: "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}
	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

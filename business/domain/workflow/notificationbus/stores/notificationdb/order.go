package notificationdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	notificationbus.OrderByID:          "id",
	notificationbus.OrderByPriority:    "priority",
	notificationbus.OrderByIsRead:      "is_read",
	notificationbus.OrderByCreatedDate: "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

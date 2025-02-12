package userapprovalstatusdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	userapprovalstatusbus.OrderByID:     "user_approval_status_id",
	userapprovalstatusbus.OrderByIconID: "icon_id",
	userapprovalstatusbus.OrderByName:   "name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package approvalstatusdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	approvalstatusbus.OrderByID:     "approval_status_id",
	approvalstatusbus.OrderByIconID: "icon_id",
	approvalstatusbus.OrderByName:   "name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

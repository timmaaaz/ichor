package actionpermissionsdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	actionpermissionsbus.OrderByID:         "id",
	actionpermissionsbus.OrderByRoleID:     "role_id",
	actionpermissionsbus.OrderByActionType: "action_type",
	actionpermissionsbus.OrderByCreatedAt:  "created_at",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

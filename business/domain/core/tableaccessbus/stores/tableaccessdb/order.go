package tableaccessdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	tableaccessbus.OrderByID:        "id",
	tableaccessbus.OrderByRoleID:    "role_id",
	tableaccessbus.OrderByTableName: "table_name",
	tableaccessbus.OrderByCanCreate: "can_create",
	tableaccessbus.OrderByCanRead:   "can_read",
	tableaccessbus.OrderByCanUpdate: "can_update",
	tableaccessbus.OrderByCanDelete: "can_delete",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

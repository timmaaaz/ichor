package rolepagedb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	rolepagebus.OrderByID:        "id",
	rolepagebus.OrderByRoleID:    "role_id",
	rolepagebus.OrderByPageID:    "page_id",
	rolepagebus.OrderByCanAccess: "can_access",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

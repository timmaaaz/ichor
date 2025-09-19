package commentdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	commentbus.OrderByID:          "id",
	commentbus.OrderByComment:     "comment",
	commentbus.OrderByCommenterID: "commenter_id",
	commentbus.OrderByUserID:      "user_id",
	commentbus.OrderByCreatedDate: "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

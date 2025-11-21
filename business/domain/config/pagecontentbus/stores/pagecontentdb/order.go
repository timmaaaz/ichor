package pagecontentdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	pagecontentbus.OrderByID:          "id",
	pagecontentbus.OrderByOrderIndex:  "order_index",
	pagecontentbus.OrderByLabel:       "label",
	pagecontentbus.OrderByContentType: "content_type",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package pageactiondb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	pageactionbus.OrderByID:           "id",
	pageactionbus.OrderByPageConfigID: "page_config_id",
	pageactionbus.OrderByActionType:   "action_type",
	pageactionbus.OrderByActionOrder:  "action_order",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	// When ordering by action_order, group by page_config_id first for logical grouping
	// Always include id as final sort for deterministic ordering
	if orderBy.Field == pageactionbus.OrderByActionOrder {
		return fmt.Sprintf(" ORDER BY page_config_id ASC, %s %s, id ASC", by, orderBy.Direction), nil
	}

	// For other orderings, just add id as secondary sort
	return fmt.Sprintf(" ORDER BY %s %s, id ASC", by, orderBy.Direction), nil
}

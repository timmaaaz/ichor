package alertdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	alertbus.OrderByID:          "id",
	alertbus.OrderByAlertType:   "alert_type",
	alertbus.OrderBySeverity:    "severity",
	alertbus.OrderByStatus:      "status",
	alertbus.OrderByCreatedDate: "created_date",
	alertbus.OrderByUpdatedDate: "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

// orderByClauseWithPrefix generates an ORDER BY clause with table prefix for join queries.
func orderByClauseWithPrefix(orderBy order.By, prefix string) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + prefix + "." + by + " " + orderBy.Direction, nil
}

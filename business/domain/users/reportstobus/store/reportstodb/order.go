package reportstodb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	reportstobus.OrderByID:         "id",
	reportstobus.OrderByBossID:     "boss_id",
	reportstobus.OrderByReporterID: "reporter_id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}
	return "ORDER BY " + by + " " + orderBy.Direction, nil
}

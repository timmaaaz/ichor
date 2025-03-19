package costhistorydb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/finance/costhistorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	costhistorybus.OrderByCostHistoryID: "history_id",
	costhistorybus.OrderByProductID:     "product_id",
	costhistorybus.OrderByCostType:      "cost_type",
	costhistorybus.OrderByAmount:        "amount",
	costhistorybus.OrderByCurrency:      "currency",
	costhistorybus.OrderByEffectiveDate: "effective_date",
	costhistorybus.OrderByEndDate:       "end_date",
	costhistorybus.OrderByCreatedDate:   "created_date",
	costhistorybus.OrderByUpdatedDate:   "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

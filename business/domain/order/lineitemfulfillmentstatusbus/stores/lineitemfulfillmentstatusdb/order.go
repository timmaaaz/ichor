package lineitemfulfillmentstatusdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/order/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	lineitemfulfillmentstatusbus.LineItemByID:          "id",
	lineitemfulfillmentstatusbus.LineItemByName:        "name",
	lineitemfulfillmentstatusbus.LineItemByDescription: "description",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

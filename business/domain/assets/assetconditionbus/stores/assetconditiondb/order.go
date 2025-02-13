package assetconditiondb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	assetconditionbus.OrderByID:          "asset_condition_id",
	assetconditionbus.OrderByName:        "name",
	assetconditionbus.OrderByDescription: "description",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

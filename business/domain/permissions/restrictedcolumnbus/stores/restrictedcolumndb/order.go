package restrictedcolumndb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/permissions/restrictedcolumnbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	restrictedcolumnbus.OrderByID:         "restricted_column_id",
	restrictedcolumnbus.OrderByTableName:  "table_name",
	restrictedcolumnbus.OrderByColumnName: "column_name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

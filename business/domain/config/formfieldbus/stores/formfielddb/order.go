package formfielddb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	formfieldbus.OrderByID:         "id",
	formfieldbus.OrderByFormID:     "form_id",
	formfieldbus.OrderByName:       "name",
	formfieldbus.OrderByFieldOrder: "field_order",
	formfieldbus.OrderByFieldType:  "field_type",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
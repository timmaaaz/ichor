package branddb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	brandbus.OrderByID:             "brand_id",
	brandbus.OrderByName:           "name",
	brandbus.OrderByCreatedDate:    "created_date",
	brandbus.OrderByUpdatedDate:    "updated_date",
	brandbus.OrderByManufacturerID: "manufacturer_id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

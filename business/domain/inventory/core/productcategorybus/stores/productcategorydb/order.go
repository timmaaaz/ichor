package productcategorydb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	productcategorybus.OrderByID:          "product_category_id",
	productcategorybus.OrderByName:        "name",
	productcategorybus.OrderByDescription: "description",
	productcategorybus.OrderByCreatedDate: "created_date",
	productcategorybus.OrderByUpdatedDate: "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

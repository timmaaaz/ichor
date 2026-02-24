package putawaytaskdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	putawaytaskbus.OrderByID:              "id",
	putawaytaskbus.OrderByProductID:       "product_id",
	putawaytaskbus.OrderByLocationID:      "location_id",
	putawaytaskbus.OrderByQuantity:        "quantity",
	putawaytaskbus.OrderByReferenceNumber: "reference_number",
	putawaytaskbus.OrderByStatus:          "status",
	putawaytaskbus.OrderByAssignedTo:      "assigned_to",
	putawaytaskbus.OrderByCreatedBy:       "created_by",
	putawaytaskbus.OrderByCreatedDate:     "created_date",
	putawaytaskbus.OrderByUpdatedDate:     "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package transferorderdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/movement/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	transferorderbus.OrderByTransferID:     "transfer_id",
	transferorderbus.OrderByProductID:      "product_id",
	transferorderbus.OrderByFromLocationID: "from_location_id",
	transferorderbus.OrderByToLocationID:   "to_location_id",
	transferorderbus.OrderByRequestedByID:  "requested_by",
	transferorderbus.OrderByApprovedByID:   "approved_by",
	transferorderbus.OrderByQuantity:       "quantity",
	transferorderbus.OrderByStatus:         "status",
	transferorderbus.OrderByTransferDate:   "transfer_date",
	transferorderbus.OrderByCreatedDate:    "created_date",
	transferorderbus.OrderByUpdatedDate:    "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

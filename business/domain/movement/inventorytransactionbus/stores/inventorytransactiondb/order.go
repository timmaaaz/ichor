package inventorytransactiondb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	inventorytransactionbus.OrderByCreatedDate:            "created_date",
	inventorytransactionbus.OrderByUpdatedDate:            "updated_date",
	inventorytransactionbus.OrderByInventoryTransactionID: "transaction_id",
	inventorytransactionbus.OrderByProductID:              "product_id",
	inventorytransactionbus.OrderByQuantity:               "quantity",
	inventorytransactionbus.OrderByLocationID:             "location_id",
	inventorytransactionbus.OrderByTransactionType:        "transaction_type",
	inventorytransactionbus.OrderByReferenceNumber:        "reference_number",
	inventorytransactionbus.OrderByTransactionDate:        "transaction_date",
	inventorytransactionbus.OrderByUserID:                 "user_id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

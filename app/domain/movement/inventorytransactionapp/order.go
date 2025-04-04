package inventorytransactionapp

import (
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("transaction_id", order.ASC)

var orderByFields = map[string]string{
	"transaction_id":   inventorytransactionbus.OrderByInventoryTransactionID,
	"product_id":       inventorytransactionbus.OrderByProductID,
	"location_id":      inventorytransactionbus.OrderByLocationID,
	"user_id":          inventorytransactionbus.OrderByUserID,
	"quantity":         inventorytransactionbus.OrderByQuantity,
	"transaction_type": inventorytransactionbus.OrderByTransactionType,
	"reference_number": inventorytransactionbus.OrderByReferenceNumber,
	"transaction_date": inventorytransactionbus.OrderByTransactionDate,
	"created_date":     inventorytransactionbus.OrderByCreatedDate,
	"updated_date":     inventorytransactionbus.OrderByUpdatedDate,
}

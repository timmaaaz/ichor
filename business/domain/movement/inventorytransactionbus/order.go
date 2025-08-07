package inventorytransactionbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByInventoryTransactionID, order.ASC)

const (
	OrderByInventoryTransactionID = "id"
	OrderByProductID              = "product_id"
	OrderByLocationID             = "location_id"
	OrderByUserID                 = "user_id"
	OrderByQuantity               = "quantity"
	OrderByTransactionType        = "transaction_type"
	OrderByReferenceNumber        = "reference_number"
	OrderByTransactionDate        = "transaction_date"
	OrderByCreatedDate            = "created_date"
	OrderByUpdatedDate            = "updated_date"
)

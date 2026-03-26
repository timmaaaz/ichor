package cyclecountitembus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for cycle count item queries.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID             = "id"
	OrderBySessionID      = "session_id"
	OrderByProductID      = "product_id"
	OrderByLocationID     = "location_id"
	OrderBySystemQuantity = "system_quantity"
	OrderByStatus         = "status"
	OrderByCreatedDate    = "created_date"
)

package purchaseorderlineitemstatusbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderBySortOrder, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID          = "id"
	OrderByName        = "name"
	OrderByDescription = "description"
	OrderBySortOrder   = "sort_order"
)
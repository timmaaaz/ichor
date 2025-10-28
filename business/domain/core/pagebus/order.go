package pagebus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderBySortOrder, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID        = "id"
	OrderByPath      = "path"
	OrderByName      = "name"
	OrderByModule    = "module"
	OrderBySortOrder = "sort_order"
	OrderByIsActive  = "is_active"
)

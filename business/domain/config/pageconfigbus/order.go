package pageconfigbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy specifies the default ordering for page config queries
var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

// Ordering field constants for page config
const (
	OrderByID        = "id"
	OrderByName      = "name"
	OrderByIsDefault = "is_default"
)

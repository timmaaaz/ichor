package currencybus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderBySortOrder, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID            = "id"
	OrderByCode          = "code"
	OrderByName          = "name"
	OrderBySortOrder     = "sort_order"
	OrderByIsActive      = "is_active"
	OrderByDecimalPlaces = "decimal_places"
)

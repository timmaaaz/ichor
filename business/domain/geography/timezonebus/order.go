package timezonebus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByDisplayName, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID          = "id"
	OrderByName        = "name"
	OrderByDisplayName = "display_name"
	OrderByUTCOffset   = "utc_offset"
	OrderByIsActive    = "is_active"
)

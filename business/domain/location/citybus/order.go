package citybus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID       = "city_id"
	OrderByRegionID = "region_id"
	OrderByName     = "name"
)

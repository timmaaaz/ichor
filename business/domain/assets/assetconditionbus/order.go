package assetconditionbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID          = "asset_condition_id"
	OrderByName        = "name"
	OrderByDescription = "description"
)

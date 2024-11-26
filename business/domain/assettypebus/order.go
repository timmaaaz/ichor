package assettypebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

// Set of fields that can be ordered by
const (
	OrderByID   = "asset_type_id"
	OrderByName = "name"
)

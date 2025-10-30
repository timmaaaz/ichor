package pageactionbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByActionOrder, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID           = "id"
	OrderByPageConfigID = "page_config_id"
	OrderByActionType   = "action_type"
	OrderByActionOrder  = "action_order"
)

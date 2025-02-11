package userapprovalstatusbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID     = "user_approval_status_id"
	OrderByIconID = "icon_id"
	OrderByName   = "name"
)

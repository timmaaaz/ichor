package rolepagebus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID        = "id"
	OrderByRoleID    = "role_id"
	OrderByPageID    = "page_id"
	OrderByCanAccess = "can_access"
)

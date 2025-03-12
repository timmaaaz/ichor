package userrolebus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByUserID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID     = "user_role_id"
	OrderByUserID = "user_id"
	OrderByRoleID = "role_id"
)

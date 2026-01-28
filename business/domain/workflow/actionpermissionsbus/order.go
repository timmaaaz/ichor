package actionpermissionsbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering by action type ascending.
var DefaultOrderBy = order.NewBy(OrderByActionType, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID         = "id"
	OrderByRoleID     = "role_id"
	OrderByActionType = "action_type"
	OrderByCreatedAt  = "created_at"
)

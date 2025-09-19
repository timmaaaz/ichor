package tableaccessbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByTableName, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID        = "id"
	OrderByRoleID    = "role_id"
	OrderByTableName = "table_name"
	OrderByCanCreate = "can_create"
	OrderByCanRead   = "can_read"
	OrderByCanUpdate = "can_update"
	OrderByCanDelete = "can_delete"
)

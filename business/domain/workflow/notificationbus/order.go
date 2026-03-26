package notificationbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for notifications (newest first).
var DefaultOrderBy = order.NewBy(OrderByCreatedDate, order.DESC)

// Set of fields that can be used for ordering.
const (
	OrderByID          = "id"
	OrderByPriority    = "priority"
	OrderByIsRead      = "is_read"
	OrderByCreatedDate = "created_date"
)

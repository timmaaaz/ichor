package notificationapp

import (
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	"id":          notificationbus.OrderByID,
	"priority":    notificationbus.OrderByPriority,
	"isRead":      notificationbus.OrderByIsRead,
	"createdDate": notificationbus.OrderByCreatedDate,
}

// DefaultOrderBy is the default ordering for notification queries.
var DefaultOrderBy = order.NewBy(notificationbus.OrderByCreatedDate, order.DESC)

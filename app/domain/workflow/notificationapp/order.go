package notificationapp

import (
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	notificationbus.OrderByID:          notificationbus.OrderByID,
	notificationbus.OrderByPriority:    notificationbus.OrderByPriority,
	notificationbus.OrderByIsRead:      notificationbus.OrderByIsRead,
	notificationbus.OrderByCreatedDate: notificationbus.OrderByCreatedDate,
}

// DefaultOrderBy is the default ordering for notification queries.
var DefaultOrderBy = order.NewBy(notificationbus.OrderByCreatedDate, order.DESC)

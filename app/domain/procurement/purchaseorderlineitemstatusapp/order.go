package purchaseorderlineitemstatusapp

import (
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("sort_order", order.ASC)

var orderByFields = map[string]string{
	"id":          purchaseorderlineitemstatusbus.OrderByID,
	"name":        purchaseorderlineitemstatusbus.OrderByName,
	"description": purchaseorderlineitemstatusbus.OrderByDescription,
	"sortOrder":   purchaseorderlineitemstatusbus.OrderBySortOrder,
}

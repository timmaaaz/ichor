package purchaseorderstatusapp

import (
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("sort_order", order.ASC)

var orderByFields = map[string]string{
	"id":          purchaseorderstatusbus.OrderByID,
	"name":        purchaseorderstatusbus.OrderByName,
	"description": purchaseorderstatusbus.OrderByDescription,
	"sortOrder":   purchaseorderstatusbus.OrderBySortOrder,
}

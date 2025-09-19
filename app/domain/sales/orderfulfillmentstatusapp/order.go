package orderfulfillmentstatusapp

import (
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":          orderfulfillmentstatusbus.OrderByID,
	"name":        orderfulfillmentstatusbus.OrderByName,
	"description": orderfulfillmentstatusbus.OrderByDescription,
}

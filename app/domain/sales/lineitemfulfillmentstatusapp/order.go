package lineitemfulfillmentstatusapp

import (
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":          lineitemfulfillmentstatusbus.OrderByID,
	"name":        lineitemfulfillmentstatusbus.OrderByName,
	"description": lineitemfulfillmentstatusbus.OrderByDescription,
}

package paymenttermapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":          paymenttermbus.OrderByID,
	"name":        paymenttermbus.OrderByName,
	"description": paymenttermbus.OrderByDescription,
}

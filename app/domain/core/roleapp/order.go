package roleapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(rolebus.OrderByName, order.ASC)

var orderByFields = map[string]string{
	"id":          rolebus.OrderByID,
	"name":        rolebus.OrderByName,
	"description": rolebus.OrderByDescription,
}

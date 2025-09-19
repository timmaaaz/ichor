package titleapp

import (
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":          titlebus.OrderByID,
	"name":        titlebus.OrderByName,
	"description": titlebus.OrderByDescription,
}

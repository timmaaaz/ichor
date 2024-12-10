package titleapp

import (
	"github.com/timmaaaz/ichor/business/domain/titlebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"title_id":    titlebus.OrderByID,
	"name":        titlebus.OrderByName,
	"description": titlebus.OrderByDescription,
}

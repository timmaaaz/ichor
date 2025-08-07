package tagapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":          tagbus.OrderByID,
	"name":        tagbus.OrderByName,
	"description": tagbus.OrderByDescription,
}

package assetconditionapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":          assetconditionbus.OrderByID,
	"name":        assetconditionbus.OrderByName,
	"description": assetconditionbus.OrderByDescription,
}

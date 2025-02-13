package assetconditionapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"asset_condition_id": assetconditionbus.OrderByID,
	"name":               assetconditionbus.OrderByName,
	"description":        assetconditionbus.OrderByDescription,
}

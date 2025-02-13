package assettypeapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"asset_type_id": assettypebus.OrderByID,
	"name":          assettypebus.OrderByName,
	"description":   assettypebus.OrderByDescription,
}

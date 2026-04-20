package scenarioapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(scenariobus.OrderByName, order.ASC)

var orderByFields = map[string]string{
	"id":           scenariobus.OrderByID,
	"name":         scenariobus.OrderByName,
	"created_date": scenariobus.OrderByCreatedDate,
}

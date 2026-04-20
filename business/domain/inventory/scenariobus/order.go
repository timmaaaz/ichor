package scenariobus

import "github.com/timmaaaz/ichor/business/sdk/order"

const (
	OrderByID          = "scenario_id"
	OrderByName        = "name"
	OrderByCreatedDate = "created_date"
)

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

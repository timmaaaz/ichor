package officebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderByID       = "office_id"
	OrderByName     = "name"
	OrderByStreetID = "street_id"
)

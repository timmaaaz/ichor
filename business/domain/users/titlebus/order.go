package titlebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderByID          = "id"
	OrderByName        = "name"
	OrderByDescription = "description"
)

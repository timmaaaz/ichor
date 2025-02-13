package tagbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderByID          = "tag_id"
	OrderByName        = "name"
	OrderByDescription = "description"
)

package fulfillmentstatusbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderByID     = "id"
	OrderByIconID = "icon_id"
	OrderByName   = "name"
)

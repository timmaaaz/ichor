package fulfillmentstatusbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderByID     = "fulfillment_status_id"
	OrderByIconID = "icon_id"
	OrderByName   = "name"
)

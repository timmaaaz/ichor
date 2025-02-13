package fulfillmentstatusapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"fulfillment_status_id": fulfillmentstatusbus.OrderByID,
	"icon_id":               fulfillmentstatusbus.OrderByIconID,
	"name":                  fulfillmentstatusbus.OrderByName,
}

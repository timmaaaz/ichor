package lineitemfulfillmentstatusbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultLineItemBy = order.NewBy(LineItemByName, order.ASC)

const (
	LineItemByID          = "id"
	LineItemByName        = "name"
	LineItemByDescription = "description"
)

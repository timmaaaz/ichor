package lotlocationapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":           lotlocationbus.OrderByID,
	"lot_id":       lotlocationbus.OrderByLotID,
	"location_id":  lotlocationbus.OrderByLocationID,
	"quantity":     lotlocationbus.OrderByQuantity,
	"created_date": lotlocationbus.OrderByCreatedDate,
	"updated_date": lotlocationbus.OrderByUpdatedDate,
}

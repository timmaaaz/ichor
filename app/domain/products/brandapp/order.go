package brandapp

import (
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":           brandbus.OrderByID,
	"name":         brandbus.OrderByName,
	"created_date": brandbus.OrderByCreatedDate,
	"updated_date": brandbus.OrderByUpdatedDate,
}

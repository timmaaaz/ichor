package productuomapp

import (
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(productuombus.OrderByName, order.ASC)

var orderByFields = map[string]string{
	"id":           productuombus.OrderByID,
	"product_id":   productuombus.OrderByProductID,
	"name":         productuombus.OrderByName,
	"created_date": productuombus.OrderByCreatedDate,
}

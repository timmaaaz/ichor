package productcategoryapp

import (
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":           productcategorybus.OrderByID,
	"name":         productcategorybus.OrderByName,
	"description":  productcategorybus.OrderByDescription,
	"created_date": productcategorybus.OrderByCreatedDate,
	"updated_date": productcategorybus.OrderByUpdatedDate,
}

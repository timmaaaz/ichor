package productcategoryapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"product_category_id": productcategorybus.OrderByID,
	"name":                productcategorybus.OrderByName,
	"description":         productcategorybus.OrderByDescription,
	"created_date":        productcategorybus.OrderByCreatedDate,
	"updated_date":        productcategorybus.OrderByUpdatedDate,
}

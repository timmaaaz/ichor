package productapp

import (
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("product_id", order.ASC)

var orderByFields = map[string]string{
	"product_id": productbus.OrderByProductID,
	"name":       productbus.OrderByName,
	"cost":       productbus.OrderByCost,
	"quantity":   productbus.OrderByQuantity,
	"user_id":    productbus.OrderByUserID,
}

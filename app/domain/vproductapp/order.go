package vproductapp

import (
	"bitbucket.org/superiortechnologies/ichor/business/domain/vproductbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("product_id", order.ASC)

var orderByFields = map[string]string{
	"product_id": vproductbus.OrderByProductID,
	"user_id":    vproductbus.OrderByUserID,
	"name":       vproductbus.OrderByName,
	"cost":       vproductbus.OrderByCost,
	"quantity":   vproductbus.OrderByQuantity,
	"user_name":  vproductbus.OrderByUserName,
}

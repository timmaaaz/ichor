package homeapp

import (
	"bitbucket.org/superiortechnologies/ichor/business/domain/homebus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("home_id", order.ASC)

var orderByFields = map[string]string{
	"home_id": homebus.OrderByID,
	"type":    homebus.OrderByType,
	"user_id": homebus.OrderByUserID,
}

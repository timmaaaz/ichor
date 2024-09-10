package regionapp

import (
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/regionbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"region_id":  regionbus.OrderByID,
	"country_id": regionbus.OrderByCountryID,
	"name":       regionbus.OrderByName,
	"code":       regionbus.OrderByCode,
}

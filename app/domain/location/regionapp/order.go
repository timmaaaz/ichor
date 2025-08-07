package regionapp

import (
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":         regionbus.OrderByID,
	"country_id": regionbus.OrderByCountryID,
	"name":       regionbus.OrderByName,
	"code":       regionbus.OrderByCode,
}

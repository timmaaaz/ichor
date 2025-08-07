package cityapp

import (
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":        citybus.OrderByID,
	"region_id": citybus.OrderByRegionID,
	"name":      citybus.OrderByName,
}

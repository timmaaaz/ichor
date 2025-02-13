package officeapp

import (
	"github.com/timmaaaz/ichor/business/domain/location/officebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"office_id": officebus.OrderByID,
	"name":      officebus.OrderByName,
	"street_id": officebus.OrderByStreetID,
}

package officeapp

import (
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":        officebus.OrderByID,
	"name":      officebus.OrderByName,
	"street_id": officebus.OrderByStreetID,
}

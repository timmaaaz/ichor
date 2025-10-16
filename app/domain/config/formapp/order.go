package formapp

import (
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(formbus.OrderByName, order.ASC)

var orderByFields = map[string]string{
	formbus.OrderByID:   formbus.OrderByID,
	formbus.OrderByName: formbus.OrderByName,
}
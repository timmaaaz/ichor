package countryapp

import (
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("number", order.ASC)

var orderByFields = map[string]string{
	"country_id": countrybus.OrderByID,
	"number":     countrybus.OrderByNumber,
	"name":       countrybus.OrderByName,
	"alpha_2":    countrybus.OrderByAlpha2,
	"alpha_3":    countrybus.OrderByAlpha3,
}

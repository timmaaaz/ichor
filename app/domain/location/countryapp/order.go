package countryapp

import (
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("number", order.ASC)

var orderByFields = map[string]string{
	"country_id": countrybus.OrderByID,
	"number":     countrybus.OrderByNumber,
	"name":       countrybus.OrderByName,
	"alpha_2":    countrybus.OrderByAlpha2,
	"alpha_3":    countrybus.OrderByAlpha3,
}

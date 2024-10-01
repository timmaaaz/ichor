package streetapp

import (
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("line_1", order.ASC)

var orderByFields = map[string]string{
	"street_id":   streetbus.OrderByID,
	"city_id":     streetbus.OrderByCityID,
	"line_1":      streetbus.OrderByLine1,
	"line_2":      streetbus.OrderByLine2,
	"postal_code": streetbus.OrderByPostalCode,
}

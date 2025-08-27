package validassetapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":                   validassetbus.OrderByID,
	"type_id":              validassetbus.OrderByTypeID,
	"name":                 validassetbus.OrderByName,
	"est_price":            validassetbus.OrderByEstPrice,
	"price":                validassetbus.OrderByPrice,
	"maintenance_interval": validassetbus.OrderByMaintenance,
	"life_expectancy":      validassetbus.OrderByLifeExpectancy,
	"model_number":         validassetbus.OrderByModelNumber,
	"is_enabled":           validassetbus.OrderByIsEnabled,
	"created_date":         validassetbus.OrderByCreatedDate,
	"updated_date":         validassetbus.OrderByUpdatedDate,
	"created_by":           validassetbus.OrderByCreatedBy,
	"updated_by":           validassetbus.OrderByUpdatedBy,
}

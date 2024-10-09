package assetapp

import (
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"asset_id":             assetbus.OrderByID,
	"type_id":              assetbus.OrderByTypeID,
	"condition_id":         assetbus.OrderByConditionID,
	"name":                 assetbus.OrderByName,
	"est_price":            assetbus.OrderByEstPrice,
	"price":                assetbus.OrderByPrice,
	"maintenance_interval": assetbus.OrderByMaintenance,
	"life_expectancy":      assetbus.OrderByLifeExpectancy,
	"model_number":         assetbus.OrderByModelNumber,
	"is_enabled":           assetbus.OrderByIsEnabled,
	"date_created":         assetbus.OrderByDateCreated,
	"date_updated":         assetbus.OrderByDateUpdated,
	"created_by":           assetbus.OrderByCreatedBy,
	"updated_by":           assetbus.OrderByUpdatedBy,
}

package assetapp

import (
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("asset_id", order.ASC)

var orderByFields = map[string]string{
	"asset_id":              assetbus.OrderByID,
	"valid_asset_id":        assetbus.OrderByValidAssetID,
	"asset_condition_id":    assetbus.OrderByConditionID,
	"last_maintenance_time": assetbus.OrderByLastMaintenance,
	"serial_number":         assetbus.OrderBySerialNumber,
}

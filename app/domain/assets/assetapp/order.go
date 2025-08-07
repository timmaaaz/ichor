package assetapp

import (
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(assetbus.OrderByID, order.ASC)

var orderByFields = map[string]string{
	"id":                    assetbus.OrderByID,
	"valid_asset_id":        assetbus.OrderByValidAssetID,
	"asset_condition_id":    assetbus.OrderByConditionID,
	"last_maintenance_time": assetbus.OrderByLastMaintenance,
	"serial_number":         assetbus.OrderBySerialNumber,
}

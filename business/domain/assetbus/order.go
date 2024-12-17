package assetbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID              = "asset_id"
	OrderByConditionID     = "asset_condition_id"
	OrderByLastMaintenance = "last_maintenance_time"
	OrderByValidAssetID    = "valid_asset_id"
	OrderBySerialNumber    = "serial_number"
)

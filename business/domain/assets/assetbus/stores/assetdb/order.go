package assetdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	assetbus.OrderByID:              "id",
	assetbus.OrderByConditionID:     "asset_condition_id",
	assetbus.OrderByLastMaintenance: "last_maintenance_time",
	assetbus.OrderBySerialNumber:    "serial_number",
	assetbus.OrderByValidAssetID:    "valid_asset_id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

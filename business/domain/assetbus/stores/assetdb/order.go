package assetdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	assetbus.OrderByID:             "asset_id",
	assetbus.OrderByTypeID:         "type_id",
	assetbus.OrderByName:           "name",
	assetbus.OrderByEstPrice:       "est_price",
	assetbus.OrderByPrice:          "price",
	assetbus.OrderByMaintenance:    "maintenance_interval",
	assetbus.OrderByLifeExpectancy: "life_expectancy",
	assetbus.OrderBySerialNumber:   "serial_number",
	assetbus.OrderByModelNumber:    "model_number",
	assetbus.OrderByIsEnabled:      "is_enabled",
	assetbus.OrderByDateCreated:    "date_created",
	assetbus.OrderByDateUpdated:    "date_updated",
	assetbus.OrderByCreatedBy:      "created_by",
	assetbus.OrderByUpdatedBy:      "updated_by",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

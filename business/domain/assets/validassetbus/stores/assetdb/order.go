package validassetdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	validassetbus.OrderByID:             "id",
	validassetbus.OrderByTypeID:         "type_id",
	validassetbus.OrderByName:           "name",
	validassetbus.OrderByEstPrice:       "est_price",
	validassetbus.OrderByPrice:          "price",
	validassetbus.OrderByMaintenance:    "maintenance_interval",
	validassetbus.OrderByLifeExpectancy: "life_expectancy",
	validassetbus.OrderBySerialNumber:   "serial_number",
	validassetbus.OrderByModelNumber:    "model_number",
	validassetbus.OrderByIsEnabled:      "is_enabled",
	validassetbus.OrderByDateCreated:    "date_created",
	validassetbus.OrderByDateUpdated:    "date_updated",
	validassetbus.OrderByCreatedBy:      "created_by",
	validassetbus.OrderByUpdatedBy:      "updated_by",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

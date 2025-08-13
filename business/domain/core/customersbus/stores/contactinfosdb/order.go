package customersdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/core/customersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	customersbus.OrderByID:                "id",
	customersbus.OrderByName:              "name",
	customersbus.OrderByContactID:         "contact_id",
	customersbus.OrderByDeliveryAddressID: "delivery_address_id",
	customersbus.OrderByNotes:             "notes",
	customersbus.OrderByCreatedBy:         "created_by",
	customersbus.OrderByUpdatedBy:         "updated_by",
	customersbus.OrderByCreatedDate:       "created_date",
	customersbus.OrderByUpdatedDate:       "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

package timezonedb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	timezonebus.OrderByID:          "id",
	timezonebus.OrderByName:        "name",
	timezonebus.OrderByDisplayName: "display_name",
	timezonebus.OrderByUTCOffset:   "utc_offset",
	timezonebus.OrderByIsActive:    "is_active",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

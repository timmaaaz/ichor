package organizationalunitdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	organizationalunitbus.OrderByID:                    "organizational_unit_id",
	organizationalunitbus.OrderByParentID:              "parent_id",
	organizationalunitbus.OrderByName:                  "name",
	organizationalunitbus.OrderByLevel:                 "level",
	organizationalunitbus.OrderByPath:                  "path",
	organizationalunitbus.OrderByCanInheritPermissions: "can_inherit_permissions",
	organizationalunitbus.OrderByCanRollupData:         "can_rollup_data",
	organizationalunitbus.OrderByUnitType:              "unit_type",
	organizationalunitbus.OrderByIsActive:              "is_active",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

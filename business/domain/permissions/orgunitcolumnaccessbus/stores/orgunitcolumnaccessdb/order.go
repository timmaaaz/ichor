package orgunitcolumnaccessdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/permissions/orgunitcolumnaccessbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	orgunitcolumnaccessbus.OrderByID:                    "org_unit_column_access_id",
	orgunitcolumnaccessbus.OrderByOrganizationalUnitID:  "organizational_unit_id",
	orgunitcolumnaccessbus.OrderByTableName:             "table_name",
	orgunitcolumnaccessbus.OrderByColumnName:            "column_name",
	orgunitcolumnaccessbus.OrderByCanRead:               "can_read",
	orgunitcolumnaccessbus.OrderByCanUpdate:             "can_update",
	orgunitcolumnaccessbus.OrderByCanInheritPermissions: "can_inherit_permissions",
	orgunitcolumnaccessbus.OrderByCanRollupData:         "can_rollup_data",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}
	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

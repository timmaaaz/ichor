package orgunitcolumnaccessbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByOrganizationalUnitID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID                    = "org_unit_column_access_id"
	OrderByOrganizationalUnitID  = "organizational_unit_id"
	OrderByTableName             = "table_name"
	OrderByColumnName            = "column_name"
	OrderByCanRead               = "can_read"
	OrderByCanUpdate             = "can_update"
	OrderByCanInheritPermissions = "can_inherit_permissions"
	OrderByCanRollupData         = "can_rollup_data"
)

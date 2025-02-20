package organizationalunitbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID                    = "organizational_unit_id"
	OrderByParentID              = "parent_id"
	OrderByName                  = "name"
	OrderByLevel                 = "level"
	OrderByPath                  = "path"
	OrderByCanInheritPermissions = "can_inherit_permissions"
	OrderByCanRollupData         = "can_rollup_data"
	OrderByUnitType              = "unit_type"
	OrderByIsActive              = "is_active"
)
